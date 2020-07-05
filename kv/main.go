package main

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

	cloudflare "github.com/cloudflare/cloudflare-go"

	"github.com/cdnjs/tools/util"
)

var (
	// TODO, update README.md
	namespaceID = util.GetEnv("WORKERS_KV_NAMESPACE_ID")
	accountID   = util.GetEnv("WORKERS_KV_ACCOUNT_ID")
	apiKey      = util.GetEnv("WORKERS_KV_API_KEY")
	email       = util.GetEnv("WORKERS_KV_EMAIL")
	api         = getAPI()
	basePath    = util.GetCDNJSPackages()
	rootKey     = "/"
)

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

func writeKVBulk(kvs cloudflare.WorkersKVBulkWriteRequest) {
	r, err := api.WriteWorkersKVBulk(context.Background(), namespaceID, kvs)
	util.Check(err)
	if !r.Success {
		panic(r)
	}
}

// Root ..
// list of packages
// top level metadata?
type Root struct {
	Packages []string `json:"packages"`
}

// Package ..
// can store other metadata like fields in package.json
type Package struct {
	Versions []string `json:"versions"`
}

// Version ..
//
type Version struct {
	Files []File `json:"files"`
}

// File ...
type File struct {
	Name string `json:"name"`
	SRI  string `json:"sri"`
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

func updateRoot(pkg string) *cloudflare.WorkersKVPair {
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

	return &cloudflare.WorkersKVPair{
		Key:   key,
		Value: string(v),
	}
}

func updatePackage(pkg, version string) *cloudflare.WorkersKVPair {
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

	return &cloudflare.WorkersKVPair{
		Key:   key,
		Value: string(v),
	}
}

func updateVersion(pkg, version string, files []File) *cloudflare.WorkersKVPair {
	key := path.Join(pkg, version)

	v, err := json.Marshal(Version{Files: files})
	util.Check(err)

	return &cloudflare.WorkersKVPair{
		Key:   key,
		Value: string(v),
	}
}

func updateFiles(pkg, version, fullPathToVersion string, fromVersionPaths []string) ([]*cloudflare.WorkersKVPair, []File) {
	baseKeyPath := path.Join(pkg, version)
	kvs := make([]*cloudflare.WorkersKVPair, len(fromVersionPaths))
	files := make([]File, len(fromVersionPaths))

	for i, fromVersionPath := range fromVersionPaths {
		fullPath := path.Join(fullPathToVersion, fromVersionPath)
		bytes, err := ioutil.ReadFile(fullPath)
		util.Check(err)

		kvs[i] = &cloudflare.WorkersKVPair{
			Key:    path.Join(baseKeyPath, fromVersionPath),
			Value:  encodeToBase64(bytes),
			Base64: true,
		}

		files[i] = File{
			Name: fromVersionPath,
			SRI:  sri.CalculateFileSRI(fullPath),
		}
	}

	return kvs, files
}

func updateKV(pkg, version, fullPathToVersion string, fromVersionPaths []string) {
	// ensure not over limit, break into more reqs when > 100
	// make sure limit actually is 100
	var kvs []*cloudflare.WorkersKVPair
	pairs, files := updateFiles(pkg, version, fullPathToVersion, fromVersionPaths)
	kvs = append(kvs, pairs...)
	kvs = append(kvs, updateVersion(pkg, version, files))
	kvs = append(kvs, updatePackage(pkg, version))
	kvs = append(kvs, updateRoot(pkg))

	// fmt.Println(kvs)
	writeKVBulk(kvs)
}

// thoughts:

// bot finds new version
// downloads to a path

// then inserts directly to kv, does not move in disk
// move from temp dir directly to kv
// then remove temp dir

// calculates sri, also puts into kv
// puts package.json metadata into kv as well

// fullpath will be useful if the version is downloaded into a temp directory
// so it is not just path.Join(basePath, pkg, version)
func insertVersionToKV(pkg, version, fullPathToVersion string) {
	fromVersionPaths, err := util.ListFilesInVersion(context.Background(), fullPathToVersion)
	util.Check(err)
	updateKV(pkg, version, fullPathToVersion, fromVersionPaths)
}

func main() {
	deleteAllEntries()

	//insertVersionToKV("1000hz-bootstrap-validator", "0.10.0", "/Users/tylercaslin/go/src/fake-smaller-repo/cdnjs/ajax/libs/1000hz-bootstrap-validator/0.10.0")
	//insertVersionToKV("1000hz-bootstrap-validator", "0.10.0", "/Users/tylercaslin/go/src/fake-smaller-repo/cdnjs/ajax/libs/1000hz-bootstrap-validator/0.10.0")

	basePath := util.GetCDNJSPackages()

	pkgs, err := ioutil.ReadDir(basePath)
	util.Check(err)

	for i, pkg := range pkgs {
		if i > 5 {
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
