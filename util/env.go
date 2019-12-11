package util

import (
	"fmt"
	"os"
)

func GetEnv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		panic(fmt.Sprintf("Env %s is missing\n", name))
	}
	return v
}
