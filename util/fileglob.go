package util

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/karrick/godirwalk"
)

// ListFilesGlob is the legacy, slower version that uses
// the node glob tool found here: https://github.com/cdnjs/glob
func ListFilesGlob(ctx context.Context, base string, pattern string) ([]string, error) {
	list := make([]string, 0)

	// check if the version is hidden
	if isHidden(base) {
		Debugf(ctx, "ignoring hidden version %s", base)
		return list, nil
	}

	if _, err := os.Stat(base); os.IsNotExist(err) {
		Debugf(ctx, "match %s in %s but doesn't exists", pattern, base)
		return list, nil
	}

	cmd := exec.Command(path.Join(GetBotBasePath(), "glob", "index.js"), pattern)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	cmd.Dir = base
	err := cmd.Run()
	if err != nil {
		fmt.Printf("%s: %s\n", err, out.String())
		return list, err
	}

	for _, line := range strings.Split(out.String(), "\n") {
		if strings.Trim(line, " ") != "" {
			list = append(list, line)
		}

	}
	return list, nil
}

// Determines if a file path contains a hidden file or directory.
// Either it starts with . or contains '/.' to be considered hidden.
func isHidden(fp string) bool {
	return strings.HasPrefix(fp, ".") || strings.Contains(fp, "/.")
}

// ListFilesInVersion is the new, optimized version of ListFilesGlob created for
// package generation. It lists all of the files within a particular cdnjs package version in
// the same manner as ListFilesGlob with a '**' glob pattern.
// Note that hidden cdnjs versions and hidden files/directories are ignored.
// It utilizes the fast godirwalk library found here: https://github.com/karrick/godirwalk
func ListFilesInVersion(ctx context.Context, base string) ([]string, error) {
	list := make([]string, 0)

	// check if the version is hidden
	if isHidden(base) {
		Debugf(ctx, "ignoring hidden version %s", base)
		return list, nil
	}

	// walk the files recursively within the cdnjs package version directory
	err := godirwalk.Walk(base, &godirwalk.Options{
		Callback: func(fp string, de *godirwalk.Dirent) error {
			// trim a full path to a path relative to the base directory (inside package version dir)
			// trim any leading '/' for consistency with legacy ListFilesGlob implementation
			fp = strings.TrimLeft(strings.TrimPrefix(fp, base), "/")
			// path must not be a directory, not be hidden, and not be empty
			if !de.IsDir() && !isHidden(fp) && fp != "" {
				list = append(list, fp)
			}
			return nil
		},
		FollowSymbolicLinks: true,
	})

	return list, err
}
