package kv

// import (
// 	"context"
// 	"fmt"

// 	"github.com/cdnjs/tools/packages"
// )

// GetPackage gets the package metadata from KV.
// It will validate against the non-human-readable schema, returning
// a packages.InvalidSchemaError if the schema is invalid, a KeyNotFoundError
// if the KV key is not found, and an AuthError if there is an authentication error.
// func GetPackage(ctx context.Context, key string) (*packages.Package, error) {
// 	bytes, err := read(key, packagesNamespaceID)

// 	if err != nil {
// 		return nil, err
// 	}

// 	// enforce schema when reading non-human package JSON
// 	return packages.ReadNonHumanJSONBytes(ctx, key, bytes)
// }

// UpdateKVPackage gets the request to update a package metadata entry in KV with a new version.
// Must have the `version` field by now.
// func UpdateKVPackage(ctx context.Context, p *packages.Package) error {
// 	// marshal package into JSON
// 	v, err := p.Marshal()
// 	if err != nil {
// 		return fmt.Errorf("failed to marshal KV package JSON: %s", *p.Name)
// 	}

// 	// enforce schema when writing non-human package JSON
// 	_, err = packages.ReadNonHumanJSONBytes(ctx, *p.Name, v)
// 	if err != nil {
// 		return err
// 	}

// 	req := &writeRequest{
// 		key:   *p.Name,
// 		value: v,
// 	}

// 	_, err = encodeAndWriteKVBulk(ctx, []*writeRequest{req}, packagesNamespaceID, true)
// 	return err
// }
