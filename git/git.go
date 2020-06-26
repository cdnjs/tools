package git

import (
	"context"
	"strings"

	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/util"
)

// Version represents a version of a git repo.
type Version struct {
	Tag     string
	Version string
}

// Get gets the version of a particular Version.
func (g Version) Get() string {
	return g.Version
}

// Download will git check out a particular version.
func (g Version) Download(args ...interface{}) string {
	ctx, p, dir := args[0].(context.Context), args[1].(*packages.Package), args[2].(string)
	packages.GitForceCheckout(ctx, p, dir, g.Tag)
	return dir // download dir is the same as original dir
}

// Clean is used to satisfy the checker's version interface.
func (g Version) Clean(_ string) {
}

// GetVersions gets all of the versions associated with a git repo.
func GetVersions(ctx context.Context, pckg *packages.Package, packageGitcache string) []Version {
	gitTags := packages.GitTags(ctx, pckg, packageGitcache)
	util.Debugf(ctx, "found tags in git: %s\n", gitTags)

	gitVersions := make([]Version, 0)
	for _, tag := range gitTags {
		gitVersions = append(gitVersions, Version{
			Tag:     tag,
			Version: strings.TrimPrefix(tag, "v"),
		})
	}

	return gitVersions
}
