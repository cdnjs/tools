package util

import (
	"fmt"
	"os"
)

const (
	SRI_PATH = "../SRIs"
)

func GetEnv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		panic(fmt.Sprintf("Env %s is missing\n", name))
	}
	return v
}

func IsDebug() bool {
	return os.Getenv("DEBUG") != ""
}
