package kv

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"sort"

	"github.com/cdnjs/tools/sri"
	"github.com/cdnjs/tools/util"
)

// Perform a binary search, inserting a string into the sorted list if not present.
func insertToSortedListIfNotPresent(sorted []string, s string) []string {
	i := sort.SearchStrings(sorted, s)
	if i == len(sorted) {
		return append(sorted, s) // insert at back of list
	}
	if sorted[i] == s {
		return sorted // already exists in list
	}
	return append(sorted[:i], append([]string{s}, sorted[i:]...)...) // insert to list
}

// Gets the request to update the root entry in KV with a new package.
func updateRootRequest(pkg string) *writeRequest {
	r, err := GetRoot()
	if err != nil {
		// assume key not found or malformed JSON
		// so we will rewrite this entry
		r.Packages = []string{pkg}
	} else {
		r.Packages = insertToSortedListIfNotPresent(r.Packages, pkg)
	}

	v, err := json.Marshal(r)
	util.Check(err)

	return &writeRequest{
		Key:   rootKey,
		Value: v,
	}
}

// Gets the request to update a package entry in KV with a new version.
func updatePackageRequest(pkg, version string) *writeRequest {
	key := pkg
	p, err := GetPackage(key)
	if err != nil {
		// assume key not found or malformed JSON
		// so we will rewrite this entry
		p.Versions = []string{version}
	} else {
		p.Versions = insertToSortedListIfNotPresent(p.Versions, version)
	}

	v, err := json.Marshal(p)
	util.Check(err)

	return &writeRequest{
		Key:   key,
		Value: v,
	}
}

// Gets the request to update a version entry in KV with a number of Files.
func updateVersionRequest(pkg, version string, files []File) *writeRequest {
	key := path.Join(pkg, version)

	v, err := json.Marshal(Version{Files: files})
	util.Check(err)

	return &writeRequest{
		Key:   key,
		Value: v,
	}
}

// Gets the requests to update a number of files in KV in compressed format.
func updateCompressedFilesRequests(uncompressedFiles []File) ([]*writeRequest, []File) {
	for _, f := range uncompressedFiles {
		fmt.Println(f)
	}
	// iterate over files, compress each via brotli/gzip unless if it is woff2
	return nil, nil
}

// Gets the requests to update a number of files in KV in uncompressed format.
func updateUncompressedFilesRequests(pkg, version, fullPathToVersion string, fromVersionPaths []string) ([]*writeRequest, []File) {
	baseKeyPath := path.Join(pkg, version)
	kvs := make([]*writeRequest, len(fromVersionPaths))
	files := make([]File, len(fromVersionPaths))

	for i, fromVersionPath := range fromVersionPaths {
		fullPath := path.Join(fullPathToVersion, fromVersionPath)
		bytes, err := ioutil.ReadFile(fullPath)
		util.Check(err)

		kvs[i] = &writeRequest{
			Key:   path.Join(baseKeyPath, fromVersionPath),
			Value: bytes,
		}

		files[i] = File{
			Name: fromVersionPath,
			SRI:  sri.CalculateFileSRI(fullPath),
		}
	}

	return kvs, files
}

// TODO:
// Will want to push to a queue or write to disk journal somewhere
// when an operation is about to be attempted and when an
// operation completes successfully. This is to help recover from
// silent failures that result in inconsistent states.
func pushToTaskQueue(_ string) {
}

// Updates KV with new version, writing to all of the necessary data structures.
func updateKV(pkg, version, fullPathToVersion string, fromVersionPaths []string) {
	var kvs []*writeRequest
	uncompressedReqs, uncompressedFiles := updateUncompressedFilesRequests(pkg, version, fullPathToVersion, fromVersionPaths)
	compressedReqs, compressedFiles := updateCompressedFilesRequests(uncompressedFiles)
	kvs = append(kvs, uncompressedReqs...)
	kvs = append(kvs, compressedReqs...)
	kvs = append(kvs, updateVersionRequest(pkg, version, append(uncompressedFiles, compressedFiles...)))
	kvs = append(kvs, updatePackageRequest(pkg, version))
	kvs = append(kvs, updateRootRequest(pkg))

	pushToTaskQueue("TODO -- WRITING TO KV")
	encodeAndWriteKVBulk(kvs)
	pushToTaskQueue("TODO -- WROTE TO KV SUCCESSFULLY")
}
