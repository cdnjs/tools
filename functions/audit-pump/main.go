package audit_pump

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/cdnjs/tools/gcp"
	"github.com/cdnjs/tools/sentry"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

const (
	GH_OWNER = "cdnjs"
	GH_REPO  = "logs"
	GH_NAME  = "robocdnjs"
	GH_EMAIL = "cdnjs-github@cloudflare.com"
)

var (
	GH_TOKEN = os.Getenv("GH_TOKEN")
)

// FIXME: move to gcp package
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

	archive, err := gcp.ReadObject(ctx, e.Bucket, e.Name)
	if err != nil {
		return fmt.Errorf("could not read object: %v", err)
	}

	sris := make(map[string]string)
	kvKeys := make([]string, 0)

	onFile := func(name string, r io.Reader) error {
		ext := filepath.Ext(name)
		// remove leading slash
		key := fmt.Sprintf("%s/%s/%s", pkgName, version, name[1:])
		filename := key[0 : len(key)-len(ext)]

		if ext == ".sri" {
			content, err := ioutil.ReadAll(r)
			if err != nil {
				return errors.Wrap(err, "could not read file")
			}

			sris[filename] = string(content)
			return nil
		}

		if ext == ".gz" || ext == ".br" {
			kvKeys = append(kvKeys, key)
		}

		return nil
	}
	if err := inflate(bytes.NewReader(archive), onFile); err != nil {
		return fmt.Errorf("could not inflate archive: %s", err)
	}

	if err := createAuditFile(ctx, pkgName, version, sris, kvKeys); err != nil {
		return fmt.Errorf("could not read object: %v", err)
	}

	return nil
}

func createAuditFile(ctx context.Context, pkgName string, version string,
	sris map[string]string, keys []string) error {
	firstLetter := pkgName[0:1]
	file := fmt.Sprintf("packages/%s/%s/%s.log", firstLetter, pkgName, version)
	message := fmt.Sprintf("Push %s %s", pkgName, version)

	content := bytes.NewBufferString("")
	fmt.Fprint(content, "KV keys:\n")
	for _, key := range keys {
		fmt.Fprintf(content, "- %s\n", key)
	}
	fmt.Fprint(content, "SRIs:\n")
	for name, sri := range sris {
		fmt.Fprintf(content, "- %s: %s\n", name, sri)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GH_TOKEN},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	commitOption := &github.RepositoryContentFileOptions{
		Branch:  github.String("test2"),
		Message: github.String(message),
		Committer: &github.CommitAuthor{
			Name:  github.String(GH_NAME),
			Email: github.String(GH_EMAIL),
		},
		Author: &github.CommitAuthor{
			Name:  github.String(GH_NAME),
			Email: github.String(GH_EMAIL),
		},
		Content: content.Bytes(),
	}

	c, resp, err := client.Repositories.CreateFile(ctx, GH_OWNER, GH_REPO, file, commitOption)
	if err != nil {
		return errors.Wrap(err, "could not create file")
	}
	log.Printf("resp.Status=%v commit=%s", resp.Status, *c.SHA)
	return nil
}

// TODO: share with *-pump?
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
