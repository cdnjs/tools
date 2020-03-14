package util

import (
	"bytes"
	"fmt"
	"os/exec"
	"path"
	"strings"
)

var (
	GLOB_EXECUTABLE = path.Join(GetEnv("BOT_BASE_PATH"), "glob", "index.js")
)

func ListFilesGlob(base string, pattern string) []string {
	fmt.Println("match", pattern, "in", base)

	cmd := exec.Command(GLOB_EXECUTABLE, pattern)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	cmd.Dir = base
	err := cmd.Run()
	if err != nil {
		fmt.Printf("%s: %s\n", err, out.String())
		Check(err)
	}

	return strings.Split(out.String(), "\n")
}
