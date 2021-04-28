package check_pkg_updates

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/cdnjs/tools/npm"
	"github.com/cdnjs/tools/packages"

	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
)

var (
	GCS_BUCKET = os.Getenv("GCS_BUCKET")
)

func storeGCS(fileName string, buff bytes.Buffer, pckg *packages.Package, version npm.Version) error {
	// Create GCS connection
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("HTTP response error: %v", err)
	}

	bucket := client.Bucket(GCS_BUCKET)
	obj := bucket.Object(fileName)
	w := obj.NewWriter(ctx)
	w.ACL = []storage.ACLRule{
		{Entity: storage.AllUsers, Role: storage.RoleReader},
	}

	if _, err := io.Copy(w, bytes.NewReader(buff.Bytes())); err != nil {
		return fmt.Errorf("Failed to copy to bucket: %v", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("Failed to close: %v", err)
	}

	// Files that will be copied on the CDN, intended to process-version
	var fileMap []packages.FileMap
	if pckg.Autoupdate != nil {
		fileMap = pckg.Autoupdate.FileMap
	}

	fileMapBytes, err := json.Marshal(fileMap)
	if err != nil {
		return fmt.Errorf("failed to marshal filemap: %v", err)
	}

	// update the metadata once the object is written
	_, err = obj.Update(ctx, storage.ObjectAttrsToUpdate{
		Metadata: map[string]string{
			"version":            version.Version,
			"package":            *pckg.Name,
			"autoupdate-filemap": string(fileMapBytes),
		},
	})
	if err != nil {
		return errors.Wrap(err, "could not update metadata")
	}

	return nil
}
