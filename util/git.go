package util

import (
	"context"
	"os/exec"
	"strings"
)

// UpdateGitRepo git pulls and rebases the repository.
func UpdateGitRepo(ctx context.Context, gitpath string) {
	args := []string{"pull", "--rebase"}

	cmd := exec.Command("git", args...)
	cmd.Dir = gitpath
	Debugf(ctx, "%s: run %s\n", gitpath, cmd)
	CheckCmd(cmd.CombinedOutput())
}

// IsPathIgnoredByGit determines if a path is git ignored.
func IsPathIgnoredByGit(ctx context.Context, gitpath string, path string) bool {
	// We don't know if "path" is a file or a directory, so let's try with and without /
	return isPathIgnoredByGit(ctx, gitpath, path) || isPathIgnoredByGit(ctx, gitpath, path+"/")
}

func isPathIgnoredByGit(ctx context.Context, gitpath string, path string) bool {
	// We need a relative path, so let's remove "gitpath"
	path = strings.TrimPrefix(path, gitpath)
	// In case "path" is a absolute path, we need to remove "/" afterwards to get a relative path
	path = strings.TrimPrefix(path, "/")
	args := []string{"check-ignore", "--quiet", "--no-index", path}

	cmd := exec.Command("git", args...)
	cmd.Dir = gitpath
	Debugf(ctx, "%s: run %s\n", gitpath, cmd)

	return cmd.Run() == nil
}
