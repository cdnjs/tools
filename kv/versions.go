package kv

import (
	"context"
	"encoding/json"
	"path"
	"sort"

	"github.com/cdnjs/tools/util"
)

// GetVersions gets the list of KV version keys for a particular package.
func GetVersions(pckgname string) ([]string, error) {
	return listByPrefixNamesOnly(pckgname+"/", versionsNamespaceID)
}

// GetVersion gets metadata for a particular version.
func GetVersion(ctx context.Context, key string) ([]string, error) {
	bytes, err := read(key, versionsNamespaceID)
	if err != nil {
		return nil, err
	}
	var assets []string
	err = json.Unmarshal(bytes, &assets)
	return assets, err
}

// Gets the request to update a version entry in KV with a number of file assets.
// Note: for now, a `version` entry is just a []string of assets, but this could become
// a struct if more metadata is added.
func updateVersionRequest(pkg, version string, fromVersionPaths []string) ([]string, *writeRequest) {
	key := path.Join(pkg, version)

	sort.Strings(fromVersionPaths)
	v, err := json.Marshal(fromVersionPaths)
	util.Check(err)

	return fromVersionPaths, &writeRequest{
		key:   key,
		value: v,
	}
}

// Updates KV with new version's metadata.
// The []string of `fromVersionPaths` will already contain the optimized/minified files by now.
func updateKVVersion(ctx context.Context, pkg, version string, fromVersionPaths []string) ([]string, []byte, error) {
	fromVersionPaths, req := updateVersionRequest(pkg, version, fromVersionPaths)
	_, err := encodeAndWriteKVBulk(ctx, []*writeRequest{req}, versionsNamespaceID)
	return fromVersionPaths, req.value, err
}
