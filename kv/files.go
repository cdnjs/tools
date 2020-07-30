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
	// these file extensions will be uploaded to KV
	// but not compessed
	doNotCompress = map[string]bool{
		".woff":  true,
		".woff2": true,
	}
)

// GetFiles gets the list of KV file keys for a particular package.
// The `key` must be the package/version (ex. `a-happy-tyler/1.0.0`)
func GetFiles(key string) ([]string, error) {
	return ListByPrefix(key+"/", filesNamespaceID)
}

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

		if _, ok := doNotCompress[ext]; ok {
			// will only insert uncompressed to KV
			kvs = append(kvs, &writeRequest{
				key:   baseFileKey,
				value: bytes,
				meta:  meta,
			})
			continue
		}

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

// Updates KV with new version's files.
// The []string of `fromVersionPaths` will already contain the optimized/minified files by now.
func updateKVFiles(ctx context.Context, pkg, version, fullPathToVersion string, fromVersionPaths []string) ([]string, error) {
	// create bulk of requests
	reqs, err := getFileWriteRequests(ctx, pkg, version, fullPathToVersion, fromVersionPaths)
	if err != nil {
		return fromVersionPaths, err
	}

	// write bulk to KV
	return fromVersionPaths, encodeAndWriteKVBulk(ctx, reqs, filesNamespaceID)
}
