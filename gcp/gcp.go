package gcp

import (
	"os"
	"path"

	"github.com/cdnjs/tools/util"

	"golang.org/x/net/context"
	"google.golang.org/api/option"

	"cloud.google.com/go/storage"
)

func getCredentialsFile() string {
	home, err := os.UserHomeDir()
	util.Check(err)

	return path.Join(home, "google_storage_cdnjs_assets.json")
}

// GetStorageClient gets the GCP Storage Client.
func GetStorageClient(ctx context.Context) (*storage.Client, error) {
	return storage.NewClient(ctx, option.WithCredentialsFile(getCredentialsFile()))
}
