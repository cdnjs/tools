package kv

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cdnjs/tools/packages"
)

// GetPackage gets the package metadata from KV.
//
// TODO:
// - currently unused, will eventually replace reading `package.json` files from disk
func GetPackage(ctx context.Context, key string) (*packages.Package, error) {
	bytes, err := Read(key, packagesNamespaceID)
	if err != nil {
		return nil, err
	}
	return packages.ReadPackageJSONBytes(ctx, key, bytes)
}

// Gets the request to update a package metadata entry in KV with a new version.
func UpdateKVPackage(ctx context.Context, p *packages.Package) error {
	v, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal KV package JSON: %s", *p.Name)
	}

	req := &writeRequest{
		key:   *p.Name,
		value: v,
	}

	return encodeAndWriteKVBulk(ctx, []*writeRequest{req}, packagesNamespaceID)
}
