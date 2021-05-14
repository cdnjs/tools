package check_pkg_updates

import (
	"context"
	"log"
	"path"
	"sort"
	"strings"

	"github.com/cdnjs/tools/audit"
	"github.com/cdnjs/tools/gcp"
	"github.com/cdnjs/tools/npm"
	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/util"

	"github.com/pkg/errors"
)

func updateNpm(ctx context.Context, pkg *packages.Package) error {
	existingVersionSet, err := getExistingVersions(pkg)
	if err != nil {
		return errors.Wrap(err, "could not detect existing versions")
	}
	log.Printf("%s: existing versions: %s\n", *pkg.Name, strings.Join(existingVersionSet, ","))

	npmVersions, _ := npm.GetVersions(ctx, *pkg.Autoupdate.Target)
	lastExistingVersion, _ := npm.GetMostRecentExistingVersion(ctx, existingVersionSet, npmVersions)

	if lastExistingVersion != nil {
		log.Printf("%s: last existing version: %s\n", *pkg.Name, lastExistingVersion.Version)

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

		if err := DoUpdateNpm(ctx, pkg, newNpmVersions); err != nil {
			return errors.Wrap(err, "failed to update new version")
		}
	} else {
		if len(existingVersionSet) > 0 {
			// all existing versions are not on npm anymore
			log.Printf("%s: all existing versions not on npm\n", *pkg.Name)
		}
		// Import all the versions since we have no current npm versions locally.
		// Limit the number of version to an arbitrary number to avoid publishing
		// too many outdated versions.
		sort.Sort(sort.Reverse(npm.ByTimeStamp(npmVersions)))

		if len(npmVersions) > util.ImportAllMaxVersions {
			npmVersions = npmVersions[len(npmVersions)-util.ImportAllMaxVersions:]
		}

		// Reverse the array to have the older versions first
		// It matters when we will commit the updates
		sort.Sort(sort.Reverse(npm.ByTimeStamp(npmVersions)))

		if err := DoUpdateNpm(ctx, pkg, npmVersions); err != nil {
			return errors.Wrap(err, "failed to import all version")
		}
	}
	return nil
}

func DoUpdateNpm(ctx context.Context, pkg *packages.Package, versions []npm.Version) error {
	if len(versions) == 0 {
		return nil
	}
	// only update one versions at a time to reduce race conditions
	version := versions[0]

	log.Printf("%s: new version detected: %s\n", *pkg.Name, version.Version)
	tarball := npm.DownloadTar(ctx, version.Tarball)
	if err := gcp.AddIncomingFile(path.Base(version.Tarball), tarball, pkg, version); err != nil {
		return errors.Wrap(err, "could not store in GCS: %s")
	}

	if err := audit.NewVersionDetected(ctx, *pkg.Name, version.Version); err != nil {
		return errors.Wrap(err, "could not audit")
	}

	return nil
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
