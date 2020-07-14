package git

import (
	"context"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/util"
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
	packages.GitForceCheckout(ctx, dir, g.Tag)
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
func GetVersions(ctx context.Context, pckg *packages.Package, packageGitcache string) ([]Version, *string) {
	gitTags := packages.GitTags(ctx, packageGitcache)
	util.Debugf(ctx, "found tags in git: %s\n", gitTags)

	gitVersions := make([]Version, 0)
	for _, tag := range gitTags {
		version := strings.TrimPrefix(tag, "v")

		if _, err := semver.Parse(version); err != nil {
			util.Debugf(ctx, "ignoring non-semver git version: %s\n", version)
			continue
		}

		gitVersions = append(gitVersions, Version{
			Tag:       tag,
			Version:   version,
			TimeStamp: packages.GitTimeStamp(ctx, packageGitcache, tag),
		})
	}

	if latest := GetMostRecentVersion(gitVersions); latest != nil {
		return gitVersions, &latest.Version
	}
	return gitVersions, nil
}
