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
	"github.com/cdnjs/tools/audit"
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

func getExistingVersionsFromAggregatedMetadata(p *packages.Package) ([]string, error) {
	cfapi, err := cloudflare.NewWithAPIToken(KV_TOKEN, cloudflare.UsingAccount(CF_ACCOUNT_ID))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloudflare API client")
	}

	log.Printf("Fetching versions from aggregated metadata for: `%s`\n", *p.Name)
	versions, err := kv.GetVersionsFromAggregatedMetadata(cfapi, *p.Name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get versions")
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
	versions, err := getExistingVersionsFromAggregatedMetadata(pkg)
	if err != nil {
		return fmt.Errorf("failed to retrieve existing versions: %s", err)
	}
	archive, err := gcp.ReadObject(ctx, e.Bucket, e.Name)
	if err != nil {
		return fmt.Errorf("could not read object: %v", err)
	}

	sris := make(map[string]string)
	files := make([]string, 0)
	onFile := func(name string, r io.Reader) error {
		ext := filepath.Ext(name)
		// remove leading slash
		name = name[1:]
		filename := name[0 : len(name)-len(ext)]

		if ext == ".sri" {
			content, err := ioutil.ReadAll(r)
			if err != nil {
				return errors.Wrap(err, "could not read file")
			}
			sris[filename] = string(content)
		}

		if ext == ".gz" || ext == ".woff2" {
			files = append(files, filename)
		}
		return nil
	}
	if err := gcp.Inflate(bytes.NewReader(archive), onFile); err != nil {
		return fmt.Errorf("could not inflate archive: %s", err)
	}

	log.Printf("%s: %d files, SRIs: %s\n", pkgName, len(files), sris)

	if len(files) > 0 {
		// add the current version in case it was yet present in KV
		versions = append(versions, currVersion)
	}

	// Update package's current version and fix filename if needed
	pkg.Version = packages.GetLatestStableVersion(versions)
	if err := packages.UpdateFilenameIfMissing(ctx, pkg, files); err != nil {
		return errors.Wrap(err, "failed to fix missing filename")
	}

	log.Printf("%s: updating %s in search index (last version %s)\n", pkgName, pkgName, printStrPtr(pkg.Version))
	if ENV != "prod" {
		log.Printf("%s: algolia doesn't update in %s\n", pkgName, ENV)
		return nil
	}

	index := algolia.GetProdIndex(algolia.GetClient())

	entry, err := algolia.IndexPackage(pkg, index, sris)
	if err != nil {
		return fmt.Errorf("failed to update algolia index: %v", err)
	}
	if err := audit.WroteAlgolia(ctx, pkgName, currVersion, pkg.Version, entry); err != nil {
		return fmt.Errorf("failed to audit: %s", err)
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
