package check_pkg_updates

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/sentry"
	"github.com/cdnjs/tools/util"

	"github.com/pkg/errors"
)

type version interface {
	Get() string             // Get the version.
	GetTimeStamp() time.Time // GetTimeStamp gets the time stamp associated with the version.
}

type newVersionToCommit struct {
	versionPath string
	newVersion  string
	pckg        *packages.Package
	timestamp   time.Time
}

// Get is used to get the new version.
func (n newVersionToCommit) Get() string {
	return n.newVersion
}

// GetTimeStamp gets the time stamp associated with the new version.
func (n newVersionToCommit) GetTimeStamp() time.Time {
	return n.timestamp
}

func Invoke(w http.ResponseWriter, r *http.Request) {
	sentry.Init()
	defer sentry.PanicHandler()

	list, err := packages.FetchPackages()
	if err != nil {
		http.Error(w, "failed to fetch packages", 500)
		fmt.Println(err)
		return
	}

	for _, pkg := range list {
		if err := checkPackage(pkg); err != nil {
			log.Printf("failed to update package %s: %s", *pkg.Name, err)
		}
	}

	fmt.Fprint(w, "OK")
}

func checkPackage(pkg *packages.Package) error {
	logger := util.GetStandardLogger()
	ctx := util.ContextWithEntries(
		util.GetStandardEntries(*pkg.Name, logger)...)

	var newVersionsToCommit []newVersionToCommit
	var allVersions []version

	if pkg.Autoupdate == nil {
		// package not configured to auto update; skip.
		return nil
	}

	var err error

	switch *pkg.Autoupdate.Source {
	case "npm":
		{
			newVersionsToCommit, allVersions, err = updateNpm(ctx, pkg)
		}
		// case "git":
		// 	{
		// 		util.Debugf(ctx, "running git update")
		// 		newVersionsToCommit, allVersions = updateGit(ctx, pckg)
		// 	}
		// default:
		// 	{
		// 		panic(fmt.Sprintf("%s invalid autoupdate source: %s", *pckg.Name, *pckg.Autoupdate.Source))
		// 	}
	}

	if err != nil {
		return errors.Wrap(err, "failed to update package")
	}

	// If there are no versions, do not write any metadata.
	if len(allVersions) <= 0 {
		return nil
	}

	log.Println(newVersionsToCommit)
	return nil
}
