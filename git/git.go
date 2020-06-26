package git

import (
	"context"
	"strings"

	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/util"
)

type GitVersion struct {
	Tag     string
	Version string
}

// Get gets the version of a particular GitVersion.
func (g *GitVersion) Get() string {
	return g.Version
}

// Download will git check out a particular version.
func (g *GitVersion) Download(args ...interface{}) {
	ctx, p, dir := args[0].(context.Context), args[1].(*packages.Package), args[2].(string)
	packages.GitForceCheckout(ctx, p, dir, g.Tag)
}

// Clean is used to satisfy the checker's version interface.
func (g *GitVersion) Clean() {
}

func GetVersions(ctx context.Context, pckg *packages.Package, packageGitcache string) []GitVersion {
	gitTags := packages.GitTags(ctx, pckg, packageGitcache)
	util.Debugf(ctx, "found tags in git: %s\n", gitTags)

	gitVersions := make([]GitVersion, 0)
	for _, tag := range gitTags {
		gitVersions = append(gitVersions, GitVersion{
			Tag:     tag,
			Version: strings.TrimPrefix(tag, "v"),
		})
	}

	return gitVersions
}
