package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/blang/semver"

	"github.com/cdnjs/tools/git"
	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/util"
)

var (
	GIT_CACHE = path.Join(BASE_PATH, "git-cache")
)

func updateGit(ctx context.Context, pckg *packages.Package) []newVersionToCommit {
	var newVersionsToCommit []newVersionToCommit

	packageGitcache := path.Join(GIT_CACHE, pckg.Name)

	// If the local copy of the package's git doesn't exists, Clone it. If it does
	// just fetch new tags
	if _, err := os.Stat(packageGitcache); os.IsNotExist(err) {
		util.Check(os.MkdirAll(packageGitcache, os.ModePerm))

		out, err := packages.GitClone(ctx, pckg, packageGitcache)
		if err != nil {
			util.Printf(ctx, "could not clone repo: %s: %s", err, out)
			return newVersionsToCommit
		}
	} else {
		packages.GitFetch(ctx, packageGitcache)
	}

	gitVersions := packages.GitTags(ctx, pckg, packageGitcache)
	util.Debugf(ctx, "found versions in git: %s", gitVersions)

	existingVersionSet := getSemverOnly(pckg.Versions())
	if len(existingVersionSet) > 0 {
		lastExistingVersion, err := semver.Make(existingVersionSet[len(existingVersionSet)-1])
		util.Check(err)
		util.Debugf(ctx, "last exists version: %s", lastExistingVersion)

		versionDiff := gitVersionDiff(gitVersions, existingVersionSet)

		newGitVersions := make([]string, 0)

		for i := len(versionDiff) - 1; i >= 0; i-- {
			gitVersion, err := semver.Make(versionDiff[i])
			if err != nil {
				continue
			}

			if gitVersion.Compare(lastExistingVersion) == 1 {
				newGitVersions = append(newGitVersions, versionDiff[i])
			}
		}

		util.Debugf(ctx, "new versions: %s", newGitVersions)
		newVersionsToCommit = doUpdateGit(ctx, pckg, packageGitcache, newGitVersions)
	} else {
		// Import all the versions since we have none locally.
		// Limit the number of version to an abrirary number to avoid publishing
		// too many outdated versions.
		sort.Sort(sort.Reverse(git.ByGitVersion(gitVersions)))

		if len(gitVersions) > util.IMPORT_ALL_MAX_VERSIONS {
			gitVersions = gitVersions[len(gitVersions)-util.IMPORT_ALL_MAX_VERSIONS:]
		}

		// Reverse the array to have the older versions first
		// It matters when we will commit the updates
		sort.Sort(sort.Reverse(git.ByGitVersion(gitVersions)))

		newVersionsToCommit = doUpdateGit(ctx, pckg, packageGitcache, gitVersions)
	}

	return newVersionsToCommit
}

func doUpdateGit(ctx context.Context, pckg *packages.Package, gitpath string, versions []string) []newVersionToCommit {
	newVersionsToCommit := make([]newVersionToCommit, 0)

	if len(versions) == 0 {
		return newVersionsToCommit
	}

	for _, version := range versions {
		packages.GitCheckout(ctx, pckg, gitpath, version)
		filesToCopy := pckg.NpmFilesFrom(gitpath)

		// Remove the v prefix in the version, for example v1.0.1
		// Note that we do it after the checkout so the git tag is still valid
		version = strings.TrimPrefix(version, "v")

		pckgpath := path.Join(pckg.Path(), version)

		if _, err := os.Stat(pckgpath); !os.IsNotExist(err) {
			continue
		}

		util.Check(os.MkdirAll(pckgpath, os.ModePerm))

		if len(filesToCopy) > 0 {
			for _, fileMoveOp := range filesToCopy {
				absFrom := path.Join(gitpath, fileMoveOp.From)
				absDest := path.Join(pckgpath, fileMoveOp.To)

				if _, err := os.Stat(path.Dir(absDest)); os.IsNotExist(err) {
					util.Check(os.MkdirAll(path.Dir(absDest), os.ModePerm))
				}

				util.Debugf(ctx, "%s -> %s\n", absFrom, absDest)

				err := util.MoveFile(
					absFrom,
					absDest,
				)
				if err != nil {
					fmt.Println("could not move file:", err)
				}
			}

			newVersionsToCommit = append(newVersionsToCommit, newVersionToCommit{
				versionPath: pckgpath,
				newVersion:  version,
				pckg:        pckg,
			})
		}
	}

	return newVersionsToCommit
}

func gitVersionDiff(a []string, b []string) []string {
	diff := make([]string, 0)
	m := make(map[string]bool)

	for _, item := range b {
		m[item] = true
	}

	for _, item := range a {
		if _, ok := m[item]; !ok {
			diff = append(diff, item)
		}
	}

	return diff
}
