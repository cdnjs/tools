package packages

import (
	"bytes"
	"context"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/cdnjs/tools/util"
)

// GitListPackageVersions first lists all the versions (and top-level package.json)
// in the package and passes the list to git ls-tree which filters out
// those not in the tree.
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

// GitAdd adds to the next commit.
func GitAdd(ctx context.Context, gitpath, relpath string) {
	args := []string{"add", relpath}

	cmd := exec.Command("git", args...)
	cmd.Dir = gitpath
	util.Debugf(ctx, "run %s\n", cmd)
	util.CheckCmd(cmd.CombinedOutput())
}

// GitCommit makes a new commit.
func GitCommit(ctx context.Context, gitpath, msg string) {
	args := []string{"commit", "-m", msg}

	cmd := exec.Command("git", args...)
	cmd.Dir = gitpath
	util.Debugf(ctx, "run %s\n", cmd)
	util.CheckCmd(cmd.CombinedOutput())
}

// GitFetch fetches objs/refs to the repository.
func GitFetch(ctx context.Context, gitpath string) ([]byte, error) {
	args := []string{"fetch"}

	cmd := exec.Command("git", args...)
	cmd.Dir = gitpath
	util.Debugf(ctx, "%s: run %s\n", gitpath, cmd)
	return cmd.CombinedOutput()
}

// GitPush pushes to a git repository.
func GitPush(ctx context.Context, gitpath string) {
	args := []string{"push"}

	cmd := exec.Command("git", args...)
	cmd.Dir = gitpath
	util.Debugf(ctx, "run %s\n", cmd)
	util.CheckCmd(cmd.CombinedOutput())
}

// GitClone clones a git repository.
func GitClone(ctx context.Context, pckg *Package, gitpath string) ([]byte, error) {
	args := []string{"clone", pckg.Autoupdate.Target, "."}

	cmd := exec.Command("git", args...)
	cmd.Dir = gitpath
	util.Debugf(ctx, "%s: run %s\n", gitpath, cmd)
	out, err := cmd.CombinedOutput()
	return out, err
}

// GitTags returns the []string of git tags for a package.
func GitTags(ctx context.Context, gitpath string) []string {
	args := []string{"tag"}

	cmd := exec.Command("git", args...)
	cmd.Dir = gitpath
	util.Debugf(ctx, "run %s\n", cmd)
	out := util.CheckCmd(cmd.CombinedOutput())

	tags := make([]string, 0)
	for _, line := range strings.Split(out, "\n") {
		if strings.Trim(line, " ") != "" {
			tags = append(tags, line)
		}
	}

	return tags
}

// GitTimeStamp gets the time stamp for a particular tag (ex. v1.0).
func GitTimeStamp(ctx context.Context, gitpath, tag string) time.Time {
	args := []string{"log", "-1", "--format=%aI", tag}

	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command("git", args...)
	cmd.Dir = gitpath
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	util.Debugf(ctx, "run %s\n", cmd)
	err := cmd.Run()

	if err != nil {
		util.Errf(ctx, "%s: %s\n", err, stderr.String())
		return time.Unix(0, 0)
	}

	t, err := time.Parse(time.RFC3339, strings.TrimSpace(out.String()))
	util.Check(err)

	return t
}

// GitForceCheckout force checkouts a particular tag.
func GitForceCheckout(ctx context.Context, gitpath, tag string) {
	args := []string{"checkout", tag, "-f"}

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
