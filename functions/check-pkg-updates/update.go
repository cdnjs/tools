package check_pkg_updates

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/cdnjs/tools/audit"
	"github.com/cdnjs/tools/gcp"
	"github.com/cdnjs/tools/git"
	"github.com/cdnjs/tools/npm"
	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/util"
	"github.com/cdnjs/tools/version"

	"github.com/pkg/errors"
)

func updatePackage(ctx context.Context, pkg *packages.Package, src string) error {
	existingVersionSet, err := getExistingVersions(pkg)
	if err != nil {
		return errors.Wrap(err, "could not detect existing versions")
	}
	log.Printf("%s: existing versions: %s\n", *pkg.Name, strings.Join(existingVersionSet, ","))

	var versions []version.Version

	switch src {
	case "git":
		versions, err = git.GetVersions(ctx, pkg.Autoupdate)
		if err != nil {
			return errors.Wrap(err, "failed to get git versions")
		}
	case "npm":
		versions, _ = npm.GetVersions(ctx, pkg.Autoupdate)
	default:
		panic("unreachable")
	}

	lastExistingVersion, _ := version.GetMostRecentExistingVersion(ctx, existingVersionSet, versions)

	if lastExistingVersion != nil {
		log.Printf("%s: last existing version: %s\n", *pkg.Name, lastExistingVersion.Version)

		versionDiff := version.VersionDiff(versions, existingVersionSet)
		sort.Sort(version.ByDate(versionDiff))

		newVersions := make([]version.Version, 0)

		for i := len(versionDiff) - 1; i >= 0; i-- {
			v := versionDiff[i]
			if v.Date.After(lastExistingVersion.Date) {
				newVersions = append(newVersions, v)
			}
		}

		sort.Sort(sort.Reverse(version.ByDate(versions)))

		go func(ctx context.Context, pkg *packages.Package, versions []version.Version) {
			if err := DoUpdate(ctx, pkg, newVersions); err != nil {
				log.Printf("%s: failed to update new version: %s\n", *pkg.Name, err)
			}
		}(ctx, pkg, versions)
	} else {
		if len(existingVersionSet) > 0 {
			log.Printf("%s: all existing versions not on %s\n", *pkg.Name, src)
		}
		// Import all the versions since we have no current git/npm versions locally.
		// Limit the number of version to an arbitrary number to avoid publishing
		// too many outdated versions.
		sort.Sort(sort.Reverse(version.ByDate(versions)))

		if len(versions) > util.ImportAllMaxVersions {
			versions = versions[len(versions)-util.ImportAllMaxVersions:]
		}

		// Reverse the array to have the older versions first
		// It matters when we will commit the updates
		sort.Sort(sort.Reverse(version.ByDate(versions)))

		go func(ctx context.Context, pkg *packages.Package, versions []version.Version) {
			if err := DoUpdate(ctx, pkg, versions); err != nil {
				log.Printf("%s: failed to import all versions: %s\n", *pkg.Name, err)
			}
		}(ctx, pkg, versions)
	}
	return nil
}

func DoUpdate(ctx context.Context, pkg *packages.Package, versions []version.Version) error {
	if len(versions) == 0 {
		return nil
	}
	// only update one versions at a time to reduce race conditions
	v := versions[0]

	log.Printf("%s: new version detected: %s\n", *pkg.Name, v.Version)
	tarball := version.DownloadTar(ctx, v)
	filename := fmt.Sprintf("%s-%s.tgz", *pkg.Name, v.Version)
	if err := gcp.AddIncomingFile(filename, tarball, pkg, v); err != nil {
		return errors.Wrap(err, "could not store in GCS: %s")
	}

	if err := audit.NewVersionDetected(ctx, *pkg.Name, v.Version); err != nil {
		return errors.Wrap(err, "could not audit")
	}

	return nil
}
