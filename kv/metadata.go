package kv

import (
	"context"
	"encoding/json"
	"path"
	"sort"

	"github.com/cdnjs/tools/util"
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

// Perform a binary search, inserting a string into the sorted list if not present.
func insertToSortedListIfNotPresent(sorted []string, s string) []string {
	i := sort.SearchStrings(sorted, s)
	if i == len(sorted) {
		return append(sorted, s) // insert at back of list
	}
	if sorted[i] == s {
		return sorted // already exists in list
	}
	return append(sorted[:i], append([]string{s}, sorted[i:]...)...) // insert into list
}

// Gets the request to update the root metadata entry in KV with a new package.
func updateRootRequest(pkg string) *writeRequest {
	r, err := GetRoot()
	if err != nil {
		// assume key not found or malformed JSON
		// so we will rewrite this entry
		// FIX THIS -- WE SHOULD HAVE NO ASSUMPTIONS -> PARSE THE ERROR INSTEAD!
		r.Packages = []string{pkg}
	} else {
		r.Packages = insertToSortedListIfNotPresent(r.Packages, pkg)
	}

	v, err := json.Marshal(r)
	util.Check(err)

	return &writeRequest{
		key:   rootKey,
		value: v,
	}
}

// Gets the request to update a package metadata entry in KV with a new version.
func updatePackageRequest(pkg, version string) *writeRequest {
	key := pkg
	p, err := GetPackage(key)
	if err != nil {
		// assume key not found or malformed JSON
		// so we will rewrite this entry
		// FIX THIS -- WE SHOULD HAVE NO ASSUMPTIONS -> PARSE THE ERROR INSTEAD!
		p.Versions = []string{version}
	} else {
		p.Versions = insertToSortedListIfNotPresent(p.Versions, version)
	}

	// TODO: Add additional metadata found in package.min.js generation.

	v, err := json.Marshal(p)
	util.Check(err)

	return &writeRequest{
		key:   key,
		value: v,
	}
}

// Gets the request to update a version entry in KV with a number of file assets.
func updateVersionRequest(pkg, version string, fromVersionPaths []string) *writeRequest {
	key := path.Join(pkg, version)

	sort.Strings(fromVersionPaths)
	v, err := json.Marshal(fromVersionPaths)
	util.Check(err)

	return &writeRequest{
		key:   key,
		value: v,
	}
}

// Updates KV with new version's metadata, writing to all of the necessary data structures.
// The []string of `fromVersionPaths` will already contain the optimized/minified files by now.
func updateKVMetadata(ctx context.Context, pkg, version string, fromVersionPaths []string) error {
	var kvs []*writeRequest

	kvs = append(kvs, updateRootRequest(pkg))
	kvs = append(kvs, updatePackageRequest(pkg, version))
	kvs = append(kvs, updateVersionRequest(pkg, version, fromVersionPaths))

	// write bulk to KV
	return encodeAndWriteKVBulk(ctx, kvs, metadataNamespaceID)
}
