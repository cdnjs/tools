package util

import (
	"fmt"
	"os"
	"path"
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

// GetHumanPackagesPath gets the path to the cdnjs/packages repo.
func GetHumanPackagesPath() string {
	return path.Join(GetBotBasePath(), "packages", "packages")
}

// GetSRIsPath gets the path to the cdnjs/SRIs repo.
func GetSRIsPath() string {
	return path.Join(GetBotBasePath(), "SRIs")
}

// GetCDNJSLibrariesPath gets the path to the cdnjs libraries.
func GetCDNJSLibrariesPath() string {
	return path.Join(GetCDNJSPath(), "ajax", "libs")
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
