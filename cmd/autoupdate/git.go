package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"sort"

	"github.com/cdnjs/tools/git"
	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/util"
)

var (
	gitCache = path.Join(basePath, "git-cache")
)

func isValidGit(ctx context.Context, pckgdir string) bool {
	_, err := os.Stat(path.Join(pckgdir, ".git"))
	return !os.IsNotExist(err)
}

func updateGit(ctx context.Context, pckg *packages.Package) ([]newVersionToCommit, *newVersionToCommit) {
	var newVersionsToCommit []newVersionToCommit

	packageGitcache := path.Join(gitCache, pckg.Name)
	// If the local copy of the package's git doesn't exists, Clone it. If it does
	// just fetch new tags
	if _, err := os.Stat(packageGitcache); os.IsNotExist(err) {
		util.Check(os.MkdirAll(packageGitcache, os.ModePerm))

		out, err := packages.GitClone(ctx, pckg, packageGitcache)
		if err != nil {
			util.Errf(ctx, "could not clone repo: %s: %s\n", err, out)
			return newVersionsToCommit, nil
		}
	} else {
		if isValidGit(ctx, packageGitcache) {
			out, fetcherr := packages.GitFetch(ctx, packageGitcache)
			if fetcherr != nil {
				util.Errf(ctx, "could not fetch repo %s: %s\n", fetcherr, out)
				return newVersionsToCommit, nil
			}
		} else {
			util.Errf(ctx, "invalid git repo\n")
			return newVersionsToCommit, nil
		}
	}

	gitVersions, latestGitVersion := git.GetVersions(ctx, pckg, packageGitcache)
	existingVersionSet := pckg.Versions()
	lastExistingVersion := git.GetMostRecentExistingVersion(ctx, existingVersionSet, gitVersions)

	if lastExistingVersion != nil {
		util.Debugf(ctx, "last existing version: %s\n", lastExistingVersion.Version)

		versionDiff := gitVersionDiff(gitVersions, existingVersionSet)

		newGitVersions := make([]git.Version, 0)

		for i := len(versionDiff) - 1; i >= 0; i-- {
			v := versionDiff[i]
			if v.TimeStamp.After(lastExistingVersion.TimeStamp) {
				newGitVersions = append(newGitVersions, v)
			}
		}

		util.Debugf(ctx, "new versions: %s\n", newGitVersions)

		sort.Sort(sort.Reverse(git.ByTimeStamp(newGitVersions)))

		newVersionsToCommit = doUpdateGit(ctx, pckg, packageGitcache, newGitVersions)
	} else {
		if len(existingVersionSet) > 0 {
			// all existing versions are not on git anymore
			// so we will ignore this package
			util.Debugf(ctx, "ignoring misconfigured git package: %s", pckg.Name)
		} else {
			// Import all the versions since we have none locally.
			// Limit the number of version to an abrirary number to avoid publishing
			// too many outdated versions.
			sort.Sort(sort.Reverse(git.ByTimeStamp(gitVersions)))

			if len(gitVersions) > util.ImportAllMaxVersions {
				gitVersions = gitVersions[len(gitVersions)-util.ImportAllMaxVersions:]
			}

			// Reverse the array to have the older versions first
			// It matters when we will commit the updates
			sort.Sort(sort.Reverse(git.ByTimeStamp(gitVersions)))

			newVersionsToCommit = doUpdateGit(ctx, pckg, packageGitcache, gitVersions)
		}
	}

	return newVersionsToCommit, &newVersionToCommit{commitType: "newLatestVersion", newVersion: latestGitVersion, pckg: pckg}
}

func doUpdateGit(ctx context.Context, pckg *packages.Package, gitpath string, versions []git.Version) []newVersionToCommit {
	newVersionsToCommit := make([]newVersionToCommit, 0)

	if len(versions) == 0 {
		return newVersionsToCommit
	}

	for _, gitversion := range versions {
		packages.GitForceCheckout(ctx, gitpath, gitversion.Tag)
		filesToCopy := pckg.NpmFilesFrom(gitpath)

		pckgpath := path.Join(pckg.Path(), gitversion.Version)

		if _, err := os.Stat(pckgpath); !os.IsNotExist(err) {
			util.Debugf(ctx, "%s already exists; aborting\n", pckgpath)
			continue
		}

		if util.IsPathIgnoredByGit(ctx, util.GetCDNJSPath(), pckgpath) {
			util.Debugf(ctx, "%s is ignored by git; aborting\n", pckgpath)
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
				commitType: "newVersion",
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

func gitVersionDiff(a []git.Version, b []string) []git.Version {
	diff := make([]git.Version, 0)
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
