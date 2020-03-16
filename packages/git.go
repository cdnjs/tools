package packages

import (
	"context"
	"fmt"
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
		if !strings.HasSuffix(file, ".do_not_update") && !strings.HasSuffix(file, ".donotoptimizepng") && !strings.HasSuffix(file, "package.json") && strings.Trim(file, " ") != "" {
			filteredFilesOnFs = append(filteredFilesOnFs, file)
		}
	}

	// no local version, no need to check what's in git
	if len(filteredFilesOnFs) == 0 {
		return make([]string, 0)
	}

	args := []string{
		"ls-tree", "--name-only", "origin/master",
	}
	args = append(args, filteredFilesOnFs...)

	cmd := exec.Command("git", args...)
	cmd.Dir = basePath
	util.Debugf(ctx, "run %s from %s\n", cmd, basePath)
	out := checkCmd(cmd.CombinedOutput())

	outFiles := strings.Split(out, "\n")

	filteredOutFiles := make([]string, 0)
	// remove basePath from the output
	for _, v := range outFiles {
		if strings.Trim(v, " ") != "" {
			filteredOutFiles = append(
				filteredOutFiles, strings.ReplaceAll(v, basePath+"/", ""))
		}
	}

	if util.IsDebug() {
		diff := arrDiff(filteredFilesOnFs, outFiles)
		if len(diff) > 0 {
			util.Printf(ctx, "found staged version: %+q\n", diff)
		}
	}

	return filteredOutFiles
}

func GitAdd(ctx context.Context, gitpath string, relpath string) {
	args := []string{"add", relpath}

	cmd := exec.Command("git", args...)
	cmd.Dir = gitpath
	util.Debugf(ctx, "run %s\n", cmd)
	checkCmd(cmd.CombinedOutput())
}

func GitCommit(ctx context.Context, gitpath string, msg string) {
	args := []string{"commit", "-m", msg}

	cmd := exec.Command("git", args...)
	cmd.Dir = gitpath
	util.Debugf(ctx, "run %s\n", cmd)
	checkCmd(cmd.CombinedOutput())
}

func GitPush(ctx context.Context, gitpath string) {
	args := []string{"push"}

	cmd := exec.Command("git", args...)
	cmd.Dir = gitpath
	util.Debugf(ctx, "run %s\n", cmd)
	checkCmd(cmd.CombinedOutput())
}

func checkCmd(out []byte, err error) string {
	if err != nil {
		fmt.Println(string(out))
	}
	util.Check(err)
	return string(out)
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
