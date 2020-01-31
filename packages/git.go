package packages

import (
	"context"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/cdnjs/tools/util"
)

// We first list all the versions (and top-level package.json)
// in the package and pass the list to git ls-tree which is
// going to filter out those not in the tree
func GitListPackageVersions(ctx context.Context, basePath string) []string {
	filesOnFs, err := filepath.Glob(path.Join(basePath, "*"))
	util.Check(err)

	filteredFilesOnFs := make([]string, 0)

	// filter out package.json
	for _, file := range filesOnFs {
		if !strings.HasSuffix(file, ".donotoptimizepng") && !strings.HasSuffix(file, "package.json") && strings.Trim(file, " ") != "" {
			filteredFilesOnFs = append(filteredFilesOnFs, file)
		}
	}

	args := []string{
		"ls-tree", "--name-only", "origin/master",
	}
	args = append(args, filteredFilesOnFs...)

	out, err := exec.Command("git", args...).Output()
	util.Check(err)

	outFiles := strings.Split(string(out), "\n")

	filteredOutFiles := make([]string, 0)
	// remove basePath from the output
	for _, v := range outFiles {
		if strings.Trim(v, " ") != "" {
			filteredOutFiles = append(
				filteredOutFiles, strings.ReplaceAll(v, basePath+"/", ""))
		}
	}

	// Debug mode
	diff := arrDiff(filteredFilesOnFs, outFiles)
	if len(diff) > 0 {
		util.Printf(ctx, "found staged version: %+q\n", diff)
	}
	// Debug mode

	return filteredOutFiles
}

func arrDiff(a, b []string) (diff []string) {
	m := make(map[string]bool)

	for _, item := range b {
		m[item] = true
	}

	for _, item := range a {
		if _, ok := m[item]; !ok {
			diff = append(diff, item)
		}
	}
	return
}
