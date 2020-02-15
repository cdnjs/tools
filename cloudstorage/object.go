package cloudstorage

import (
	"github.com/cdnjs/tools/gcp"
	"github.com/cdnjs/tools/util"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
)

func GetAssetsBucket(ctx context.Context) (*storage.BucketHandle, error) {
	client, err := gcp.GetStorageClient(ctx)
	util.Check(err)
	bkt := client.Bucket("cdnjs-assets")
	return bkt, nil
}

func GetRobotcdnjsBucket(ctx context.Context) (*storage.BucketHandle, error) {
	client, err := gcp.GetStorageClient(ctx)
	util.Check(err)
	bkt := client.Bucket("robotcdnjs")
	return bkt, nil
}
