package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	cloudflare "github.com/cloudflare/cloudflare-go"

	"github.com/cdnjs/tools/util"
)

var (
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

// func delete(key string) {
// 	fmt.Printf("deleting %s\n", key)
// 	resp, err := api.DeleteWorkersKV(context.Background(), namespaceID, key)
// 	util.Check(err)
// 	if !resp.Success {
// 		log.Fatalf("delete failure %v\n", resp)
// 	}
// }

// func deleteTestEntries(startsWith string) {
// 	kvs := getKVs()
// 	for _, res := range kvs.Result {
// 		if key := res.Name; strings.HasPrefix(key, startsWith) {
// 			delete(key)
// 		}
// 	}
// }

func worker(basePath string, paths <-chan string, kvPairs chan<- *cloudflare.WorkersKVPair) {
	fmt.Println("worker start!", basePath)
	for p := range paths {
		bytes, err := ioutil.ReadFile(path.Join(basePath, p))
		if err != nil {
			panic(err)
		}
		// resp, err := api.WriteWorkersKV(context.Background(), namespaceID, p, bytes)
		// util.Check(err)
		// fmt.Println(resp.Success, p)
		// kvPairs <- nil
		kvPairs <- &cloudflare.WorkersKVPair{
			Key:   p,
			Value: string(bytes),
		}
	}
}

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

// type ReadKVErrs struct {
// 	Message string `json:"message"`
// }

func readKV(key string) ([]byte, error) {
	return api.ReadWorkersKV(context.Background(), namespaceID, key)
}

// func writeKVString(k, v string) {
// 	api.WriteWorkersKV(context.Background(), namespaceID, k, []byte(v))
// }

// func writeKV(k string, v []byte) {
// 	r, err := api.WriteWorkersKV(context.Background(), namespaceID, k, v)
// 	util.Check(err)
// 	if !r.Success {
// 		panic(r)
// 	}
// }

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
// list files in version
// maybe keep sris here in the future
type Version struct {
	Files []string `json:"files"`
}

func updateRoot(pkg string) *cloudflare.WorkersKVPair {
	var r Root
	key := rootKey
	if bytes, err := readKV(key); err != nil {
		// assume key is not found (could also be auth error)
		r.Packages = []string{pkg}
	} else {
		util.Check(json.Unmarshal(bytes, &r))
		r.Packages = append(r.Packages, pkg)
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
		p.Versions = append(p.Versions, version)
	}

	v, err := json.Marshal(p)
	util.Check(err)

	return &cloudflare.WorkersKVPair{
		Key:   key,
		Value: string(v),
	}
}

func updateVersion(pkg, version string, fromVersionPaths []string) *cloudflare.WorkersKVPair {
	key := path.Join(pkg, version)

	v, err := json.Marshal(Version{Files: fromVersionPaths})
	util.Check(err)

	return &cloudflare.WorkersKVPair{
		Key:   key,
		Value: string(v),
	}
}

func updateFiles(pkg, version, fullPathToVersion string, fromVersionPaths []string) []*cloudflare.WorkersKVPair {
	baseKeyPath := path.Join(pkg, version)
	kvs := make([]*cloudflare.WorkersKVPair, len(fromVersionPaths))

	for i, fromVersionPath := range fromVersionPaths {
		bytes, err := ioutil.ReadFile(path.Join(fullPathToVersion, fromVersionPath))
		util.Check(err)

		kvs[i] = &cloudflare.WorkersKVPair{
			Key:    path.Join(baseKeyPath, fromVersionPath),
			Value:  encodeToBase64(bytes),
			Base64: true,
		}
	}

	return kvs
}

func updateKV(pkg, version, fullPathToVersion string, fromVersionPaths []string) {
	// ensure not over limit
	// make sure limit actually is 100
	var kvs []*cloudflare.WorkersKVPair
	kvs = append(kvs, updateRoot(pkg))
	kvs = append(kvs, updatePackage(pkg, version))
	kvs = append(kvs, updateVersion(pkg, version, fromVersionPaths))
	kvs = append(kvs, updateFiles(pkg, version, fullPathToVersion, fromVersionPaths)...)
	// fmt.Println(kvs)
	writeKVBulk(kvs)
}

// bot finds new version
// downloads to a path
//
// then inserts directly to kv, does not move in disk
// move from temp dir directly to kv
// then remove temp dir
//
// calculates sri, also puts into kv
// puts package.json metadata into kv as well

func insertVersionToKV(pkg, version, fullPathToVersion string) {
	fromVersionPaths, err := util.ListFilesInVersion(context.Background(), fullPathToVersion)
	util.Check(err)
	updateKV(pkg, version, fullPathToVersion, fromVersionPaths)
}

func main() {
	deleteAllEntries()

	insertVersionToKV("1000hz-bootstrap-validator", "0.10.0", "/Users/tylercaslin/go/src/fake-smaller-repo/cdnjs/ajax/libs/1000hz-bootstrap-validator/0.10.0")

	os.Exit(1)
}
