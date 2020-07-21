package kv

import (
	"context"
	"encoding/json"
)

const (
	rootKey = "/packages"
)

// Root contains the list of all packages.
type Root struct {
	Packages []string `json:"packages"`
}

// Package contains the list of versions
// for a particular package.
type Package struct {
	// TODO: - add package-level metadata
	// 	     - clean up/copy from packages.min.js generation
	// 		 - eventually remove packages.min.js entirely
	Versions []string `json:"versions"`
}

// Version contains the list of Files for a
// particular version.
// For now, it is just a []string, but may be a struct containing
// further metadata in the future.
type Version []string

// GetRoot gets the root node in KV containing the list of packages.
func GetRoot() (Root, error) {
	var r Root
	bytes, err := ReadMetadata(rootKey)
	if err != nil {
		return r, err
	}
	err = json.Unmarshal(bytes, &r)
	return r, err
}

// GetPackage gets the package metadata from KV.
func GetPackage(key string) (Package, error) {
	var p Package
	bytes, err := ReadMetadata(key)
	if err != nil {
		return p, err
	}
	err = json.Unmarshal(bytes, &p)
	return p, err
}

// GetVersion gets the version metadata from KV.
func GetVersion(key string) (Version, error) {
	var v Version
	bytes, err := ReadMetadata(key)
	if err != nil {
		return v, err
	}
	err = json.Unmarshal(bytes, &v)
	return v, err
}

// Updates KV with new version's metadata, writing to all of the necessary data structures.
// The []string of `fromVersionPaths` will already contain the optimized/minified files by now.
func updateKVMetadata(ctx context.Context, pkg, version string, fromVersionPaths []string) error {
	var kvs []*writeRequest

	// write bulk to KV
	return encodeAndWriteKVBulk(ctx, kvs, metadataNamespaceID)
}
