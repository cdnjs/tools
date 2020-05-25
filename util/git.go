package util

import (
	"context"
	"os/exec"
)

func UpdateGitRepo(ctx context.Context, gitpath string) {
	args := []string{"pull", "--rebase"}

	cmd := exec.Command("git", args...)
	cmd.Dir = gitpath
	Debugf(ctx, "run %s\n", cmd)
	CheckCmd(cmd.CombinedOutput())
}
