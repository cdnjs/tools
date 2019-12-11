package cloudstorage

import (
	"os"
	"path"

	"github.com/xtuc/cdnjs-go/util"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
)

func getCredentialsFile() string {
	home, err := os.UserHomeDir()
	util.Check(err)

	return path.Join(home, "google_storage_cdnjs_assets.json")
}

const bucketName = "cdnjs-assets"

func GetClient(ctx context.Context) (*storage.Client, error) {
	return storage.NewClient(ctx, option.WithCredentialsFile(getCredentialsFile()))
}

func GetBucket(ctx context.Context) (*storage.BucketHandle, error) {
	client, err := GetClient(ctx)
	util.Check(err)
	bkt := client.Bucket(bucketName)
	return bkt, nil
}
