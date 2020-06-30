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
)
