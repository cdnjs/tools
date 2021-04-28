package process_version

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/sentry"

	"cloud.google.com/go/iam/credentials/apiv1"
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
	credentialspb "google.golang.org/genproto/googleapis/iam/credentials/v1"
)

var (
	TOPIC            = os.Getenv("PROCESSING_QUEUE")
	PROJECT          = os.Getenv("PROJECT")
	OUTGOING_BUCKET  = os.Getenv("OUTGOING_BUCKET")
	GOOGLE_ACCESS_ID = os.Getenv("GOOGLE_ACCESS_ID")
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

	pkg := e.Metadata["package"].(string)
	version := e.Metadata["version"].(string)
	config := e.Metadata["config"].(string)

	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", e.Bucket, e.Name)

	if err := publish(url, pkg, version, config); err != nil {
		return fmt.Errorf("failed to publish: %v", err)
	}
	return nil
}

// FIXME: share with process-version
type Message struct {
	OutgoingSignedURL string           `json:"outgoingSignedURL"`
	Tar               string           `json:"tar"`
	Pkg               string           `json:"package"`
	Version           string           `json:"version"`
	Config            packages.Package `json:"config"`
}

func publish(tar, pkg, version, configStr string) error {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, PROJECT)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient: %v", err)
	}
	t := client.Topic(TOPIC)

	signedURL, err := generateV4SignedURL(ctx, pkg, version,
		fmt.Sprintf("%s/%s/files.tgz", pkg, version)), configStr
	if err != nil {
		return errors.Wrap(err, "could not generate signed URL")
	}

	var config packages.Package
	if err := json.Unmarshal([]byte(configStr), &config); err != nil {
		return errors.Wrap(err, "could not unmarshal filemap")
	}

	msg := Message{
		OutgoingSignedURL: signedURL,
		Tar:               tar,
		Pkg:               pkg,
		Version:           version,
		Config:            config,
	}
	bytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("could not marshal message: %s", err)
	}
	result := t.Publish(ctx, &pubsub.Message{Data: bytes})

	// The Get method blocks until a server-generated ID or
	// an error is returned for the published message.
	id, err := result.Get(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to publish")
	}
	log.Printf("Published msg ID: %v\n", id)
	return nil
}

func generateV4SignedURL(ctx context.Context, pkg string, version string, config string, dst string) (string, error) {
	c, err := credentials.NewIamCredentialsClient(ctx)
	if err != nil {
		return "", errors.Wrap(err, "could not create IAM client")
	}

	opts := &storage.SignedURLOptions{
		Headers: []string{
			"x-goog-meta-package:" + pkg,
			"x-goog-meta-version:" + version,
			"x-goog-meta-config:" + config,
		},
		Scheme:         storage.SigningSchemeV4,
		Method:         "PUT",
		GoogleAccessID: GOOGLE_ACCESS_ID,
		Expires:        time.Now().Add(7*24*time.Hour - 1), // 7 days (-1h) is the max
		SignBytes: func(b []byte) ([]byte, error) {
			req := &credentialspb.SignBlobRequest{
				Payload: b,
				Name:    GOOGLE_ACCESS_ID,
			}
			resp, err := c.SignBlob(ctx, req)
			if err != nil {
				return nil, errors.Wrap(err, "could not sign blob")
			}
			return resp.SignedBlob, err
		},
	}
	url, err := storage.SignedURL(OUTGOING_BUCKET, dst, opts)
	if err != nil {
		return "", errors.Wrap(err, "failed to sign URL")
	}
	return url, nil
}
