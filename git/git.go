package git

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

// ListPackageVersions first lists all the versions (and top-level package.json)
// in the package and passes the list to git ls-tree which filters out
// those not in the tree.
func ListPackageVersions(ctx context.Context, basePath string) []string {
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

// Add adds to the next commit.
func Add(ctx context.Context, gitpath, relpath string) {
	args := []string{"add", relpath}

	cmd := exec.Command("git", args...)
	cmd.Dir = gitpath
	util.Debugf(ctx, "run %s\n", cmd)
	util.CheckCmd(cmd.CombinedOutput())
}

// Commit makes a new commit.
func Commit(ctx context.Context, gitpath, msg string) {
	args := []string{"commit", "-m", msg}

	cmd := exec.Command("git", args...)
	cmd.Dir = gitpath
	util.Debugf(ctx, "run %s\n", cmd)
	util.CheckCmd(cmd.CombinedOutput())
}

// Fetch fetches objs/refs to the repository.
func Fetch(ctx context.Context, gitpath string) ([]byte, error) {
	args := []string{"fetch"}

	cmd := exec.Command("git", args...)
	cmd.Dir = gitpath
	util.Debugf(ctx, "%s: run %s\n", gitpath, cmd)
	return cmd.CombinedOutput()
}

// Push pushes to a git repository.
func Push(ctx context.Context, gitpath string) {
	args := []string{"push"}

	cmd := exec.Command("git", args...)
	cmd.Dir = gitpath
	util.Debugf(ctx, "run %s\n", cmd)
	util.CheckCmd(cmd.CombinedOutput())
}

// Clone clones a git repository.
func Clone(ctx context.Context, target string, gitpath string) ([]byte, error) {
	args := []string{"clone", target, "."}

	cmd := exec.Command("git", args...)
	cmd.Dir = gitpath
	util.Debugf(ctx, "%s: run %s\n", gitpath, cmd)
	out, err := cmd.CombinedOutput()
	return out, err
}

// TimeStamp gets the time stamp for a particular tag (ex. v1.0).
func TimeStamp(ctx context.Context, gitpath, tag string) time.Time {
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

// ForceCheckout force checkouts a particular tag.
func ForceCheckout(ctx context.Context, gitpath, tag string) {
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

// UpdateRepo git pulls and rebases the repository.
func UpdateRepo(ctx context.Context, gitpath string) {
	args := []string{"pull", "--rebase"}

	cmd := exec.Command("git", args...)
	cmd.Dir = gitpath
	util.Debugf(ctx, "%s: run %s\n", gitpath, cmd)
	util.CheckCmd(cmd.CombinedOutput())
}

// IsPathIgnored determines if a path is git ignored.
func IsPathIgnored(ctx context.Context, gitpath string, path string) bool {
	// We don't know if "path" is a file or a directory, so let's try with and without /
	return isPathIgnored(ctx, gitpath, path) || isPathIgnored(ctx, gitpath, path+"/")
}

func isPathIgnored(ctx context.Context, gitpath string, path string) bool {
	// We need a relative path, so let's remove "gitpath"
	path = strings.TrimPrefix(path, gitpath)
	// In case "path" is a absolute path, we need to remove "/" afterwards to get a relative path
	path = strings.TrimPrefix(path, "/")
	args := []string{"check-ignore", "--quiet", "--no-index", path}

	cmd := exec.Command("git", args...)
	cmd.Dir = gitpath
	util.Debugf(ctx, "%s: run %s\n", gitpath, cmd)

	return cmd.Run() == nil
}
