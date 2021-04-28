package kv_pump

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cdnjs/tools/kv"
	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/sentry"
	"github.com/cdnjs/tools/sri"

	"cloud.google.com/go/storage"
	"github.com/agnivade/levenshtein"
	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
)

var (
	KV_TOKEN         = os.Getenv("KV_TOKEN")
	CF_ACCOUNT_ID    = os.Getenv("CF_ACCOUNT_ID")
	KV_NAMESPACE     = os.Getenv("KV_NAMESPACE")
	SRI_KV_NAMESPACE = os.Getenv("WORKERS_KV_SRIS_NAMESPACE_ID")
)

// GCSEvent is the payload of a GCS event.
type GCSEvent struct {
	Kind                    string                 `json:"kind"`
	ID                      string                 `json:"id"`
	SelfLink                string                 `json:"selfLink"`
	Name                    string                 `json:"name"`
	Bucket                  string                 `json:"bucket"`
	Generation              string                 `json:"generation"`
	Metageneration          string                 `json:"metageneration"`
	ContentType             string                 `json:"contentType"`
	TimeCreated             time.Time              `json:"timeCreated"`
	Updated                 time.Time              `json:"updated"`
	TemporaryHold           bool                   `json:"temporaryHold"`
	EventBasedHold          bool                   `json:"eventBasedHold"`
	RetentionExpirationTime time.Time              `json:"retentionExpirationTime"`
	StorageClass            string                 `json:"storageClass"`
	TimeStorageClassUpdated time.Time              `json:"timeStorageClassUpdated"`
	Size                    string                 `json:"size"`
	MD5Hash                 string                 `json:"md5Hash"`
	MediaLink               string                 `json:"mediaLink"`
	ContentEncoding         string                 `json:"contentEncoding"`
	ContentDisposition      string                 `json:"contentDisposition"`
	CacheControl            string                 `json:"cacheControl"`
	Metadata                map[string]interface{} `json:"metadata"`
	CRC32C                  string                 `json:"crc32c"`
	ComponentCount          int                    `json:"componentCount"`
	Etag                    string                 `json:"etag"`
	CustomerEncryption      struct {
		EncryptionAlgorithm string `json:"encryptionAlgorithm"`
		KeySha256           string `json:"keySha256"`
	}
	KMSKeyName    string `json:"kmsKeyName"`
	ResourceState string `json:"resourceState"`
}

func Invoke(ctx context.Context, e GCSEvent) error {
	sentry.Init()
	defer sentry.PanicHandler()

	log.Printf("File: %v\n", e.Name)
	log.Printf("Metadata: %v\n", e.Metadata)

	pkgName := e.Metadata["package"].(string)
	version := e.Metadata["version"].(string)

	archive, err := readObject(ctx, e.Bucket, e.Name)
	if err != nil {
		return fmt.Errorf("could not read object: %v", err)
	}

	cfapi, err := cloudflare.NewWithAPIToken(KV_TOKEN, cloudflare.UsingAccount(CF_ACCOUNT_ID))
	if err != nil {
		return errors.Wrap(err, "failed to create cloudflare API client")
	}

	var pairs []*kv.WriteRequest
	newKVFiles := make([]string, 0)
	sris := make(map[string]string)

	onFile := func(name string, r io.Reader) error {
		newKVFiles = append(newKVFiles, name)

		// remove leading slash
		key := fmt.Sprintf("%s/%s/%s", pkgName, version, name[1:])
		bytes, err := ioutil.ReadAll(r)
		if err != nil {
			return errors.Wrap(err, "could not read file")
		}

		ext := filepath.Ext(key)
		filename := key[0 : len(key)-len(ext)]

		if _, ok := sris[filename]; !ok {
			sris[filename] = sri.CalculateSRI(bytes)
		}

		meta := newMetadata(len(bytes))
		writePair := &kv.WriteRequest{
			Key:   key,
			Name:  key,
			Value: bytes,
			Meta:  meta,
		}
		metaStr, _ := json.Marshal(meta)
		log.Printf("%s (%s)\n", key, metaStr)
		pairs = append(pairs, writePair)
		return nil
	}
	if err := inflate(bytes.NewReader(archive), onFile); err != nil {
		return fmt.Errorf("could not inflate archive: %s", err)
	}

	res, err := kv.EncodeAndWriteKVBulk(ctx, cfapi, pairs, KV_NAMESPACE, false)
	if err != nil {
		return fmt.Errorf("failed to write KV: %s", err)
	}
	log.Println("files", res)

	newFiles := cleanNewKVFiles(newKVFiles)

	if err := updateVersions(ctx, cfapi, pkgName, version, newFiles); err != nil {
		log.Fatalf("failed to update versions: %s", err)
	}

	if err := updateAggregatedMetadata(ctx, cfapi, pkgName, version, newFiles); err != nil {
		log.Fatalf("failed to update aggregated metadata: %s", err)
	}

	if err := updatePackage(ctx, cfapi, pkgName, version, newFiles); err != nil {
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
		// remove leading slash
		out = append(out, name[1:])
	}

	return out
}

func updateVersions(ctx context.Context, cfapi *cloudflare.API, pkgName string,
	version string, files []string) error {
	_, err := kv.UpdateKVVersion(ctx, cfapi, pkgName, version, files)
	if err != nil {
		return errors.Wrap(err, "failed to update version in KV")
	}
	log.Println("add version in KV")
	return nil
}

func updatePackage(ctx context.Context, cfapi *cloudflare.API, pkgName string,
	version string, files []string) error {
	pkg, err := kv.GetPackage(ctx, cfapi, pkgName)
	if err != nil && strings.Contains(err.Error(), "key not found") {
		// Package is not in KV yet, retrieve original config
		pkg, err = packages.GetRepoPackage(pkgName)
		if err != nil {
			return errors.Wrap(err, "could not retrieve package")
		}
	} else if err != nil {
		return errors.Wrap(err, "could not retrieve package")
	}

	if pkg == nil {
		return errors.Wrap(err, "could not get package")
	}

	pkg.Version = &version
	log.Println(pkg)

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
	pkgName string, version string, newFiles []string) error {
	// Update aggregated package metadata for cdnjs API.
	newAssets := packages.Asset{
		Version: version,
		Files:   newFiles,
	}
	kvWrites, _, err := kv.UpdateAggregatedMetadata(cfapi, ctx, pkgName, version, newAssets)
	if err != nil {
		return (errors.Errorf("(%s) failed to update aggregated metadata: %s", pkgName, err))
	}
	if len(kvWrites) == 0 {
		return (errors.Errorf("(%s) failed to update aggregated metadata (no KV writes!)", pkgName))
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

func readObject(ctx context.Context, bucket string, name string) ([]byte, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "could not create client")
	}

	bkt := client.Bucket(bucket)
	obj := bkt.Object(name)

	r, err := obj.NewReader(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "could not create client")
	}
	defer r.Close()

	var buff bytes.Buffer
	w := bufio.NewWriter(&buff)

	if _, err := io.Copy(w, r); err != nil {
		return nil, errors.Wrap(err, "could not read object")
	}

	return buff.Bytes(), nil
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
