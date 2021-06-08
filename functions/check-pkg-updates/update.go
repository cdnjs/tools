package check_pkg_updates

import (
	"context"
	"log"
	"path"
	"sort"
	"strings"

	"github.com/cdnjs/tools/audit"
	"github.com/cdnjs/tools/gcp"
	"github.com/cdnjs/tools/git"
	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/util"
	"github.com/cdnjs/tools/version"

	"github.com/pkg/errors"
)

func updatePackage(ctx context.Context, pkg *packages.Package) error {
	existingVersionSet, err := getExistingVersions(pkg)
	if err != nil {
		return errors.Wrap(err, "could not detect existing versions")
	}
	log.Printf("%s: existing versions: %s\n", *pkg.Name, strings.Join(existingVersionSet, ","))

	gitVersions, _ := git.GetVersions(ctx, pkg.Autoupdate)
	lastExistingVersion, _ := version.GetMostRecentExistingVersion(ctx, existingVersionSet, gitVersions)

	if lastExistingVersion != nil {
		log.Printf("%s: last existing version: %s\n", *pkg.Name, lastExistingVersion.Version)

		versionDiff := version.VersionDiff(gitVersions, existingVersionSet)
		sort.Sort(version.ByDate(versionDiff))

		newGitVersions := make([]version.Version, 0)

		for i := len(versionDiff) - 1; i >= 0; i-- {
			v := versionDiff[i]
			if v.Date.After(lastExistingVersion.Date) {
				newGitVersions = append(newGitVersions, v)
			}
		}

		sort.Sort(sort.Reverse(version.ByDate(gitVersions)))

		go func(ctx context.Context, pkg *packages.Package, gitVersions []version.Version) {
			if err := DoUpdate(ctx, pkg, newGitVersions); err != nil {
				log.Printf("%s: failed to update new version: %s\n", *pkg.Name, err)
			}
		}(ctx, pkg, gitVersions)
	} else {
		if len(existingVersionSet) > 0 {
			// all existing versions are not on git anymore
			log.Printf("%s: all existing versions not on git\n", *pkg.Name)
		}
		// Import all the versions since we have no current git versions locally.
		// Limit the number of version to an arbitrary number to avoid publishing
		// too many outdated versions.
		sort.Sort(sort.Reverse(version.ByDate(gitVersions)))

		if len(gitVersions) > util.ImportAllMaxVersions {
			gitVersions = gitVersions[len(gitVersions)-util.ImportAllMaxVersions:]
		}

		// Reverse the array to have the older versions first
		// It matters when we will commit the updates
		sort.Sort(sort.Reverse(version.ByDate(gitVersions)))

		go func(ctx context.Context, pkg *packages.Package, gitVersions []version.Version) {
			if err := DoUpdate(ctx, pkg, gitVersions); err != nil {
				log.Printf("%s: failed to import all versions: %s\n", *pkg.Name, err)
			}
		}(ctx, pkg, gitVersions)
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
	if err := gcp.AddIncomingFile(path.Base(v.Tarball), tarball, pkg, v); err != nil {
		return errors.Wrap(err, "could not store in GCS: %s")
	}

	if err := audit.NewVersionDetected(ctx, *pkg.Name, v.Version); err != nil {
		return errors.Wrap(err, "could not audit")
	}

	return nil
}
