package kv

import (
	"context"
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
	return packages.ReadNonHumanJSONBytes(ctx, key, bytes)
}

// UpdateKVPackage gets the request to update a package metadata entry in KV with a new version.
// Must have the `version` field by now.
func UpdateKVPackage(ctx context.Context, p *packages.Package) error {
	// marshal package into JSON
	v, err := p.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal KV package JSON: %s", *p.Name)
	}

	fmt.Printf("Enforcing schema for: %s\n", v)
	// enforce schema when writing non-human package JSON
	_, err = packages.ReadNonHumanJSONBytes(ctx, *p.Name, v)
	if err != nil {
		return err
	}

	req := &writeRequest{
		key:   *p.Name,
		value: v,
	}

	return encodeAndWriteKVBulk(ctx, []*writeRequest{req}, packagesNamespaceID)
}
