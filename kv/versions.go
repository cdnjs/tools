package kv

import (
	"context"
	"encoding/json"
	"path"
	"strings"

	"github.com/cdnjs/tools/util"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
)

// GetVersions gets the list of KV version keys for a particular package.
func GetVersions(api *cloudflare.API, pckgname string) ([]string, error) {
	list, err := listByPrefixNamesOnly(api, pckgname+"/", versionsNamespaceID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list versions")
	}

	versions := make([]string, len(list))
	for i, item := range list {
		parts := strings.Split(item, "/")
		versions[i] = parts[1]
	}

	return versions, nil
}

// // GetVersion gets metadata for a particular version.
func GetVersion(ctx context.Context, api *cloudflare.API, key string) ([]string, error) {
	bytes, err := read(api, key, versionsNamespaceID)
	if err != nil {
		return nil, err
	}
	var assets []string
	err = json.Unmarshal(bytes, &assets)
	return assets, err
}

// // Gets the request to update a version entry in KV with a number of file assets.
// // Note: for now, a `version` entry is just a []string of assets, but this could become
// // a struct if more metadata is added.
func updateVersionRequest(pkg, version string, files []string) *WriteRequest {
	key := path.Join(pkg, version)

	v, err := json.Marshal(files)
	util.Check(err)

	return &WriteRequest{
		Key:   key,
		Value: v,
	}
}

// // Updates KV with new version's metadata.
// // The []string of `files` will already contain the optimized/minified files by now.
func UpdateKVVersion(ctx context.Context, api *cloudflare.API, pkg, version string, files []string) ([]byte, error) {
	req := updateVersionRequest(pkg, version, files)
	_, err := EncodeAndWriteKVBulk(ctx, api, []*WriteRequest{req}, versionsNamespaceID, true)
	return req.Value, err
}
