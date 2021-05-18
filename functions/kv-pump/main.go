package kv_pump

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cdnjs/tools/audit"
	"github.com/cdnjs/tools/gcp"
	"github.com/cdnjs/tools/kv"
	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/sentry"

	"github.com/agnivade/levenshtein"
	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
)

var (
	KV_TOKEN              = os.Getenv("KV_TOKEN")
	CF_ACCOUNT_ID         = os.Getenv("CF_ACCOUNT_ID")
	FILES_KV_NAMESPACE_ID = os.Getenv("FILES_KV_NAMESPACE_ID")
	SRI_KV_NAMESPACE      = os.Getenv("WORKERS_KV_SRIS_NAMESPACE_ID")
)

func Invoke(ctx context.Context, e gcp.GCSEvent) error {
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

	archive, err := gcp.ReadObject(ctx, e.Bucket, e.Name)
	if err != nil {
		return fmt.Errorf("could not read object: %v", err)
	}

	cfapi, err := cloudflare.NewWithAPIToken(KV_TOKEN, cloudflare.UsingAccount(CF_ACCOUNT_ID))
	if err != nil {
		return errors.Wrap(err, "failed to create cloudflare API client")
	}

	var pairs []*kv.WriteRequest
	kvKeys := make([]string, 0)
	sris := make(map[string]string)
	kvfiles := make([]string, 0)

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

		if ext == ".gz" || ext == ".br" || ext == ".woff2" {
			kvKeys = append(kvKeys, key)
			kvfiles = append(kvfiles, name)

			meta := newMetadata(len(content))
			writePair := &kv.WriteRequest{
				Key:   key,
				Name:  key,
				Value: content,
				Meta:  meta,
			}
			metaStr, _ := json.Marshal(meta)
			log.Printf("%s (%s)\n", key, metaStr)
			pairs = append(pairs, writePair)
		}
		return nil
	}
	if err := gcp.Inflate(bytes.NewReader(archive), onFile); err != nil {
		return fmt.Errorf("could not inflate archive: %s", err)
	}

	if len(pairs) > 0 {
		res, err := kv.EncodeAndWriteKVBulk(ctx, cfapi, pairs, FILES_KV_NAMESPACE_ID, false)
		if err != nil {
			return fmt.Errorf("failed to write KV: %s", err)
		}
		log.Println("files", res)
		if err := audit.WroteKV(ctx, pkgName, version, sris, kvKeys, string(configStr)); err != nil {
			log.Printf("failed to audit: %s\n", err)
		}

		newFiles := cleanNewKVFiles(kvfiles)

		pkg := new(packages.Package)
		if err := json.Unmarshal([]byte(configStr), &pkg); err != nil {
			return fmt.Errorf("failed to parse config: %s", err)
		}

		if err := updateVersions(ctx, cfapi, pkg, version, newFiles); err != nil {
			return fmt.Errorf("failed to update versions: %s", err)
		}

		if err := updateAggregatedMetadata(ctx, cfapi, pkg, version, newFiles); err != nil {
			return fmt.Errorf("failed to update aggregated metadata: %s", err)
		}

		if err := updatePackage(ctx, cfapi, pkg, version, newFiles); err != nil {
			return fmt.Errorf("failed to update package: %s", err)
		}

		if err := updateSRIs(ctx, cfapi, sris); err != nil {
			return fmt.Errorf("failed to update SRIs: %s", err)
		}
	} else {
		log.Println("no files to publish")
		return nil
	}

	return nil
}

// KV has optimized files (ending in .gz/br), if we want the original files we
// need to dedup them and remove their compression ext
func cleanNewKVFiles(files []string) []string {
	seen := make(map[string]bool)
	out := make([]string, 0)
	for _, file := range files {
		ext := filepath.Ext(file)
		name := file[0 : len(file)-len(ext)]

		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = true
		out = append(out, name)
	}

	return out
}

func getExistingVersions(cfapi *cloudflare.API, p *packages.Package) ([]string, error) {
	versions, err := kv.GetVersions(cfapi, *p.Name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get verions")
	}

	return versions, nil
}

func updateVersions(ctx context.Context, cfapi *cloudflare.API, pkg *packages.Package,
	version string, files []string) error {
	_, err := kv.UpdateKVVersion(ctx, cfapi, *pkg.Name, version, files)
	if err != nil {
		return errors.Wrap(err, "failed to update version in KV")
	}
	log.Println("add version in KV")
	return nil
}

