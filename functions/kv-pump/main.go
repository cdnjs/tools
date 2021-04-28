package kv_pump

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"cloud.google.com/go/functions/metadata"
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
	meta, err := metadata.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("metadata.FromContext: %v", err)
	}
	log.Printf("Event ID: %v\n", meta.EventID)
	log.Printf("Event type: %v\n", meta.EventType)
	log.Printf("File: %v\n", e.Name)
	log.Printf("Metadata: %v\n", e.Metadata)

	// pkg := e.Metadata["package"].(string)
	// version := e.Metadata["version"].(string)
	// pkg := "hi-sven"
	// version := "1.10.0"

	archive, err := readObject(ctx, e.Bucket, e.Name)
	if err != nil {
		return fmt.Errorf("could not read object: %v", err)
	}

	cfapi, err := cloudflare.NewWithAPIToken(KV_TOKEN, cloudflare.UsingAccount(CF_ACCOUNT_ID))
	if err != nil {
		return errors.Wrap(err, "failed to create cloudflare API client")
	}

	onFile := func(name string, r io.Reader) error {
		log.Printf("name: %s\n", name)
		// remove leading slash
		key := name[1:]
		bytes, err := ioutil.ReadAll(r)
		if err != nil {
			return errors.Wrap(err, "could not read file")
		}

		resp, err := cfapi.WriteWorkersKV(context.Background(), KV_NAMESPACE, key, bytes)
		if err != nil {
			return errors.Wrap(err, "failed to write in KV")
		}

		fmt.Println(resp)

		return nil
	}
	if err := inflate(bytes.NewReader(archive), onFile); err != nil {
		return fmt.Errorf("could not inflate archive: %s", err)
	}

	return nil
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
