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

	"github.com/pkg/errors"

	"github.com/cdnjs/tools/algolia"
	"github.com/cdnjs/tools/gcp"
	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/sentry"
)

var (
	ENV = os.Getenv("ENV")
)

func Invoke(ctx context.Context, e gcp.GCSEvent) error {
	if ENV != "prod" {
		log.Printf("Algolia doesn't update in %s\n", ENV)
		return nil
	}

	sentry.Init()
	defer sentry.PanicHandler()

	log.Printf("File: %v\n", e.Name)
	log.Printf("Metadata: %v\n", e.Metadata)

	pkgName := e.Metadata["package"].(string)
	version := e.Metadata["version"].(string)

	configStr, err := b64.StdEncoding.DecodeString(e.Metadata["config"].(string))
	if err != nil {
		return fmt.Errorf("could not decode config: %v", err)
	}

	pkg := new(packages.Package)
	if err := json.Unmarshal([]byte(configStr), &pkg); err != nil {
		return fmt.Errorf("could not decode config: %v", err)
	}
	// update package version with latest
	pkg.Version = &version

	archive, err := gcp.ReadObject(ctx, e.Bucket, e.Name)
	if err != nil {
		return fmt.Errorf("could not read object: %v", err)
	}

	sris := make(map[string]string)
	onFile := func(name string, r io.Reader) error {
		ext := filepath.Ext(name)
		// remove leading slash
		name = name[1:]
		key := fmt.Sprintf("%s/%s/%s", pkgName, version, name)

		content, err := ioutil.ReadAll(r)
		if err != nil {
			return errors.Wrap(err, "could not read file")
		}

		if ext == ".sri" {
			filename := key[0 : len(key)-len(ext)]
			sris[filename] = string(content)
			return nil
		}
		return nil
	}
	if err := gcp.Inflate(bytes.NewReader(archive), onFile); err != nil {
		return fmt.Errorf("could not inflate archive: %s", err)
	}

	index := algolia.GetProdIndex(algolia.GetClient())

	if err := algolia.IndexPackage(pkg, index, sris); err != nil {
		return fmt.Errorf("failed to update algolia index: %v", err)
	}
	return nil
}
