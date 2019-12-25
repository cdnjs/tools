package util

import (
	"os"
	"path/filepath"
	"strings"

	ohmyglob "github.com/pachyderm/ohmyglob"
)

func ListFilesGlob(base string, pattern string) []string {
	list := make([]string, 0)
	glob, err := ohmyglob.Compile(pattern)
	Check(err)

	Check(filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return filepath.SkipDir
		}

		relativePath := strings.ReplaceAll(path, base+"/", "")

		if !info.IsDir() && glob.Match(relativePath) {
			list = append(list, relativePath)
		}

		return nil
	}))

	return list
}
