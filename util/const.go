package util

const (
	// ImportAllMaxVersions is the maximum number of versions we will import.
	// When no versions exist in cdnjs and we are trying to import all of them,
	// limit it to the a few last versions to avoid publishing too many outdated
	// versions.
	ImportAllMaxVersions = 10

	// MaxFileSize is the file size in bytes accepted by cdnjs (10MiB).
	MaxFileSize int64 = 10485760

	// MinNpmMonthlyDownloads is the minimum number of monthly downloads
	// from npm needed for a library to be accepted into cdnjs.
	MinNpmMonthlyDownloads = 800

	// MinGitHubStars is the minimum number of stars for a library to be
	// accepted into cdnjs.
	MinGitHubStars = 200

	// MaxBulkWritePayload is the maximum payload size for a request to write
	// to Workers KV in bulk. Note that this will be the sum of all values only.
	// The max bulk request size is 100MiB (104857600), so this will provide
	// enough leeway for any metadata stored with each key (up to 1024 bytes),
	// long keys, and verbose JSON syntax.
	MaxBulkWritePayload int64 = 1e8

	// MaxMetadataSize is the maximum metadata in bytes that can be stored for a
	// particular KV entry.
	MaxMetadataSize int64 = 1024

	// MaxBulkKeys is the maximum number of keys that can be pushed to KV in one bulk request.
	MaxBulkKeys int64 = 1e4

	// MaxKVAttempts is the maximum number of attempts to perform a KV read/write
	// if the error returned is a 502 service failure.
	MaxKVAttempts = 3
)
