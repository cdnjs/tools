package cloudstorage

import (
	"github.com/cdnjs/tools/gcp"
	"github.com/cdnjs/tools/util"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
)

// GetAssetsBucket gets the GCP BucketHandle for cdnjs assets.
func GetAssetsBucket(ctx context.Context) (*storage.BucketHandle, error) {
	client, err := gcp.GetStorageClient(ctx)
	util.Check(err)
	bkt := client.Bucket("cdnjs-assets")
	return bkt, nil
}

// GetRobotcdnjsBucket gets the GCP BucketHandle for robotcdnjs.
func GetRobotcdnjsBucket(ctx context.Context) (*storage.BucketHandle, error) {
	client, err := gcp.GetStorageClient(ctx)
	util.Check(err)
	bkt := client.Bucket("robotcdnjs")
	return bkt, nil
}
