package util

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
)

func ListFilesGlob(ctx context.Context, base string, pattern string) []string {
	if _, err := os.Stat(base); os.IsNotExist(err) {
		Debugf(ctx, "match %s in %s but doesn't exists", pattern, base)
		return []string{}
	}

	// fmt.Println("match", pattern, "in", base)

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
