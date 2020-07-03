package git

import (
	"context"
	"strings"
	"time"

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
func GetVersions(ctx context.Context, gitpath string) ([]Version, *string) {
	gitTags := packages.GitTags(ctx, gitpath)
	util.Debugf(ctx, "found tags in git: %s\n", gitTags)

	// Per: https://github.com/git/git/blob/101b3204f37606972b40fc17dec84560c22f69f6/builtin/clone.c#L1003
	isRemoteRepository := strings.Contains(gitpath, ":")

	gitVersions := make([]Version, 0)
	for _, tag := range gitTags {
		version := strings.TrimPrefix(tag, "v")
		var timeStamp time.Time
		if isRemoteRepository {
			timeStamp = time.Unix(0, 0)
		} else {
			timeStamp = packages.GitTimeStamp(ctx, gitpath, tag)
		}

		gitVersions = append(gitVersions, Version{
			Tag:       tag,
			Version:   version,
			TimeStamp: timeStamp,
		})
	}

	if !isRemoteRepository {
		if latest := GetMostRecentVersion(gitVersions); latest != nil {
			return gitVersions, &latest.Version
		}
	}
	return gitVersions, nil
}
