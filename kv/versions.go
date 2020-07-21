package kv

import (
	"context"
	"encoding/json"
	"path"
	"sort"

	"github.com/cdnjs/tools/util"
)

// Gets the request to update a version entry in KV with a number of file assets.
// Note: for now, a `version` entry is just a []string of assets, but this could become
// a struct if more metadata is added.
func updateVersionRequest(pkg, version string, fromVersionPaths []string) *writeRequest {
	key := path.Join(pkg, version)

	sort.Strings(fromVersionPaths)
	v, err := json.Marshal(fromVersionPaths)
	util.Check(err)

	return &writeRequest{
		key:   key,
		value: v,
	}
}

// Updates KV with new version's metadata.
// The []string of `fromVersionPaths` will already contain the optimized/minified files by now.
func updateKVVersion(ctx context.Context, pkg, version string, fromVersionPaths []string) error {
	req := updateVersionRequest(pkg, version, fromVersionPaths)
	return encodeAndWriteKVBulk(ctx, []*writeRequest{req}, versionsNamespaceID)
}
