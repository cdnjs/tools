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

// GetVersionsFromAggregatedMetadata gets the list version for a particular package
// using the aggregated metadata endpoint.
//
// The aggregated metadata will only contain non-empty versions, so this is useful
// for updating Algolia.
func GetVersionsFromAggregatedMetadata(api *cloudflare.API, pckgname string) ([]string, error) {
	aggPkg, err := getAggregatedMetadata(api, pckgname)
	if err != nil {
		switch err.(type) {
		case KeyNotFoundError:
			{
				return nil, nil
			}
		default:
			{
				// api error
				return nil, err
			}
		}
	}

	var versions []string
	for _, asset := range aggPkg.Assets {
		versions = append(versions, asset.Version)
	}

	return versions, nil
}

// RemoveVersionFromAggregatedMetadata will remove a particular version from
// a package's KV entry for aggregated metadata if it exists.
// This is useful for removing empty versions with no files.
func RemoveVersionFromAggregatedMetadata(api *cloudflare.API, ctx context.Context, pkg *packages.Package, version string) ([]string, bool, error) {
	aggPkg, err := getAggregatedMetadata(api, *pkg.Name)
	if err != nil {
		switch err.(type) {
		case KeyNotFoundError:
			{
				// key not found
				log.Printf("Removing version %s from aggregated metadata: KV key `%s` not found, ignoring\n", version, *pkg.Name)
				return nil, false, nil
			}
		default:
			{
				// api error
				return nil, false, err
			}
		}
	}

	if !aggPkg.HasVersion(version) {
		log.Printf("Removing version %s from aggregated metadata: version does not exist\n", version)
		return nil, false, nil
	}

	// remove the version
	log.Printf("Removing version %s from aggregated metadata: version found\n", version)
	aggPkg.RemoveVersion(version)

	successfulWrites, err := writeAggregatedMetadata(ctx, api, aggPkg)
	return successfulWrites, true, err
}

// UpdateAggregatedMetadata updates a package's KV entry for aggregated metadata.
// Returns the keys written to KV, whether the existing entry was found, and if there were any errors.
func UpdateAggregatedMetadata(api *cloudflare.API, ctx context.Context,
	pkg *packages.Package, newVersion string, newAssets packages.Asset) ([]string, bool, error) {
	aggPkg, err := getAggregatedMetadata(api, *pkg.Name)

	if aggPkg == nil {
		// pkg has never been aggregated
		aggPkg = pkg
	}

	var found bool
	if err != nil {
		switch err.(type) {
		case KeyNotFoundError:
			{
				// key not found (new package)
				log.Printf("KV key `%s` not found, inserting aggregated metadata...\n", *pkg.Name)
				aggPkg.Assets = []packages.Asset{newAssets}
			}
		default:
			{
				return nil, false, err
			}
		}
	} else {
		if !aggPkg.HasVersion(newVersion) {
			aggPkg.Assets = append(aggPkg.Assets, newAssets)
			log.Printf("Aggregated metadata for `%s` found. Updating aggregated metadata...\n", *pkg.Name)
		} else {
			log.Printf("Aggregated metadata for `%s` found. Version already exists, updating\n", *pkg.Name)
			aggPkg.UpdateVersion(newVersion, newAssets)
		}
		found = true
	}
	aggPkg.Version = &newVersion

	successfulWrites, err := writeAggregatedMetadata(ctx, api, aggPkg)
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
	req := &ConsumableWriteRequest{
		Name:  *p.Name,
		Key:   *p.Name,
		Value: compress.Gzip9Bytes(v),
	}

	// write aggregated to KV
	return EncodeAndWriteKVBulk(ctx, api, []WriteRequest{req}, aggregatedMetadataNamespaceID, true)
}
