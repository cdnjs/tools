package kv

import (
	"context"
	"encoding/json"

	"github.com/cdnjs/tools/util"
)

// Package contains the list of versions
// for a particular package.
type Package struct {
	// TODO: - add package-level metadata
	// 	     - clean up/copy from packages.min.js generation
	// 		 - eventually remove packages.min.js entirely

	Name    string `json:"name"`    // just for testing
	Version string `json:"version"` // latest version
}

// GetPackage gets the package metadata from KV.
//
// TODO:
// Currently unused. Will be used by ReadPackageJSON in the future.
func GetPackage(key string) (Package, error) {
	var p Package
	bytes, err := Read(key, packagesNamespaceID)
	if err != nil {
		return p, err
	}
	err = json.Unmarshal(bytes, &p)
	return p, err
}

// Gets the request to update a package metadata entry in KV with a new version.
// TODO:
// In the future we can probably just take a Package as an argument here,
// since ReadPackageJSON will probably just read this Package from KV.
func UpdateKVPackage(ctx context.Context, pkg, version string) error {
	p := Package{
		Name:    pkg,
		Version: version,
	}

	v, err := json.Marshal(p)
	util.Check(err)

	req := &writeRequest{
		key:   pkg,
		value: v,
	}

	return encodeAndWriteKVBulk(ctx, []*writeRequest{req}, packagesNamespaceID)
}
