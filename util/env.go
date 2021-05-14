package util

import (
	"fmt"
	"os"
)

// GetEnv gets an environment variable, panicking if it is nonexistent.
func GetEnv(name string) string {
	if v, ok := os.LookupEnv(name); ok {
		return v
	}
	panic(fmt.Sprintf("Env %s is missing\n", name))
}

// EnvExists determines if an environment variable exists.
func EnvExists(name string) bool {
	_, ok := os.LookupEnv(name)
	return ok
}

// IsDebug returns true if debug mode is enabled based
// on an environment variable.
func IsDebug() bool {
	return EnvExists("DEBUG")
}

// HasHTTPProxy returns true if the http proxy environment
// variable is set.
func HasHTTPProxy() bool {
	return EnvExists("HTTP_PROXY")
}

// GetProtocol gets the protocol, either http or https.
func GetProtocol() string {
	if HasHTTPProxy() {
		return "http"
	}
	return "https"
}
