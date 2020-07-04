package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"strings"

	cloudflare "github.com/cloudflare/cloudflare-go"

	"github.com/cdnjs/tools/util"
)

var (
	namespaceID = util.GetEnv("WORKERS_KV_NAMESPACE_ID")
	accountID   = util.GetEnv("WORKERS_KV_ACCOUNT_ID")
	apiKey      = util.GetEnv("WORKERS_KV_API_KEY")
	email       = util.GetEnv("WORKERS_KV_EMAIL")
	api         = getAPI()
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

func delete(key string) {
	fmt.Printf("deleting %s\n", key)
	resp, err := api.DeleteWorkersKV(context.Background(), namespaceID, key)
	util.Check(err)
	if !resp.Success {
		log.Fatalf("delete failure %v\n", resp)
	}
}

func deleteTestEntries(startsWith string) {
	kvs := getKVs()
	for _, res := range kvs.Result {
		if key := res.Name; strings.HasPrefix(key, startsWith) {
			delete(key)
		}
	}
}

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

func main() {
	r, err := api.ListWorkersKVs(context.Background(), namespaceID)
	util.Check(err)
	fmt.Println(r)
}
