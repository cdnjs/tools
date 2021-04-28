package kv

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/cdnjs/tools/compress"
	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/util"

	cloudflare "github.com/cloudflare/cloudflare-go"
)

// UpdateAggregatedMetadata updates a package's KV entry for aggregated metadata.
// Returns the keys written to KV, whether the existing entry was found, and if there were any errors.
func UpdateAggregatedMetadata(api *cloudflare.API, ctx context.Context,
	pkgName string, newAssets packages.Asset) ([]string, bool, error) {
	pkg, err := getAggregatedMetadata(api, pkgName)
	var found bool
	if err != nil {
		switch err.(type) {
		case KeyNotFoundError:
			{
				// key not found (new package)
				log.Printf("KV key `%s` not found, inserting aggregated metadata...\n", pkgName)
				pkg.Assets = []packages.Asset{newAssets}
			}
		default:
			{
				return nil, false, err
			}
		}
	} else {
		log.Printf("Aggregated metadata for `%s` found. Updating aggregated metadata...\n", pkgName)
		pkg.Assets = append(pkg.Assets, newAssets)
		found = true
	}

	successfulWrites, err := writeAggregatedMetadata(ctx, api, pkg)
	return successfulWrites, found, err
}

// Reads an aggregated metadata entry in KV, ungzipping it and
// unmarshalling it into a *packages.Package.
func getAggregatedMetadata(api *cloudflare.API, key string) (*packages.Package, error) {
	gzipBytes, err := read(api, key, aggregatedMetadataNamespaceID)

	if err != nil {
		return nil, err
	}

	// unmarshal and ungzip
	var p packages.Package
	util.Check(json.Unmarshal(compress.UnGzip(gzipBytes), &p))

	return &p, nil
}

// Writes an aggregated metadata entry to KV, gzipping the bytes.
func writeAggregatedMetadata(ctx context.Context, api *cloudflare.API, p *packages.Package) ([]string, error) {
	// marshal package into JSON
	v, err := p.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal KV package JSON: %s", *p.Name)
	}

	// gzip the bytes
	req := &WriteRequest{
		Name:  *p.Name,
		Key:   *p.Name,
		Value: compress.Gzip9Bytes(v),
	}

	// write aggregated to KV
	return EncodeAndWriteKVBulk(ctx, api, []*WriteRequest{req}, aggregatedMetadataNamespaceID, true)
}
