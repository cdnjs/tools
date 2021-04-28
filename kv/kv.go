package kv

import (
	// "context"
	"encoding/base64"
	// "encoding/json"
	"fmt"
	// "sort"
	// "strings"

	// "github.com/cdnjs/tools/sentry"
	// "github.com/cdnjs/tools/util"
	cloudflare "github.com/cloudflare/cloudflare-go"
)

const (
	// workaround for now since cloudflare's API does not currently
	// return a cloudflare.Response object for api.ReadWorkersKV
	keyNotFound    = "key not found"
	authError      = "Authentication error"
	serviceFailure = "service failure"
)

// var (
// 	srisNamespaceID               = util.GetEnv("WORKERS_KV_SRIS_NAMESPACE_ID")
// 	filesNamespaceID              = util.GetEnv("WORKERS_KV_FILES_NAMESPACE_ID")
// 	versionsNamespaceID           = util.GetEnv("WORKERS_KV_VERSIONS_NAMESPACE_ID")
// 	packagesNamespaceID           = util.GetEnv("WORKERS_KV_PACKAGES_NAMESPACE_ID")
// 	aggregatedMetadataNamespaceID = util.GetEnv("WORKERS_KV_AGGREGATED_METADATA_NAMESPACE_ID")
// 	accountID                     = util.GetEnv("WORKERS_KV_ACCOUNT_ID")
// 	apiToken                      = util.GetEnv("WORKERS_KV_API_TOKEN")
// 	api                           = getAPI()
// )

// KeyNotFoundError represents a KV key not found.
type KeyNotFoundError struct {
	key string
	err string
}

// Error is used to satisfy the error interface.
func (k KeyNotFoundError) Error() string {
	return fmt.Sprintf("%s (%s): %s", keyNotFound, k.key, k.err)
}

// AuthError represents an authentication error.
type AuthError struct {
	err string
}

// Error is used to satisfy the error interface.
func (a AuthError) Error() string {
	return fmt.Sprintf("%s: %s", authError, a.err)
}

// Gets a new *cloudflare.API.
// func getAPI() *cloudflare.API {
// 	a, err := cloudflare.NewWithAPIToken(apiToken, cloudflare.UsingAccount(accountID))
// 	util.Check(err)
// 	return a
// }

// read reads an entry from Workers KV.
// func read(key, namespaceID string) ([]byte, error) {
// 	var bytes []byte
// 	var err error
// 	for i := 0; i < util.MaxKVAttempts; i++ {
// 		bytes, err = api.ReadWorkersKV(context.Background(), namespaceID, key)
// 		if err != nil {
// 			errString := err.Error()

// 			// check for service failure and retry
// 			if strings.Contains(errString, serviceFailure) {
// 				continue
// 			}

// 			// check for key not found
// 			if strings.Contains(errString, keyNotFound) {
// 				return nil, KeyNotFoundError{key, errString}
// 			}

// 			// check for authentication error
// 			if strings.Contains(errString, authError) {
// 				return nil, AuthError{errString}
// 			}
// 		}

// 		break
// 	}

// 	return bytes, err
// }

// Returns all KVs that start with a prefix.
// func listByPrefix(prefix, namespaceID string) ([]cloudflare.StorageKey, error) {
// 	var cursor *string
// 	var results []cloudflare.StorageKey
// 	for {
// 		o := cloudflare.ListWorkersKVsOptions{
// 			Prefix: &prefix,
// 			Cursor: cursor,
// 		}

// 		resp, err := api.ListWorkersKVsWithOptions(context.Background(), namespaceID, o)
// 		if err != nil {
// 			return nil, err
// 		}

// 		results = append(results, resp.Result...)

// 		if resp.Cursor == "" {
// 			return results, nil
// 		}

// 		cursor = &resp.Cursor
// 	}
// }

// Lists by prefix and then returns only the names of the results.
// func listByPrefixNamesOnly(prefix, namespaceID string) ([]string, error) {
// 	results, err := listByPrefix(prefix, namespaceID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var names []string
// 	for _, r := range results {
// 		names = append(names, r.Name)
// 	}

// 	return names, nil
// }

// Encodes a byte array to a base64 string.
func encodeToBase64(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

// InsertNewVersionToKV inserts a new version to KV and returns the uploaded version files as JSON.
// The `fullPathToVersion` string will be useful if the version is downloaded to
// a temporary directory, not necessarily always in `$BOT_BASE_PATH/cdnjs/ajax/libs/`.
//
// Note that this function will also compress the files, generating brotli/gzip entries
// to KV where necessary.
//
// Note this function will NOT update package metadata.
//
// For example:
// InsertNewVersionToKV("1000hz-bootstrap-validator", "0.10.0", "/tmp/1000hz-bootstrap-validator/0.10.0")
// func InsertNewVersionToKV(ctx context.Context, pkg, version, fullPathToVersion string, metaOnly, srisOnly, filesOnly, noPush, panicOversized bool) ([]string, []byte, []string, []string, int, int, error) {
// 	fromVersionPaths, err := util.ListFilesInVersion(ctx, fullPathToVersion)
// 	if err != nil {
// 		return nil, nil, nil, nil, 0, 0, err
// 	}
// 	sort.Strings(fromVersionPaths)

// 	var versionBytes []byte
// 	if !filesOnly && !srisOnly && !noPush {
// 		// write version metadata to KV
// 		versionBytes, err = updateKVVersion(ctx, pkg, version, fromVersionPaths)
// 		if err != nil {
// 			return nil, nil, nil, nil, 0, 0, err
// 		}
// 		if metaOnly {
// 			return fromVersionPaths, versionBytes, nil, nil, 0, 0, nil
// 		}
// 	}

// 	// write files to KV
// 	srisPushedToKV, filesPushedToKV, theoreticalSRIKeys, theoreticalFileKeys, err := updateKVFiles(ctx, pkg, version, fullPathToVersion, fromVersionPaths, srisOnly, filesOnly, noPush, panicOversized)
// 	return fromVersionPaths, versionBytes, srisPushedToKV, filesPushedToKV, theoreticalSRIKeys, theoreticalFileKeys, err
// }
