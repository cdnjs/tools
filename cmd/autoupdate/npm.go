package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"sort"

	"github.com/blang/semver"

	"github.com/cdnjs/tools/npm"
	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/util"
)

func updateNpm(ctx context.Context, pckg *packages.Package) []newVersionToCommit {
	var newVersionsToCommit []newVersionToCommit
	existingVersionSet := pckg.Versions()

	// We currently ignore any non-semver versions
	// FIXME: handle them?
	existingVersionSet = getSemverOnly(existingVersionSet)

	npmVersions := npm.GetVersions(pckg.Autoupdate.Target)

	if len(existingVersionSet) > 0 {
		lastExistingVersion, err := semver.Make(existingVersionSet[len(existingVersionSet)-1])
		util.Check(err)

		versionDiff := npmVersionDiff(npmVersions, existingVersionSet)
		sort.Sort(npm.ByNpmVersion(versionDiff))

		newNpmVersions := make([]npm.NpmVersion, 0)

		for i := len(versionDiff) - 1; i >= 0; i-- {
			npmVersion, err := semver.Make(versionDiff[i].Version)
			if err != nil {
				continue
			}

			if npmVersion.Compare(lastExistingVersion) == 1 {
				newNpmVersions = append(newNpmVersions, versionDiff[i])
			}
		}

		newVersionsToCommit = doUpdateNpm(ctx, pckg, newNpmVersions)
	} else {
		// Import all the versions since we have none locally.
		// Limit the number of version to an abrirary number to avoid publishing
		// too many outdated versions.
		sort.Sort(sort.Reverse(npm.ByNpmVersion(npmVersions)))

		if len(npmVersions) > util.IMPORT_ALL_MAX_VERSIONS {
			npmVersions = npmVersions[len(npmVersions)-util.IMPORT_ALL_MAX_VERSIONS:]
		}

		npmVersionsStr := make([]string, len(npmVersions))
		for i, npmVersion := range npmVersions {
			npmVersionsStr[i] = npmVersion.Version
		}

		// Reverse the array to have the older versions first
		// It matters when we will commit the updates
		sort.Sort(sort.Reverse(npm.ByNpmVersion(npmVersions)))

		log(ctx, LogImportAllVersions{Versions: npmVersionsStr})

		newVersionsToCommit = doUpdateNpm(ctx, pckg, npmVersions)
	}

	return newVersionsToCommit
}

func doUpdateNpm(ctx context.Context, pckg *packages.Package, versions []npm.NpmVersion) []newVersionToCommit {
	newVersionsToCommit := make([]newVersionToCommit, 0)

	if len(versions) == 0 {
		log(ctx, LogNoNewVersion{})
		return newVersionsToCommit
	}

	for _, version := range versions {
		pckgpath := path.Join(pckg.Path(), version.Version)

		if _, err := os.Stat(pckgpath); !os.IsNotExist(err) {
			log(ctx, LogNewVersionExistsLocally{Version: version.Version})
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

			log(ctx, LogCreatedNewVersion{Version: version.Version})

			newVersionsToCommit = append(newVersionsToCommit, newVersionToCommit{
				versionPath: pckgpath,
				newVersion:  version.Version,
				pckg:        pckg,
			})
		} else {
			log(ctx, LogNoFilesMatchedThePattern{Version: version.Version})
		}

		// clean up temporary tarball dir
		util.Check(os.RemoveAll(tarballDir))
	}

	return newVersionsToCommit
}

func npmVersionDiff(a []npm.NpmVersion, b []string) []npm.NpmVersion {
	diff := make([]npm.NpmVersion, 0)
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
