package kv

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"

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
	// // these file extensions will be uploaded to KV
	// // but not compessed
	// doNotCompress = map[string]bool{
	// 	".woff":  true,
	// 	".woff2": true,
	// }
)

// Gets the requests to update a number of files in KV.
// In order to do this, it will create a brotli and gzip version for each uncompressed file
// that is not banned (ex. `.woff2`, `.br`, `.gz`).
func getFileWriteRequests(ctx context.Context, pkg, version, fullPathToVersion string, fromVersionPaths []string) ([]*writeRequest, error) {
	baseVersionPath := path.Join(pkg, version)
	var kvs []*writeRequest

	for _, fromVersionPath := range fromVersionPaths {
		ext := path.Ext(fromVersionPath)
		if _, ok := ignored[ext]; ok {
			util.Debugf(ctx, "file ignored from kv write: %s\n", fromVersionPath)
			continue // ignore completely
		}
		fullPath := path.Join(fullPathToVersion, fromVersionPath)
		baseFileKey := path.Join(baseVersionPath, fromVersionPath)

		// stat file
		info, err := os.Stat(fullPath)
		if err != nil {
			return kvs, err
		}

		// read file bytes
		bytes, err := ioutil.ReadFile(fullPath)
		if err != nil {
			return kvs, err
		}

		// set metadata
		lastModifiedTime := info.ModTime()
		lastModifiedSeconds := lastModifiedTime.UnixNano() / int64(time.Second)
		lastModifiedStr := lastModifiedTime.Format(http.TimeFormat)
		etag := fmt.Sprintf("%x-%x", lastModifiedSeconds, info.Size())

		meta := &FileMetadata{
			ETag:         etag,
			LastModified: lastModifiedStr,
		}

		// TODO: Stop pushing uncompressed when Worker is fixed.
		//if _, ok := doNotCompress[ext]; ok {
		// will insert to KV without compressing further
		//	util.Debugf(ctx, "file will not be compressed in kv write: %s\n", fromVersionPath)
		kvs = append(kvs, &writeRequest{
			key:   baseFileKey,
			value: bytes,
			meta:  meta,
		})
		//	continue
		//}

		// brotli
		kvs = append(kvs, &writeRequest{
			key:   baseFileKey + ".br",
			value: compress.Brotli11CLI(ctx, fullPath),
			meta:  meta,
		})

		// gzip
		kvs = append(kvs, &writeRequest{
			key:   baseFileKey + ".gz",
			value: compress.Gzip9Native(bytes),
			meta:  meta,
		})
	}

	return kvs, nil
}

// Optimizes/minifies files on disk for a particular package version.
// Note that the package's metadata in KV must be updated before this function call
// (ex. whether or not to optimize PNG).
//
// TODO:
// Eventually remove the autoupdater's `compressNewVersion()` function,
// as we will not depend on disk files such as `.donotoptimizepng`.
// Also remove `filterByExt()` since it is cleaner with a switch.
// func optimizeAndMinify(ctx context.Context, pkg, fullPathToVersion string, fromVersionPaths []string) ([]string, error) {
// 	for _, fromV := range fromVersionPaths {
// 		fullPath := path.Join(fullPathToVersion, fromV)
// 		switch path.Ext(fromV) {
// 		case ".jpg", ".jpeg":
// 			compress.Jpeg(ctx, fullPath)
// 		case ".png":
// 			compress.Png(ctx, fullPath)
// 		case ".js":
// 			compress.Js(ctx, fullPath)
// 		case ".css":
// 			compress.CSS(ctx, fullPath)
// 		}
// 	}
// 	return util.ListFilesInVersion(ctx, fullPathToVersion)
// }

// Updates KV with new version's files.
func updateKVFiles(ctx context.Context, pkg, version, fullPathToVersion string, fromVersionPaths []string) ([]string, error) {
	// minify/optimize existing files, adding any new files generated (ex: .min.js)
	// note: encoding in brotli/gzip will occur later for each of these files

	// TODO - FIX SO OCCURS IN MEMORY TO AVOID GIT SIDE EFFECTS
	// fromVersionPaths, err := optimizeAndMinify(ctx, pkg, fullPathToVersion, fromVersionPaths)
	// if err != nil {
	// 	return fromVersionPaths, err
	// }

	// create bulk of requests
	reqs, err := getFileWriteRequests(ctx, pkg, version, fullPathToVersion, fromVersionPaths)
	if err != nil {
		return fromVersionPaths, err
	}

	// write bulk to KV
	return fromVersionPaths, encodeAndWriteKVBulk(ctx, reqs, filesNamespaceID)
}
