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
