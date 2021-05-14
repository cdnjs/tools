package algolia_pump

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"

	"github.com/cdnjs/tools/algolia"
	"github.com/cdnjs/tools/gcp"
	"github.com/cdnjs/tools/kv"
	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/sentry"
)

var (
	ENV           = os.Getenv("ENV")
	KV_TOKEN      = os.Getenv("KV_TOKEN")
	CF_ACCOUNT_ID = os.Getenv("CF_ACCOUNT_ID")
)

func getExistingVersions(p *packages.Package) ([]string, error) {
	cfapi, err := cloudflare.NewWithAPIToken(KV_TOKEN, cloudflare.UsingAccount(CF_ACCOUNT_ID))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloudflare API client")
	}

	versions, err := kv.GetVersions(cfapi, *p.Name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get verions")
	}

	return versions, nil
}

func Invoke(ctx context.Context, e gcp.GCSEvent) error {
	sentry.Init()
	defer sentry.PanicHandler()

	log.Printf("File: %v\n", e.Name)
	log.Printf("Metadata: %v\n", e.Metadata)

	pkgName := e.Metadata["package"].(string)
	currVersion := e.Metadata["version"].(string)

	configStr, err := b64.StdEncoding.DecodeString(e.Metadata["config"].(string))
	if err != nil {
		return fmt.Errorf("could not decode config: %v", err)
	}

	pkg := new(packages.Package)
	if err := json.Unmarshal([]byte(configStr), &pkg); err != nil {
		return fmt.Errorf("could not decode config: %v", err)
	}
	// update package version with latest
	versions, err := getExistingVersions(pkg)
	if err != nil {
		return fmt.Errorf("failed to retrieve existing versions: %s", err)
	}
	// add the current version in case it was yet present in KV
	versions = append(versions, currVersion)

	pkg.Version = packages.GetLatestStableVersion(versions)

	archive, err := gcp.ReadObject(ctx, e.Bucket, e.Name)
	if err != nil {
		return fmt.Errorf("could not read object: %v", err)
	}

	sris := make(map[string]string)
	onFile := func(name string, r io.Reader) error {
		ext := filepath.Ext(name)
		// remove leading slash
		name = name[1:]

		content, err := ioutil.ReadAll(r)
		if err != nil {
			return errors.Wrap(err, "could not read file")
		}

		if ext == ".sri" {
			filename := name[0 : len(name)-len(ext)]
			sris[filename] = string(content)
			return nil
		}
		return nil
	}
	log.Printf("SRIs: %s\n", sris)
	if err := gcp.Inflate(bytes.NewReader(archive), onFile); err != nil {
		return fmt.Errorf("could not inflate archive: %s", err)
	}

	log.Printf("updating %s in search index (last version %s)\n", pkgName, printStrPtr(pkg.Version))
	if ENV != "prod" {
		log.Printf("Algolia doesn't update in %s\n", ENV)
		return nil
	}

	index := algolia.GetProdIndex(algolia.GetClient())

	if err := algolia.IndexPackage(pkg, index, sris); err != nil {
		return fmt.Errorf("failed to update algolia index: %v", err)
	}
	return nil
}

func printStrPtr(v *string) string {
	if v == nil {
		return "<nil>"
	} else {
		return *v
	}
}
