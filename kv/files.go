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
	"github.com/cdnjs/tools/sri"
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
	// but not compressed
	doNotCompress = map[string]bool{
		".woff2": true,
	}
	// we calculate SRIs for these file extensions
	calculateSRI = map[string]bool{
		".js":  true,
		".css": true,
	}
)

// GetFiles gets the list of KV file keys for a particular package.
// The `key` must be the package/version (ex. `a-happy-tyler/1.0.0`)
func GetFiles(key string) ([]string, error) {
	return listByPrefixNamesOnly(key+"/", filesNamespaceID)
}

// Gets the requests to update a number of files in KV, as well as the files' SRIs.
// In order to do this, it will create a brotli and gzip version for each uncompressed file
// that is not banned (ex. `.woff2`, `.br`, `.gz`).
// Returns the list of requests for pushing SRIs and list of requests for pushing files to KV.
func getFileWriteRequests(ctx context.Context, pkg, version, fullPathToVersion string, fromVersionPaths []string, srisOnly bool) ([]*writeRequest, []*writeRequest, error) {
	baseVersionPath := path.Join(pkg, version)
	var sriKVs, fileKVs []*writeRequest

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
			return nil, nil, err
		}

		// read file bytes
		bytes, err := ioutil.ReadFile(fullPath)
		if err != nil {
			return nil, nil, err
		}

		if _, ok := calculateSRI[ext]; ok {
			sriKVs = append(sriKVs, &writeRequest{
				key:  baseFileKey,
				name: fromVersionPath,
				meta: &FileMetadata{
					SRI: sri.CalculateSRI(bytes),
				},
			})
		}

		if srisOnly {
			continue
		}

		// set metadata
		lastModifiedTime := info.ModTime()
		lastModifiedSeconds := lastModifiedTime.UnixNano() / int64(time.Second)
		lastModifiedStr := lastModifiedTime.Format(http.TimeFormat)
		etag := fmt.Sprintf("%x-%x", lastModifiedSeconds, info.Size())

		fileMeta := &FileMetadata{
			ETag:         etag,
			LastModified: lastModifiedStr,
		}

		if _, ok := doNotCompress[ext]; ok {
			// will only insert uncompressed to KV
			fileKVs = append(fileKVs, &writeRequest{
				key:   baseFileKey,
				name:  fromVersionPath,
				value: bytes,
				meta:  fileMeta,
			})
			continue
		}

		// brotli
		fileKVs = append(fileKVs, &writeRequest{
			key:   baseFileKey + ".br",
			name:  fromVersionPath + ".br",
			value: compress.Brotli11CLI(ctx, fullPath),
			meta:  fileMeta,
		})

		// gzip
		fileKVs = append(fileKVs, &writeRequest{
			key:   baseFileKey + ".gz",
			name:  fromVersionPath + ".gz",
			value: compress.Gzip9Native(bytes),
			meta:  fileMeta,
		})
	}

	return sriKVs, fileKVs, nil
}

// Updates KV with new version's files.
// The []string of `fromVersionPaths` will already contain the optimized/minified files by now.
// The function will return the list of SRIs pushed to KV and the list of all files pushed to KV.
func updateKVFiles(ctx context.Context, pkg, version, fullPathToVersion string, fromVersionPaths []string, srisOnly, filesOnly, noPush, panicOversized bool) ([]string, []string, int, int, error) {
	// create bulk of requests
	sriReqs, fileReqs, err := getFileWriteRequests(ctx, pkg, version, fullPathToVersion, fromVersionPaths, srisOnly)
	if err != nil {
		return nil, nil, 0, 0, err
	}
	theoreticalSRIsKeys, theoreticalFilesKeys := len(sriReqs), len(fileReqs)

	var successfulSRIWrites []string
	if !filesOnly && !noPush {
		// write SRIs bulk to KV
		successfulSRIWrites, err = encodeAndWriteKVBulk(ctx, sriReqs, srisNamespaceID, panicOversized)
		if err != nil {
			return nil, nil, 0, 0, err
		}
		if srisOnly {
			return successfulSRIWrites, nil, 0, 0, nil
		}
	}

	successfulFileWrites, err := encodeAndWriteKVBulk(ctx, fileReqs, filesNamespaceID, panicOversized)
	return successfulSRIWrites, successfulFileWrites, theoreticalSRIsKeys, theoreticalFilesKeys, err
}
