package r2_pump

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cdnjs/tools/audit"
	"github.com/cdnjs/tools/gcp"
	"github.com/cdnjs/tools/metrics"
	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/sentry"
	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	R2_BUCKET     = os.Getenv("R2_BUCKET")
	R2_KEY_ID     = os.Getenv("R2_KEY_ID")
	R2_KEY_SECRET = os.Getenv("R2_KEY_SECRET")
	R2_ENDPOINT   = os.Getenv("R2_ENDPOINT")

	// In an attempt to spread the load across multiple gcp functions, we split
	// the upload per file extension (either gz, br or woff2). There should
	// be 50/50 for gz and br. Unknown for woff2.
	// Example: FILE_EXTENSION=gz
	FILE_EXTENSION = os.Getenv("FILE_EXTENSION")
)

func Invoke(ctx context.Context, e gcp.GCSEvent) error {
	sentry.Init()
	defer sentry.PanicHandler()

	pkgName := e.Metadata["package"].(string)
	version := e.Metadata["version"].(string)
	log.Printf("Invoke %s %s\n", pkgName, version)

	configStr, err := b64.StdEncoding.DecodeString(e.Metadata["config"].(string))
	if err != nil {
		return fmt.Errorf("could not decode config: %v", err)
	}

	archive, err := gcp.ReadObject(ctx, e.Bucket, e.Name)
	if err != nil {
		return fmt.Errorf("could not read object: %v", err)
	}

	bucket := aws.String(R2_BUCKET)

	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: R2_ENDPOINT,
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(R2_KEY_ID, R2_KEY_SECRET, "")),
	)
	if err != nil {
		return fmt.Errorf("could not load config: %s", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	keys := make([]string, 0)

	onFile := func(name string, r io.Reader) error {
		ext := filepath.Ext(name)
		// remove leading slash
		name = name[1:]
		key := fmt.Sprintf("%s/%s/%s", pkgName, version, name)

		content, err := ioutil.ReadAll(r)
		if err != nil {
			return errors.Wrap(err, "could not read file")
		}

		if ext == "."+FILE_EXTENSION {
			keys = append(keys, key)

			meta := newMetadata(len(content))

			s3Object := s3.PutObjectInput{
				Body:     bytes.NewReader(content),
				Bucket:   bucket,
				Key:      aws.String(key),
				Metadata: meta,
			}
			if err := uploadFile(ctx, s3Client, &s3Object); err != nil {
				return errors.Wrap(err, "failed to upload file")
			}
		}
		return nil
	}
	if err := gcp.Inflate(bytes.NewReader(archive), onFile); err != nil {
		return fmt.Errorf("could not inflate archive: %s", err)
	}

	if len(keys) == 0 {
		log.Printf("%s: no files to publish\n", pkgName)
	}

	pkg := new(packages.Package)
	if err := json.Unmarshal([]byte(configStr), &pkg); err != nil {
		return fmt.Errorf("failed to parse config: %s", err)
	}

	if err := audit.WroteR2(ctx, pkgName, version, keys, FILE_EXTENSION); err != nil {
		log.Printf("failed to audit: %s\n", err)
	}
	if err := metrics.NewUpdatePublishedR2(FILE_EXTENSION); err != nil {
		return errors.Wrap(err, "could not report metrics")
	}

	return nil
}

func newMetadata(size int) map[string]string {
	lastModifiedTime := time.Now()
	lastModifiedSeconds := lastModifiedTime.UnixNano() / int64(time.Second)
	lastModifiedStr := lastModifiedTime.Format(http.TimeFormat)
	etag := fmt.Sprintf("%x-%x", lastModifiedSeconds, size)

	meta := make(map[string]string)

	// https://github.com/cdnjs/origin-worker/blob/ff91d30586c9e924ff919407401dff6f52826b4d/src/index.js#L212-L213
	meta["etag"] = etag
	meta["last_modified"] = lastModifiedStr

	return meta
}

func uploadFile(ctx context.Context, s3Client *s3.Client, obj *s3.PutObjectInput) error {
	if _, err := s3Client.PutObject(ctx, obj); err != nil {
		return errors.Wrapf(err, "failed to put Object %s", *obj.Key)
	}

	return nil
}
