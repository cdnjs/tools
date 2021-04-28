package kv_pump

import (
	"context"
	"encoding/json"
	"fmt"
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

type FileMetadata struct {
	ETag         string `json:"etag,omitempty"`
	LastModified string `json:"last_modified,omitempty"`
	SRI          string `json:"sri,omitempty"`
}

// Represents a KV write request, consisting of
// a string key, a []byte value, and file metadata.
// The name field is used to identify this write request
// with a human-readable friendly name.
type writeRequest struct {
	key   string
	name  string
	value []byte
	meta  *FileMetadata
}

// Encodes key-value pairs to base64 and writes them to KV in multiple bulk requests.
// Returns the list of human-readable names of successful writes.
func encodeAndWriteKVBulk(ctx context.Context, cfapi *cloudflare.API, kvs []*writeRequest, namespaceID string, panicOversized bool) ([]string, error) {
	var bulkWrites []cloudflare.WorkersKVBulkWriteRequest
	var bulkWrite []*cloudflare.WorkersKVPair
	var successfulWrites []string
	var totalSize, totalKeys int64

	for _, kv := range kvs {
		if unencodedSize := int64(len(kv.value)); unencodedSize > util.MaxFileSize {
			// util.Printf(ctx, "ignoring oversized file: %s (%d)\n", kv.key, unencodedSize)
			sentry.NotifyError(fmt.Errorf("ignoring oversized file: %s (%d)", kv.key, unencodedSize))
			if panicOversized {
				panic(fmt.Sprintf("oversized file: %s (%d)", kv.key, unencodedSize))
			}
			continue
		}
		// Note that after encoding in base64 the size may get larger, but after decoding
		// it will be reduced, so it is okay if the size is larger than util.MaxFileSize after encoding base64.
		// However, we still need to check for the KV bulk request limit of 100MiB.
		encodedValue := encodeToBase64(kv.value)
		size := int64(len(encodedValue))
		writePair := &cloudflare.WorkersKVPair{
			Key:    kv.key,
			Value:  encodedValue,
			Base64: true,
		}
		if kv.meta != nil {
			// Marshal metadata into JSON bytes.
			bytes, err := json.Marshal(kv.meta)
			if err != nil {
				return nil, err
			}
			metasize := int64(len(bytes))
			if metasize > util.MaxMetadataSize {
				util.Printf(ctx, "ignoring oversized metadata: %s (%d)\n", kv.key, metasize)
				sentry.NotifyError(fmt.Errorf("oversized metadata: %s (%d) - %s", kv.key, metasize, bytes))
				if panicOversized {
					panic(fmt.Sprintf("oversized metadata: %s (%d)", kv.key, metasize))
				}
				continue
			}
			util.Printf(ctx, "writing metadata: %s\n", bytes)
			writePair.Metadata = kv.meta
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
		successfulWrites = append(successfulWrites, kv.name)
		totalSize += size
		totalKeys++
	}
	bulkWrites = append(bulkWrites, bulkWrite)

	for i, b := range bulkWrites {
		util.Printf(ctx, "writing bulk %d/%d (keys=%d)...\n", i+1, len(bulkWrites), len(b))
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
