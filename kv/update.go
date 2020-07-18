package kv

import (
	"context"
	"io/ioutil"
	"path"

	"github.com/cdnjs/tools/compress"
	"github.com/cdnjs/tools/util"
)

var (
	// these file extensions are ignored and will not
	// be compressed or uploaded to KV
	ignored = map[string]bool{
		".br": true,
		".gz": true,
	}
	// these file extensions will be uploaded to KV
	// but not compessed
	doNotCompress = map[string]bool{
		".woff":  true,
		".woff2": true,
	}
)

// // Perform a binary search, inserting a string into the sorted list if not present.
// func insertToSortedListIfNotPresent(sorted []string, s string) []string {
// 	i := sort.SearchStrings(sorted, s)
// 	if i == len(sorted) {
// 		return append(sorted, s) // insert at back of list
// 	}
// 	if sorted[i] == s {
// 		return sorted // already exists in list
// 	}
// 	return append(sorted[:i], append([]string{s}, sorted[i:]...)...) // insert to list
// }

// // Gets the request to update the root entry in KV with a new package.
// func updateRootRequest(pkg string) *writeRequest {
// 	r, err := GetRoot()
// 	if err != nil {
// 		// assume key not found or malformed JSON
// 		// so we will rewrite this entry
// 		r.Packages = []string{pkg}
// 	} else {
// 		r.Packages = insertToSortedListIfNotPresent(r.Packages, pkg)
// 	}

// 	v, err := json.Marshal(r)
// 	util.Check(err)

// 	return &writeRequest{
// 		Key:   rootKey,
// 		Value: v,
// 	}
// }

// // Gets the request to update a package entry in KV with a new version.
// func updatePackageRequest(pkg, version string) *writeRequest {
// 	key := pkg
// 	p, err := GetPackage(key)
// 	if err != nil {
// 		// assume key not found or malformed JSON
// 		// so we will rewrite this entry
// 		p.Versions = []string{version}
// 	} else {
// 		p.Versions = insertToSortedListIfNotPresent(p.Versions, version)
// 	}

// 	v, err := json.Marshal(p)
// 	util.Check(err)

// 	return &writeRequest{
// 		Key:   key,
// 		Value: v,
// 	}
// }

// // Gets the request to update a version entry in KV with a number of Files.
// func updateVersionRequest(pkg, version string, files []File) *writeRequest {
// 	key := path.Join(pkg, version)

// 	v, err := json.Marshal(Version{Files: files})
// 	util.Check(err)

// 	return &writeRequest{
// 		Key:   key,
// 		Value: v,
// 	}
// }

func getMetadata(payload []byte) (*Metadata, error) {
	return &Metadata{}, nil
}

// Gets the requests to update a number of files in KV in compressed format.
// In order to do this, it will create a brotli and gzip version for each uncompressed file
// that is not banned (ex. `.woff2`, `.br`, `.gz`).
//
// TODO:
// Should SRIs be calculated for all files, including compressed ones, or just uncompressed files?
func updateCompressedFilesRequests(ctx context.Context, pkg, version, fullPathToVersion string, fromVersionPaths []string) ([]*writeRequest, []File, error) {
	baseVersionPath := path.Join(pkg, version)
	var kvs []*writeRequest
	var compressedFiles []File // legacy (currently unused)

	for _, fromVersionPath := range fromVersionPaths {
		ext := path.Ext(fromVersionPath)
		if _, ok := ignored[ext]; ok {
			util.Debugf(ctx, "file ignored from kv write: %s\n", fromVersionPath)
			continue // ignore completely
		}
		fullPath := path.Join(fullPathToVersion, fromVersionPath)
		baseFileKey := path.Join(baseVersionPath, fromVersionPath)

		bytes, err := ioutil.ReadFile(fullPath)
		if err != nil {
			return kvs, compressedFiles, err
		}

		if _, ok := doNotCompress[ext]; ok {
			// will insert to KV without compressing further
			util.Debugf(ctx, "file will not be compressed in kv write: %s\n", fromVersionPath)
			kvs = append(kvs, &writeRequest{
				key:   baseFileKey,
				value: bytes,
			})
			// compressedFiles = append(compressedFiles, File{
			// 	Name: fromVersionPath,
			// 	// TODO: determine metadata
			// })
			continue
		}

		// brotli
		kvs = append(kvs, &writeRequest{
			key:   baseFileKey + ".br",
			value: compress.Brotli11CLI(ctx, fullPath),
		})
		// compressedFiles = append(compressedFiles, File{
		// 	Name: fromVersionPath + ".br",
		// 	// TODO: determine metadata
		// })

		// gzip
		kvs = append(kvs, &writeRequest{
			key:   baseFileKey + ".gz",
			value: compress.Gzip9Native(bytes),
		})
		// compressedFiles = append(compressedFiles, File{
		// 	Name: fromVersionPath + ".gz",
		// 	// TODO: determine metadata
		// })
	}

	return kvs, compressedFiles, nil
}

// Optimizes/minifies files on disk for a particular package version.
// Note that the package's metadata in KV must be updated before this function call
// (ex. whether or not to optimize PNG).
//
// TODO:
// Eventually remove the autoupdater's `compressNewVersion()` function,
// as we will not depend on disk files such as `.donotoptimizepng`.
// Also remove `filterByExt()` since it is cleaner with a switch.
func optimizeAndMinify(ctx context.Context, pkg, fullPathToVersion string, fromVersionPaths []string) ([]string, error) {
	for _, fromV := range fromVersionPaths {
		fullPath := path.Join(fullPathToVersion, fromV)
		switch path.Ext(fromV) {
		case ".jpg", ".jpeg":
			compress.Jpeg(ctx, fullPath)
		case ".png":
			compress.Png(ctx, fullPath)
		case ".js":
			compress.Js(ctx, fullPath)
		case ".css":
			compress.CSS(ctx, fullPath)
		}
	}
	return util.ListFilesInVersion(ctx, fullPathToVersion)
}

// Updates KV with new version, writing to all of the necessary data structures.
//
// TODO:
// Will want to push to a queue or write to disk journal somewhere
// when an operation is about to be attempted and when an
// operation completes successfully. This is to help recover from
// silent failures that result in inconsistent states.
func updateKV(ctx context.Context, pkg, version, fullPathToVersion string, fromVersionPaths []string) error {
	// minify/optimize existing files, adding any new files generated (ex: .min.js)
	// note: encoding in brotli/gzip will occur later for each of these files
	fromVersionPaths, err := optimizeAndMinify(ctx, pkg, fullPathToVersion, fromVersionPaths)
	if err != nil {
		return err
	}

	// create bulk of requests
	var kvs []*writeRequest
	compressedReqs, _, err := updateCompressedFilesRequests(ctx, pkg, version, fullPathToVersion, fromVersionPaths)
	if err != nil {
		return err
	}

	kvs = append(kvs, compressedReqs...)

	// TODO: decide on how much metadata will be maintained
	// allFiles := append(uncompressedFiles, compressedFiles...)
	// sort.Slice(allFiles, func(i, j int) bool { return allFiles[i].Name < allFiles[j].Name })
	// kvs = append(kvs, updateVersionRequest(pkg, version, allFiles))
	// kvs = append(kvs, updatePackageRequest(pkg, version))
	// kvs = append(kvs, updateRootRequest(pkg))

	// write bulk to KV
	return encodeAndWriteKVBulk(ctx, kvs)
}
