package util

import (
	"context"
)

// ContextKey is the key type used for context.WithValue().
type ContextKey int

// ContextEntry represents a key-value entry for a context.
type ContextEntry struct {
	Key   ContextKey
	Value interface{}
}

const (
	// LoggerPrefix is the key to the string that is outputted first when logging.
	// If the *log.Logger itself has a prefix set as well, the *log.Logger's
	// prefix will be outputted before the LoggerPrefix.
	//
	// For example, the LoggerPrefix may represent a file name and
	// the *log.Logger may have a prefix to represent the program entry point.
	// As a result, when logging "hello world" from the program "main" and path "/usr",
	// the output may look like "main /usr hello world".
	LoggerPrefix ContextKey = iota

	// Logger is the key for a *log.Logger.
	Logger

	// Debug is the LogFunc that is called when outputting a debug statement.
	Debug

	// Warn is the LogFunc that is called when outputting a warning.
	Warn

	// Err is the LogFunc that is called when outputting an error.
	Err

	// Info is the LogFunc that is called when outputting an info.
	Info
)

// ContextWithEntries creates a context with a variadic number of key-value
// entries. Internally, this context's root node is context.Background() and
// a new context is created for each new key-value entry. While a single value
// as a map may be more efficient, there are only a handful of potential ContextEntry
// entries, so the complexity can be ignored.
func ContextWithEntries(entries ...ContextEntry) context.Context {
	parent := context.Background()
	for _, entry := range entries {
		parent = context.WithValue(parent, entry.Key, entry.Value)
	}
	return parent
}