func updatePackage(ctx context.Context, cfapi *cloudflare.API, pkg *packages.Package,
	currVersion string, files []string) error {
	// update package version with latest
	versions, err := getExistingVersions(cfapi, pkg)
	if err != nil {
		return fmt.Errorf("failed to retrieve existing versions: %s", err)
	}
	// add the current version in case it was yet present in KV
	versions = append(versions, currVersion)

	pkg.Version = packages.GetLatestStableVersion(versions)
	log.Println("updated package", pkg)

	if err := updateFilenameIfMissing(ctx, cfapi, pkg, files); err != nil {
		return errors.Wrap(err, "failed to fix missing filename")
	}

	// sync with KV first, then update legacy package.json
	if err := kv.UpdateKVPackage(ctx, cfapi, pkg); err != nil {
		return errors.Wrap(err, "failed to write KV package metadata")
	}
	log.Println("updated package")

	return nil
}

// Update the package's filename if the latest
// version does not contain the filename
// Note that if the filename is nil it will stay nil.
func updateFilenameIfMissing(ctx context.Context, cfapi *cloudflare.API, pkg *packages.Package, files []string) error {
	key := pkg.LatestVersionKVKey()

	if len(files) == 0 {
		return errors.Errorf("KV version `%s` contains no files", key)
	}

	if pkg.Filename != nil {
		// check if assets contains filename
		filename := *pkg.Filename
		for _, asset := range files {
			if asset == filename {
				return nil // filename included in latest version, so return
			}
		}

		// set filename to be the most similar string in []assets
		mostSimilar := getMostSimilarFilename(filename, files)
		pkg.Filename = &mostSimilar
		log.Printf("Updated `%s` filename `%s` -> `%s`\n", key, filename, mostSimilar)
		return nil
	}
	log.Printf("Filename in `%s` missing, so will stay missing.\n", key)
	return nil
}

// Gets the most similar filename to a target filename.
// The []string of alternatives must have at least one element.
func getMostSimilarFilename(target string, filenames []string) string {
	var mostSimilar string
	var minDist int = math.MaxInt32
	for _, f := range filenames {
		if dist := levenshtein.ComputeDistance(target, f); dist < minDist {
			mostSimilar = f
			minDist = dist
		}
	}
	return mostSimilar
}

func updateAggregatedMetadata(ctx context.Context, cfapi *cloudflare.API,
	pkg *packages.Package, version string, newFiles []string) error {
	// Update aggregated package metadata for cdnjs API.
	newAssets := packages.Asset{
		Version: version,
		Files:   newFiles,
	}
	kvWrites, _, err := kv.UpdateAggregatedMetadata(cfapi, ctx, pkg, version, newAssets)
	if err != nil {
		return errors.Errorf("(%s) failed to update aggregated metadata: %s", *pkg.Name, err)
	}
	if len(kvWrites) == 0 {
		return errors.Errorf("(%s) failed to update aggregated metadata (no KV writes!)", *pkg.Name)
	}
	log.Println("updated aggregated", kvWrites)
	return nil
}

func updateSRIs(ctx context.Context, cfapi *cloudflare.API, sris map[string]string) error {
	pairs := make([]*kv.WriteRequest, 0)

	for name, sri := range sris {
		pairs = append(pairs, &kv.WriteRequest{
			Key:  name,
			Name: name,
			Meta: &kv.FileMetadata{
				SRI: sri,
			},
		})
	}

	_, err := kv.EncodeAndWriteKVBulk(ctx, cfapi, pairs, SRI_KV_NAMESPACE, false)
	if err != nil {
		return errors.Wrap(err, "could not write bulk KV")
	}
	return nil
}

func newMetadata(size int) *kv.FileMetadata {
	lastModifiedTime := time.Now()
	lastModifiedSeconds := lastModifiedTime.UnixNano() / int64(time.Second)
	lastModifiedStr := lastModifiedTime.Format(http.TimeFormat)
	etag := fmt.Sprintf("%x-%x", lastModifiedSeconds, size)

	return &kv.FileMetadata{
		ETag:         etag,
		LastModified: lastModifiedStr,
	}
}
