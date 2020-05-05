package util

const (
	// When no versions exist in cdnjs and we are trying to import all of them,
	// limit it to the a few last versions to avoid publishing too many outdated
	// versions
	IMPORT_ALL_MAX_VERSIONS = 10
)
