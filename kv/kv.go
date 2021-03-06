package kv

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	// "sort"
	"strings"

	"github.com/cdnjs/tools/sentry"
	"github.com/cdnjs/tools/util"
	cloudflare "github.com/cloudflare/cloudflare-go"
)

const (
	// workaround for now since cloudflare's API does not currently
	// return a cloudflare.Response object for api.ReadWorkersKV
	keyNotFound    = "key not found"
	authError      = "Authentication error"
	serviceFailure = "service failure"
)

var (
	versionsNamespaceID           = os.Getenv("WORKERS_KV_VERSIONS_NAMESPACE_ID")
	packagesNamespaceID           = os.Getenv("WORKERS_KV_PACKAGES_NAMESPACE_ID")
	aggregatedMetadataNamespaceID = os.Getenv("WORKERS_KV_AGGREGATED_METADATA_NAMESPACE_ID")
)

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

// Ensure a response is successful and the error is nil.
func checkSuccess(r cloudflare.Response, err error) error {
	if err != nil {
		return err
	}
	if !r.Success {
		return fmt.Errorf("kv fail: %v", r)
	}
	return nil
}

// read reads an entry from Workers KV.
func read(api *cloudflare.API, key string, namespaceID string) ([]byte, error) {
	var bytes []byte
	var err error
	for i := 0; i < util.MaxKVAttempts; i++ {
		bytes, err = api.ReadWorkersKV(context.Background(), namespaceID, key)
		if err != nil {
			errString := err.Error()

			// check for service failure and retry
			if strings.Contains(errString, serviceFailure) {
				continue
			}

			// check for key not found
			if strings.Contains(errString, keyNotFound) {
				return nil, KeyNotFoundError{key, errString}
			}

			// check for authentication error
			if strings.Contains(errString, authError) {
				return nil, AuthError{errString}
			}
		}

		break
	}

	return bytes, err
}

// Encodes a byte array to a base64 string.
func encodeToBase64(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

// Encodes key-value pairs to base64 and writes them to KV in multiple bulk requests.
// Returns the list of human-readable names of successful writes.
func EncodeAndWriteKVBulk(ctx context.Context, cfapi *cloudflare.API,
	kvs []WriteRequest, namespaceID string, panicOversized bool) ([]string, error) {
	var bulkWrites []cloudflare.WorkersKVBulkWriteRequest
	var bulkWrite []*cloudflare.WorkersKVPair
	var successfulWrites []string
	var totalSize, totalKeys int64

	for _, kv := range kvs {
		if unencodedSize := int64(len(kv.GetValue())); unencodedSize > util.MaxFileSize {
			log.Printf("ignoring oversized file: %s (%d)\n", kv.GetKey(), unencodedSize)
			sentry.NotifyError(fmt.Errorf("ignoring oversized file: %s (%d)", kv.GetKey(), unencodedSize))
			if panicOversized {
				panic(fmt.Sprintf("oversized file: %s (%d)", kv.GetKey(), unencodedSize))
			}
			continue
		}
		// Note that after encoding in base64 the size may get larger, but after decoding
		// it will be reduced, so it is okay if the size is larger than util.MaxFileSize after encoding base64.
		// However, we still need to check for the KV bulk request limit of 100MiB.
		encodedValue := encodeToBase64(kv.GetValue())
		size := int64(len(encodedValue))
		writePair := &cloudflare.WorkersKVPair{
			Key:    kv.GetKey(),
			Value:  encodedValue,
			Base64: true,
		}
		if kv.GetMeta() != nil {
			// Marshal metadata into JSON bytes.
			bytes, err := json.Marshal(kv.GetMeta())
			if err != nil {
				return nil, err
			}
			metasize := int64(len(bytes))
			if metasize > util.MaxMetadataSize {
				log.Printf("ignoring oversized metadata: %s (%d)\n", kv.GetKey(), metasize)
				sentry.NotifyError(fmt.Errorf("oversized metadata: %s (%d) - %s", kv.GetKey(), metasize, bytes))
				if panicOversized {
					panic(fmt.Sprintf("oversized metadata: %s (%d)", kv.GetKey(), metasize))
				}
				continue
			}
			writePair.Metadata = kv.GetMeta()
			size += metasize
		}
		if totalSize+size > util.MaxBulkWritePayload || totalKeys == util.MaxBulkKeys {
			// Create a new bulk since we are over a limit.
			bulkWrites = append(bulkWrites, bulkWrite)
			bulkWrite = []*cloudflare.WorkersKVPair{}
			totalSize = 0
			totalKeys = 0
		}
		bulkWrite = append(bulkWrite, writePair)
		successfulWrites = append(successfulWrites, kv.GetName())
		totalSize += size
		totalKeys++

		kv.Consumed()
	}
	bulkWrites = append(bulkWrites, bulkWrite)

	for i, b := range bulkWrites {
		log.Printf("writing bulk %d/%d (keys=%d)...\n", i+1, len(bulkWrites), len(b))
		for j := 0; j < util.MaxKVAttempts; j++ {
			r, err := cfapi.WriteWorkersKVBulk(ctx, namespaceID, b)

			// check for service failure and retry
			if err != nil && strings.Contains(err.Error(), serviceFailure) {
				if j == util.MaxKVAttempts-1 {
					return nil, err // no more attempts
				}
				continue // retry
			}

			if err = checkSuccess(r, err); err != nil {
				return nil, err
			}

			break
		}
	}

	return successfulWrites, nil
}

// Returns all KVs that start with a prefix.
func listByPrefix(api *cloudflare.API, prefix, namespaceID string) ([]cloudflare.StorageKey, error) {
	var cursor *string
	var results []cloudflare.StorageKey
	for {
		o := cloudflare.ListWorkersKVsOptions{
			Prefix: &prefix,
			Cursor: cursor,
		}

		resp, err := api.ListWorkersKVsWithOptions(context.Background(), namespaceID, o)
		if err != nil {
			return nil, err
		}

		results = append(results, resp.Result...)

		if resp.Cursor == "" {
			return results, nil
		}

		cursor = &resp.Cursor
	}
}

// Lists by prefix and then returns only the names of the results.
func listByPrefixNamesOnly(api *cloudflare.API, prefix string, namespaceID string) ([]string, error) {
	results, err := listByPrefix(api, prefix, namespaceID)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, r := range results {
		names = append(names, r.Name)
	}

	return names, nil
}
