package check_pkg_updates

import (
	"context"
	"log"

	"github.com/cdnjs/tools/git"
	"github.com/cdnjs/tools/packages"

	"github.com/pkg/errors"
)

func updateGit(ctx context.Context, pkg *packages.Package) error {
	gitVersions, _ := git.GetVersions(ctx, *pkg.Autoupdate.Target)
	existingVersionSet, err := getExistingVersions(pkg)
	if err != nil {
		return errors.Wrap(err, "could not detect existing versions")
	}

	diff := gitVersionDiff(gitVersions, existingVersionSet)
	log.Println(diff)

	return nil
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
