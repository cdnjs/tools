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
	gitv5 "github.com/go-git/go-git/v5"
)

// Version represents a version of a git repo.
type Version struct {
	Tag       string
	Version   string
	TimeStamp time.Time
}

// Get gets the version of a particular Version.
func (g Version) Get() string {
	return g.Version
}

// Download will git check out a particular version.
func (g Version) Download(args ...interface{}) string {
	ctx, dir := args[0].(context.Context), args[1].(string)
	ForceCheckout(ctx, dir, g.Tag)
	return dir // download dir is the same as original dir
}

// Clean is used to satisfy the checker's version interface.
func (g Version) Clean(_ string) {
}

// GetTimeStamp gets the time stamp for a particular git version.
func (g Version) GetTimeStamp() time.Time {
	return g.TimeStamp
}

// GetVersions gets all of the versions associated with a git repo,
// as well as the latest version.
func GetVersions(ctx context.Context, packageGitcache string) ([]Version, *string) {
	gitTags := Tags(ctx, packageGitcache)
	util.Debugf(ctx, "found tags in git: %s\n", gitTags)

	gitVersions := make([]Version, 0)
	for _, tag := range gitTags {
		version := strings.TrimPrefix(tag, "v")
		gitVersions = append(gitVersions, Version{
			Tag:       tag,
			Version:   version,
			TimeStamp: TimeStamp(ctx, packageGitcache, tag),
		})
	}

	if latest := GetMostRecentVersion(gitVersions); latest != nil {
		return gitVersions, &latest.Version
	}
	return gitVersions, nil
}

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
func Add(ctx context.Context, w *gitv5.Worktree, relpath string) {
	util.Debugf(ctx, "go-git: git add %s\n", relpath)

	_, err := w.Add(relpath)
	util.Check(err)
}

// Commit makes a new commit.
func Commit(ctx context.Context, r *gitv5.Repository, w *gitv5.Worktree, msg string) {
	util.Debugf(ctx, "go-git: git commit -m \"%s\"\n", msg)

	commit, err := w.Commit(msg, &gitv5.CommitOptions{})
	util.Check(err)

	_, err = r.CommitObject(commit)
	util.Check(err)
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
func Push(ctx context.Context, r *gitv5.Repository) {
	util.Debugf(ctx, "go-git: git push")

	util.Check(r.Push(&gitv5.PushOptions{}))
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

// Tags returns the []string of git tags for a package.
func Tags(ctx context.Context, gitpath string) []string {
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

// Repo gets a git repository and worktree from a path, panicking on error.
func Repo(repoPath string) (*gitv5.Repository, *gitv5.Worktree) {
	r, err := gitv5.PlainOpen(repoPath)
	util.Check(err)
	w, err := r.Worktree()
	util.Check(err)
	return r, w
}
