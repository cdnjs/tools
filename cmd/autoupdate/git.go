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

func isValidGit(ctx context.Context, pckgdir string) bool {
	_, err := os.Stat(path.Join(pckgdir, ".git"))
	return !os.IsNotExist(err)
}

func updateGit(ctx context.Context, pckg *packages.Package) []newVersionToCommit {
	var newVersionsToCommit []newVersionToCommit

	packageGitcache := path.Join(GIT_CACHE, pckg.Name)
	// If the local copy of the package's git doesn't exists, Clone it. If it does
	// just fetch new tags
	if _, err := os.Stat(packageGitcache); os.IsNotExist(err) {
		util.Check(os.MkdirAll(packageGitcache, os.ModePerm))

		out, err := packages.GitClone(ctx, pckg, packageGitcache)
		if err != nil {
			util.Printf(ctx, "could not clone repo: %s: %s\n", err, out)
			return newVersionsToCommit
		}
	} else {
		if isValidGit(ctx, packageGitcache) {
			packages.GitFetch(ctx, packageGitcache)
		} else {
			util.Printf(ctx, "invalid git repo\n")
			return newVersionsToCommit
		}
	}

	gitTags := packages.GitTags(ctx, pckg, packageGitcache)
	util.Debugf(ctx, "found tags in git: %s\n", gitTags)

	gitVersions := make([]git.GitVersion, 0)
	for _, tag := range gitTags {
		gitVersions = append(gitVersions, git.GitVersion{
			Tag:     tag,
			Version: strings.TrimPrefix(tag, "v"),
		})
	}

	existingVersionSet := pckg.Versions()
	lastExistingVersion := getLatestExistingVersion(existingVersionSet)

	if lastExistingVersion != nil {
		util.Debugf(ctx, "last existing version: %s\n", lastExistingVersion)

		versionDiff := gitVersionDiff(gitVersions, existingVersionSet)

		newGitVersions := make([]git.GitVersion, 0)

		for i := len(versionDiff) - 1; i >= 0; i-- {
			gitVersion, err := semver.Make(versionDiff[i].Version)
			if err != nil {
				continue
			}

			if gitVersion.Compare(*lastExistingVersion) == 1 {
				newGitVersions = append(newGitVersions, versionDiff[i])
			}
		}

		util.Debugf(ctx, "new versions: %s\n", newGitVersions)

		sort.Sort(sort.Reverse(git.ByGitVersion(newGitVersions)))

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

func doUpdateGit(ctx context.Context, pckg *packages.Package, gitpath string, versions []git.GitVersion) []newVersionToCommit {
	newVersionsToCommit := make([]newVersionToCommit, 0)

	if len(versions) == 0 {
		return newVersionsToCommit
	}

	for _, gitversion := range versions {
		packages.GitForceCheckout(ctx, pckg, gitpath, gitversion.Tag)
		filesToCopy := pckg.NpmFilesFrom(gitpath)

		pckgpath := path.Join(pckg.Path(), gitversion.Version)

		if _, err := os.Stat(pckgpath); !os.IsNotExist(err) {
			util.Debugf(ctx, "%s already exists; aborting\n", pckgpath)
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
				newVersion:  gitversion.Version,
				pckg:        pckg,
			})
		} else {
			util.Debugf(ctx, "no files matched\n")
		}
	}

	return newVersionsToCommit
}

func gitVersionDiff(a []git.GitVersion, b []string) []git.GitVersion {
	diff := make([]git.GitVersion, 0)
	m := make(map[string]bool)

	for _, item := range b {
		m[item] = true
	}

	for _, item := range a {
		if _, ok := m[item.Version]; !ok {
			diff = append(diff, item)
		}
	}

	return diff
}
