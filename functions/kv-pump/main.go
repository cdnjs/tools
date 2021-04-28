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
	"net/http"
	"os"
	"time"

	"github.com/cdnjs/tools/kv"
	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/sentry"

	"cloud.google.com/go/storage"
	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
)

var (
	KV_TOKEN      = os.Getenv("KV_TOKEN")
	CF_ACCOUNT_ID = os.Getenv("CF_ACCOUNT_ID")
	KV_NAMESPACE  = os.Getenv("KV_NAMESPACE")
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
	newFiles := make([]string, 0)

	onFile := func(name string, r io.Reader) error {
		newFiles = append(newFiles, name)

		// remove leading slash
		key := fmt.Sprintf("%s/%s/%s", pkgName, version, name[1:])
		bytes, err := ioutil.ReadAll(r)
		if err != nil {
			return errors.Wrap(err, "could not read file")
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

	// Update aggregated package metadata for cdnjs API.
	newAssets := packages.Asset{
		Version: version,
		Files:   newFiles,
	}
	kvWrites, _, err := kv.UpdateAggregatedMetadata(cfapi, ctx, pkgName, newAssets)
	if err != nil {
		panic(fmt.Sprintf("(%s) failed to update aggregated metadata: %s", pkgName, err))
	}
	if len(kvWrites) == 0 {
		panic(fmt.Sprintf("(%s) failed to update aggregated metadata (no KV writes!)", pkgName))
	}
	log.Println("aggregated", kvWrites)

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
