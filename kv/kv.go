package kv

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/cdnjs/tools/util"
	cloudflare "github.com/cloudflare/cloudflare-go"
)

var (
	namespaceID = util.GetEnv("WORKERS_KV_NAMESPACE_ID")
	accountID   = util.GetEnv("WORKERS_KV_ACCOUNT_ID")
	apiKey      = util.GetEnv("WORKERS_KV_API_KEY")
	email       = util.GetEnv("WORKERS_KV_EMAIL")
	api         = getAPI()
)

// Represents a KV write request, consisting of
// a string key and []byte value.
type writeRequest struct {
	Key   string
	Value []byte
}

// Gets a new *cloudflare.API.
func getAPI() *cloudflare.API {
	a, err := cloudflare.New(apiKey, email, cloudflare.UsingAccount(accountID))
	util.Check(err)
	return a
}

// ReadKV reads from Workers KV.
func ReadKV(key string) ([]byte, error) {
	return api.ReadWorkersKV(context.Background(), namespaceID, key)
}

// Ensure a response is successful and the error is nil.
func checkSuccess(ctx context.Context, r cloudflare.Response, err interface{}) {
	util.Check(err)
	if !r.Success {
		util.Debugf(ctx, "kv fail: %v\n", r)
		panic(r)
	}
}

// Encodes a byte array to a base64 string.
func encodeToBase64(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

// Encodes key-value pairs to base64 and writes them to KV
// in multiple bulk requests, panicking if unsuccessful.
func encodeAndWriteKVBulk(ctx context.Context, kvs []*writeRequest) {
	var bulkWrites []cloudflare.WorkersKVBulkWriteRequest
	var bulkWrite []*cloudflare.WorkersKVPair
	var totalSize int64

	for _, kv := range kvs {
		if size := int64(len(kv.Value)); size > util.MaxFileSize {
			panic(fmt.Sprintf("oversized file: %s (%d)", kv.Key, size))
		}
		// Note that after encoding in base64 the size may get larger, but after decoding
		// it will be reduced, so it is okay if the size is larger than util.MaxFileSize after encoding base64.
		// However, we still need to check for the KV bulk request limit of 100MiB.
		encoded := encodeToBase64(kv.Value)
		encodedSize := int64(len(encoded))
		if totalSize+encodedSize > util.MaxBulkWritePayload {
			// Create a new bulk since we are over the limit.
			// Note, this cannot happen on the first index,
			// since util.MaxFileSize must be less than util.MaxBulkWritePayload.
			bulkWrites = append(bulkWrites, bulkWrite)
			bulkWrite = []*cloudflare.WorkersKVPair{}
			totalSize = 0
		}
		bulkWrite = append(bulkWrite, &cloudflare.WorkersKVPair{
			Key:    kv.Key,
			Value:  encoded,
			Base64: true,
		})
		totalSize += encodedSize
	}
	bulkWrites = append(bulkWrites, bulkWrite)

	for i, b := range bulkWrites {
		util.Debugf(ctx, "writing bulk %d/%d (keys=%d)...\n", i+1, len(bulkWrites), len(b))
		r, err := api.WriteWorkersKVBulk(context.Background(), namespaceID, b)
		checkSuccess(ctx, r, err)
	}
}

// InsertNewVersionToKV inserts a new version to KV.
// The `fullPathToVersion` string will be useful if the version is downloaded to
// a temporary directory, not necessarily always in `$BOT_BASE_PATH/cdnjs/ajax/libs/`.
//
// Note that this function will also compress the files, generating brotli/gzip entries
// to KV where necessary, as well as minifying js, compressing png/jpeg/css, etc.
//
// For example:
// InsertNewVersionToKV("1000hz-bootstrap-validator", "0.10.0", "/tmp/1000hz-bootstrap-validator/0.10.0")
func InsertNewVersionToKV(ctx context.Context, pkg, version, fullPathToVersion string) {
	fromVersionPaths, err := util.ListFilesInVersion(ctx, fullPathToVersion)
	util.Check(err)
	updateKV(ctx, pkg, version, fullPathToVersion, fromVersionPaths)
}
