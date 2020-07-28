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

	// enforce schema when reading non-human package JSON
	return packages.ReadNonHumanPackageJSONBytes(ctx, key, bytes)
}

// Gets the request to update a package metadata entry in KV with a new version.
// Must have the `version` field by now.
func UpdateKVPackage(ctx context.Context, p *packages.Package) error {
	v, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal KV package JSON: %s", *p.Name)
	}
	fmt.Printf("%s\n", v)
	// enforce schema when writing non-human package JSON
	_, err = packages.ReadNonHumanPackageJSONBytes(ctx, *p.Name, v)
	if err != nil {
		return err
	}

	// req := &writeRequest{
	// 	key:   *p.Name,
	// 	value: v,
	// }

	fmt.Printf("%s\n", v)
	return nil
	//return encodeAndWriteKVBulk(ctx, []*writeRequest{req}, packagesNamespaceID)
}
