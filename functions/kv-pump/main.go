package kv_pump

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
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
	if err := inflate(bytes.NewReader(archive), onFile); err != nil {
		return fmt.Errorf("could not inflate archive: %s", err)
	}

	res, err := kv.EncodeAndWriteKVBulk(ctx, cfapi, pairs, FILES_KV_NAMESPACE_ID, false)
	if err != nil {
		return fmt.Errorf("failed to write KV: %s", err)
	}
	log.Println("files", res)
	if err := audit.WroteKV(ctx, pkgName, version, sris, kvKeys, string(configStr)); err != nil {
		return fmt.Errorf("failed to audit: %s", err)
	}

	newFiles := cleanNewKVFiles(kvfiles)

	pkg := new(packages.Package)
	if err := json.Unmarshal([]byte(configStr), &pkg); err != nil {
		log.Fatalf("failed to parse config: %s", err)
	}

	if err := updateVersions(ctx, cfapi, pkg, version, newFiles); err != nil {
		log.Fatalf("failed to update versions: %s", err)
	}

	if err := updateAggregatedMetadata(ctx, cfapi, pkg, version, newFiles); err != nil {
		log.Fatalf("failed to update aggregated metadata: %s", err)
	}

	if err := updatePackage(ctx, cfapi, pkg, version, newFiles); err != nil {
		log.Fatalf("failed to update package: %s", err)
	}

	if err := updateSRIs(ctx, cfapi, sris); err != nil {
		log.Fatalf("failed to update SRIs: %s", err)
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
	version string, files []string) error {
	pkg.Version = &version
	log.Println("updated package", pkg)

	// latestVersion := getLatestStableVersion(allVersions)

	// if latestVersion == nil {
	// 	latestVersion = getLatestVersion(allVersions)
	// }

	// latestVersion must be non-nil by now
	// since we determined len(allVersions) > 0
	// pkg.Version = latestVersion
	if err := updateFilenameIfMissing(ctx, cfapi, pkg, files); err != nil {
		return errors.Wrap(err, "failed to fix missing filename")
	}

	// destpckg, err := kv.GetPackage(ctx, *pckg.Name)
	// if err != nil {
	// 	// check for errors
	// 	// Note: currently panicking on unhandled errors, including AuthError
	// 	switch e := err.(type) {
	// 	case kv.KeyNotFoundError:
	// 		{
	// 			// key not found (new package)
	// 			util.Debugf(ctx, "KV key `%s` not found, inserting package metadata...\n", *pckg.Name)
	// 		}
	// 	case packages.InvalidSchemaError:
	// 		{
	// 			// invalid schema found
	// 			// this should not occur, so log in sentry
	// 			// and rewrite the key so it follows the JSON schema
	// 			sentry.NotifyError(fmt.Errorf("schema invalid for KV package metadata `%s`: %s", *pckg.Name, e))
	// 		}
	// 	default:
	// 		{
	// 			// unhandled error occurred
	// 			panic(fmt.Sprintf("unhandled error reading KV package metadata: %s", e.Error()))
	// 		}
	// 	}
	// } else if destpckg.Version != nil && *destpckg.Version == *pckg.Version {
	// 	// latest version is already in KV, but we still
	// 	// need to check if the `filename` changed or not
	// 	if (destpckg.Filename == nil && pckg.Filename == nil) || (destpckg.Filename != nil && pckg.Filename != nil && *destpckg.Filename == *pckg.Filename) {
	// 		return false
	// 	}
	// }

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

func inflate(gzipStream io.Reader, onFile func(string, io.Reader) error) error {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		log.Fatal("ExtractTarGz: NewReader failed")
	}

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("ExtractTarGz: Next() failed: %s", err.Error())
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// do nothing
		case tar.TypeReg:
			if err := onFile(header.Name, tarReader); err != nil {
				return errors.Wrap(err, "failed to handle file")
			}
		default:
			return errors.Errorf(
				"ExtractTarGz: uknown type: %x in %s",
				header.Typeflag,
				header.Name)
		}
	}
	return nil
}
