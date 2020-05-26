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
	out := util.CheckCmd(cmd.CombinedOutput())

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
			util.Printf(ctx, "found %d staged versions\n", len(diff))
		}
	}

	return filteredOutFiles
}

func GitAdd(ctx context.Context, gitpath string, relpath string) {
	args := []string{"add", relpath}

	cmd := exec.Command("git", args...)
	cmd.Dir = gitpath
	util.Debugf(ctx, "run %s\n", cmd)
	util.CheckCmd(cmd.CombinedOutput())
}

func GitCommit(ctx context.Context, gitpath string, msg string) {
	args := []string{"commit", "-m", msg}

	cmd := exec.Command("git", args...)
	cmd.Dir = gitpath
	util.Debugf(ctx, "run %s\n", cmd)
	util.CheckCmd(cmd.CombinedOutput())
}

func GitFetch(ctx context.Context, gitpath string) {
	args := []string{"fetch"}

	cmd := exec.Command("git", args...)
	cmd.Dir = gitpath
	util.Debugf(ctx, "run %s\n", cmd)
	util.CheckCmd(cmd.CombinedOutput())
}

func GitPush(ctx context.Context, gitpath string) {
	args := []string{"push"}

	cmd := exec.Command("git", args...)
	cmd.Dir = gitpath
	util.Debugf(ctx, "run %s\n", cmd)
	util.CheckCmd(cmd.CombinedOutput())
}

func GitClone(ctx context.Context, pckg *Package, gitpath string) ([]byte, error) {
	args := []string{"clone", pckg.Autoupdate.Target, "."}

	cmd := exec.Command("git", args...)
	cmd.Dir = gitpath
	util.Debugf(ctx, "%s: run %s\n", gitpath, cmd)
	out, err := cmd.CombinedOutput()
	return out, err
}

func GitTags(ctx context.Context, pckg *Package, gitpath string) []string {
	args := []string{"tag"}

	cmd := exec.Command("git", args...)
	cmd.Dir = gitpath
	util.Debugf(ctx, "run %s\n", cmd)
	out := util.CheckCmd(cmd.CombinedOutput())
	return strings.Split(out, "\n")
}

func GitCheckout(ctx context.Context, pckg *Package, gitpath string, tag string) {
	args := []string{"checkout", tag}

	cmd := exec.Command("git", args...)
	cmd.Dir = gitpath
	util.Debugf(ctx, "run %s\n", cmd)
	util.CheckCmd(cmd.CombinedOutput())
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
