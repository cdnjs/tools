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

// GetEnv gets an environment variable, panicking if it is empty.
func GetEnv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		panic(fmt.Sprintf("Env %s is missing\n", name))
	}
	return v
}

// IsDebug returns true if debug mode is enabled based
// on an environment variable.
func IsDebug() bool {
	return os.Getenv("DEBUG") != ""
}

// GetBotBasePath gets the bot base path from an environment variable.
func GetBotBasePath() string {
	return GetEnv("BOT_BASE_PATH")
}

// GetCDNJSPackages gets the path to the cdnjs libraries.
func GetCDNJSPackages() string {
	return path.Join(GetBotBasePath(), "cdnjs", "ajax", "libs")
}

// HasHTTPProxy returns true if the http proxy environment
// variable is set.
func HasHTTPProxy() bool {
	return os.Getenv("HTTP_PROXY") != ""
}
