package util

import (
	"fmt"
	"os"
	"path"
)

const (
	// SRIPath is the path to the directory where calculated SRIs are stored.
	SRIPath = "../SRIs"
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

// GetBotBasePath gets the bot base path from an environment variable.
func GetBotBasePath() string {
	return GetEnv("BOT_BASE_PATH")
}

// GetCDNJSPath gets the path to the cdnjs repo.
func GetCDNJSPath() string {
	return path.Join(GetBotBasePath(), "cdnjs")
}

// GetPackagesPath gets the path to the packages repo.
func GetPackagesPath() string {
	return path.Join(GetBotBasePath(), "packages", "packages")
}

// GetCDNJSPackages gets the path to the cdnjs libraries.
func GetCDNJSPackages() string {
	return path.Join(GetCDNJSPath(), "ajax", "libs")
}

// HasHTTPProxy returns true if the http proxy environment
// variable is set.
func HasHTTPProxy() bool {
	return EnvExists("HTTP_PROXY")
}

// IsKVDisabled determines if writes to KV are disabled.
func IsKVDisabled() bool {
	return EnvExists("DISABLE_KV")
}
