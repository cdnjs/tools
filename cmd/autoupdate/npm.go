package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"sort"

	"github.com/cdnjs/tools/npm"
	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/util"
)

func updateNpm(ctx context.Context, pckg *packages.Package) ([]newVersionToCommit, string) {
	var newVersionsToCommit []newVersionToCommit

	existingVersionSet := pckg.Versions()
	npmVersions, latestNpmVersion := npm.GetVersions(pckg.Autoupdate.Target)
	lastExistingVersion := npm.GetMostRecentExistingVersion(ctx, existingVersionSet, npmVersions)

	if lastExistingVersion != nil {
		util.Debugf(ctx, "last existing version: %s\n", lastExistingVersion.Version)

		versionDiff := npmVersionDiff(npmVersions, existingVersionSet)
		sort.Sort(npm.ByTimeStamp(versionDiff))

		newNpmVersions := make([]npm.Version, 0)

		for i := len(versionDiff) - 1; i >= 0; i-- {
			v := versionDiff[i]
			if v.TimeStamp.After(lastExistingVersion.TimeStamp) {
				newNpmVersions = append(newNpmVersions, v)
			}
		}

		sort.Sort(sort.Reverse(npm.ByTimeStamp(npmVersions)))

		newVersionsToCommit = doUpdateNpm(ctx, pckg, newNpmVersions)
	} else {
		if len(existingVersionSet) > 0 {
			// all existing versions are not on npm anymore
			// so we will ignore this package
			util.Debugf(ctx, "ignoring misconfigured npm package: %s", pckg.Name)
		} else {
			// Import all the versions since we have none locally.
			// Limit the number of version to an abrirary number to avoid publishing
			// too many outdated versions.
			sort.Sort(sort.Reverse(npm.ByTimeStamp(npmVersions)))

			if len(npmVersions) > util.ImportAllMaxVersions {
				npmVersions = npmVersions[len(npmVersions)-util.ImportAllMaxVersions:]
			}

			npmVersionsStr := make([]string, len(npmVersions))
			for i, npmVersion := range npmVersions {
				npmVersionsStr[i] = npmVersion.Version
			}

			// Reverse the array to have the older versions first
			// It matters when we will commit the updates
			sort.Sort(sort.Reverse(npm.ByTimeStamp(npmVersions)))

			newVersionsToCommit = doUpdateNpm(ctx, pckg, npmVersions)
		}
	}

	return newVersionsToCommit, latestNpmVersion
}

func doUpdateNpm(ctx context.Context, pckg *packages.Package, versions []npm.Version) []newVersionToCommit {
	newVersionsToCommit := make([]newVersionToCommit, 0)

	if len(versions) == 0 {
		return newVersionsToCommit
	}

	for _, version := range versions {
		pckgpath := path.Join(pckg.Path(), version.Version)

		if _, err := os.Stat(pckgpath); !os.IsNotExist(err) {
			util.Debugf(ctx, "%s already exists; aborting", pckgpath)
			continue
		}

		if util.IsPathIgnoredByGit(ctx, util.GetCDNJSPath(), pckgpath) {
			util.Debugf(ctx, "%s is ignored by git; aborting\n", pckgpath)
			continue
		}

		util.Check(os.MkdirAll(pckgpath, os.ModePerm))

		tarballDir := npm.DownloadTar(ctx, version.Tarball)
		filesToCopy := pckg.NpmFilesFrom(tarballDir)

		if len(filesToCopy) > 0 {
			for _, fileMoveOp := range filesToCopy {
				absFrom := path.Join(tarballDir, fileMoveOp.From)
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
				newVersion:  version.Version,
				pckg:        pckg,
			})
		} else {
			util.Debugf(ctx, "no files matched")
		}

		// clean up temporary tarball dir
		util.Check(os.RemoveAll(tarballDir))
	}

	return newVersionsToCommit
}

func npmVersionDiff(a []npm.Version, b []string) []npm.Version {
	diff := make([]npm.Version, 0)
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
