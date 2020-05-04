package util

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
)

func ListFilesGlob(base string, pattern string) []string {
	if _, err := os.Stat(base); os.IsNotExist(err) {
		fmt.Println("match", pattern, "in", base, "but doesn't exists")
		return []string{}
	}

	fmt.Println("match", pattern, "in", base)

	cmd := exec.Command(path.Join(GetBotBasePath(), "glob", "index.js"), pattern)
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
