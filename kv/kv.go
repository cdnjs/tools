package kv

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"sort"

	"github.com/blang/semver"
	"github.com/cdnjs/tools/sri"
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

// KV represents a string key and []byte value.
type KV struct {
	Key   string
	Value []byte
}

// Gets a new *cloudflare.API.
func getAPI() *cloudflare.API {
	a, err := cloudflare.New(apiKey, email, cloudflare.UsingAccount(accountID))
	util.Check(err)
	return a
}

func getKVs() cloudflare.ListStorageKeysResponse {
	resp, err := api.ListWorkersKVs(context.Background(), namespaceID)
	util.Check(err)
	return resp
}

func getKVsWithOptions(o cloudflare.ListWorkersKVsOptions) cloudflare.ListStorageKeysResponse {
	resp, err := api.ListWorkersKVsWithOptions(context.Background(), namespaceID, o)
	util.Check(err)
	return resp
}

// func worker(basePath string, paths <-chan string, kvPairs chan<- *cloudflare.WorkersKVPair) {
// 	fmt.Println("worker start!", basePath)
// 	for p := range paths {
// 		bytes, err := ioutil.ReadFile(path.Join(basePath, p))
// 		if err != nil {
// 			panic(err)
// 		}
// 		// resp, err := api.WriteWorkersKV(context.Background(), namespaceID, p, bytes)
// 		// util.Check(err)
// 		// fmt.Println(resp.Success, p)
// 		// kvPairs <- nil
// 		kvPairs <- &cloudflare.WorkersKVPair{
// 			Key:   p,
// 			Value: string(bytes),
// 		}
// 	}
// }

func encodeToBase64(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

func deleteAllEntries() {
	// get all kvs
	resp := getKVs()

	// make []string of keys
	keys := make([]string, len(resp.Result))
	for i, res := range resp.Result {
		keys[i] = res.Name
	}

	// delete keys
	// TODO: change to api.DeleteWorkersKVsBulk after merge is completed
	for _, key := range keys {
		resp, err := api.DeleteWorkersKV(context.Background(), namespaceID, key)
		util.Check(err)
		if !resp.Success {
			log.Fatalf("Delete failure %v\n", resp)
		}
		fmt.Printf("Deleted %s\n", key)
	}
}

func readKV(key string) ([]byte, error) {
	return api.ReadWorkersKV(context.Background(), namespaceID, key)
}

func writeKV(k string, v []byte) {
	r, err := api.WriteWorkersKV(context.Background(), namespaceID, k, v)
	util.Check(err)
	if !r.Success {
		panic(r)
	}
}

func encodeAndWriteKVBulk(kvs []*KV) {
	var bulkWrites []cloudflare.WorkersKVBulkWriteRequest
	var bulkWrite []*cloudflare.WorkersKVPair
	var totalSize int64
	for _, kv := range kvs {
		if size := int64(len(kv.Value)); size > util.MaxFileSize {
			panic(fmt.Sprintf("oversized file: %s (%d)", kv.Key, size))
		}
		// note that after encoding in base64, the size gets larger, but after decoding
		// it will be reduced, so it is okay if the size is larger than util.MaxFileSize after encoding base64,
		// but we need to watch out for the KV request limit of 100MiB
		encoded := encodeToBase64(kv.Value)
		encodedSize := int64(len(encoded))
		if totalSize+encodedSize > util.MaxBulkWritePayload {
			// split into two bulks
			// this cannot happen when i=0, since util.MaxFileSize must be less than maxBulkPayload
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
		// fmt.Printf("Writing bulk %d (size=%d): %v\n", i, len(b), b)
		r, err := api.WriteWorkersKVBulk(context.Background(), namespaceID, b)
		util.Check(err)
		if !r.Success {
			panic(r)
		}
	}
}

// perform binary search, if not present, add it in the correct index
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

func updateRoot(pkg string) *KV {
	var r Root
	key := rootKey
	if bytes, err := readKV(key); err != nil {
		// assume key is not found (could also be auth error)
		r.Packages = []string{pkg}
	} else {
		util.Check(json.Unmarshal(bytes, &r))
		r.Packages = insertToSortedListIfNotPresent(r.Packages, pkg)
	}

	v, err := json.Marshal(r)
	util.Check(err)

	return &KV{
		Key:   key,
		Value: v,
	}
}

func updatePackage(pkg, version string) *KV {
	var p Package
	key := pkg
	if bytes, err := readKV(key); err != nil {
		// assume key is not found (could also be auth error)
		p.Versions = []string{version}
	} else {
		util.Check(json.Unmarshal(bytes, &p))
		p.Versions = insertToSortedListIfNotPresent(p.Versions, version)
	}

	v, err := json.Marshal(p)
	util.Check(err)

	return &KV{
		Key:   key,
		Value: v,
	}
}

func updateVersion(pkg, version string, files []File) *KV {
	key := path.Join(pkg, version)

	v, err := json.Marshal(Version{Files: files})
	util.Check(err)

	return &KV{
		Key:   key,
		Value: v,
	}
}

func updateFiles(pkg, version, fullPathToVersion string, fromVersionPaths []string) ([]*KV, []File) {
	baseKeyPath := path.Join(pkg, version)
	kvs := make([]*KV, len(fromVersionPaths))
	files := make([]File, len(fromVersionPaths))

	for i, fromVersionPath := range fromVersionPaths {
		fullPath := path.Join(fullPathToVersion, fromVersionPath)
		bytes, err := ioutil.ReadFile(fullPath)
		util.Check(err)

		kvs[i] = &KV{
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
// Will want to push to a queue or write to disk somewhere
// when an operation is about to be attempted and when an
// operation completes successfully. This is to help recover from
// silent failures that result in inconsistent states.
func pushToTaskQueue(_ string) {
}

func updateKV(pkg, version, fullPathToVersion string, fromVersionPaths []string) {
	var kvs []*KV
	pairs, files := updateFiles(pkg, version, fullPathToVersion, fromVersionPaths)
	kvs = append(kvs, pairs...)
	kvs = append(kvs, updateVersion(pkg, version, files))
	kvs = append(kvs, updatePackage(pkg, version))
	kvs = append(kvs, updateRoot(pkg))

	pushToTaskQueue("TODO -- WRITING TO KV")
	encodeAndWriteKVBulk(kvs)
	pushToTaskQueue("TODO -- WROTE TO KV SUCCESSFULLY")
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

	basePath := util.GetCDNJSPackages()

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
