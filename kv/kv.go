package kv

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/blang/semver"
	"github.com/cdnjs/tools/util"
	cloudflare "github.com/cloudflare/cloudflare-go"
)

var (
	namespaceID = util.GetEnv("WORKERS_KV_NAMESPACE_ID")
	accountID   = util.GetEnv("WORKERS_KV_ACCOUNT_ID")
	apiKey      = util.GetEnv("WORKERS_KV_API_KEY")
	email       = util.GetEnv("WORKERS_KV_EMAIL")
	api         = getAPI()
	basePath    = util.GetCDNJSPackages()
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

// Gets all entries in Workers KV.
func getKVs() cloudflare.ListStorageKeysResponse {
	resp, err := api.ListWorkersKVs(context.Background(), namespaceID)
	util.Check(err)
	return resp
}

// Gets all entries in Workers KV with options (ex. set the limit).
func getKVsWithOptions(o cloudflare.ListWorkersKVsOptions) cloudflare.ListStorageKeysResponse {
	resp, err := api.ListWorkersKVsWithOptions(context.Background(), namespaceID, o)
	util.Check(err)
	return resp
}

// ReadKV reads from Workers KV.
func ReadKV(key string) ([]byte, error) {
	return api.ReadWorkersKV(context.Background(), namespaceID, key)
}

// Ensure a response is successful and the error is nil.
func checkSuccess(r cloudflare.Response, err interface{}) {
	if !r.Success {
		panic(r)
	}
	util.Check(err)
}

// Writes an entry to Workers KV, panicking if unsuccessful.
func writeKV(k string, v []byte) {
	r, err := api.WriteWorkersKV(context.Background(), namespaceID, k, v)
	checkSuccess(r, err)
}

// Encodes a byte array to a base64 string.
func encodeToBase64(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

// Encodes key-value pairs to base64 and writes them to KV
// in multiple bulk requests, panicking if unsuccessful.
func encodeAndWriteKVBulk(kvs []*writeRequest) {
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

	for _, b := range bulkWrites {
		// util.Debugf(ctx, "writing bulk %d/%d (keys=%d)...\n", i+1, len(bulkWrites), len(b))
		r, err := api.WriteWorkersKVBulk(context.Background(), namespaceID, b)
		checkSuccess(r, err)
	}
}

// Deletes all entries for the KV namespace.
// This is used for testing purposes only.
func deleteAllEntries() {
	resp := getKVs()

	// make []string of keys
	keys := make([]string, len(resp.Result))
	for i, res := range resp.Result {
		keys[i] = res.Name
	}

	// TODO: change to api.DeleteWorkersKVsBulk after PR is merged
	for _, key := range keys {
		resp, err := api.DeleteWorkersKV(context.Background(), namespaceID, key)
		checkSuccess(resp, err)
		fmt.Printf("Deleted %s\n", key)
	}
}

// fullpath will be useful if the version is downloaded into a temp directory
// so it is not just path.Join(basePath, pkg, version)
func insertVersionToKV(pkg, version, fullPathToVersion string) {
	fromVersionPaths, err := util.ListFilesInVersion(context.Background(), fullPathToVersion)
	util.Check(err)
	updateKV(pkg, version, fullPathToVersion, fromVersionPaths)
}

// test
func deleteAllAndInsertPkgs() {
	deleteAllEntries()

	const maxPkgs = 10

	//insertVersionToKV("1000hz-bootstrap-validator", "0.10.0", "/Users/tylercaslin/go/src/fake-smaller-repo/cdnjs/ajax/libs/1000hz-bootstrap-validator/0.10.0")
	//insertVersionToKV("1000hz-bootstrap-validator", "0.10.0", "/Users/tylercaslin/go/src/fake-smaller-repo/cdnjs/ajax/libs/1000hz-bootstrap-validator/0.10.0")

	pkgs, err := ioutil.ReadDir(basePath)
	util.Check(err)

	for i, pkg := range pkgs {
		if i > maxPkgs {
			return
		}
		if pkg.IsDir() {
			versions, err := ioutil.ReadDir(path.Join(basePath, pkg.Name()))
			util.Check(err)

			for _, version := range versions {
				if _, err := semver.Parse(version.Name()); err == nil {
					fmt.Printf("Inserting %s (%s)\n", pkg.Name(), version.Name())
					insertVersionToKV(pkg.Name(), version.Name(), path.Join(basePath, pkg.Name(), version.Name()))
				}
			}
		}
	}
}
