package util

import (
	"context"
	"os/exec"
)

func UpdateGitRepo(ctx context.Context, gitpath string) {
	args := []string{"pull", "--rebase"}

	cmd := exec.Command("git", args...)
	cmd.Dir = gitpath
	Debugf(ctx, "%s: run %s\n", gitpath, cmd)
	CheckCmd(cmd.CombinedOutput())
}
