package check_pkg_updates

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/sentry"
	"github.com/cdnjs/tools/util"
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
	logger := util.GetStandardLogger()

	list, err := packages.FetchPackages()
	if err != nil {
		http.Error(w, "failed to fetch packages", 500)
		fmt.Println(err)
		return
	}

	for _, pkg := range list {
		if *pkg.Name == "hi-sven" {
			ctx := util.ContextWithEntries(
				util.GetStandardEntries(*pkg.Name, logger)...)
			util.Infof(ctx, "process")

			// select {
			// case sig := <-c:
			// 	util.Debugf(ctx, "RECEIVED SIGNAL: %s\n", sig)
			// 	return
			// default:
			// }

			var newVersionsToCommit []newVersionToCommit
			var allVersions []version

			if pkg.Autoupdate == nil {
				// package not configured to auto update; skip.
				continue
			}

			switch *pkg.Autoupdate.Source {
			case "npm":
				{
					newVersionsToCommit, allVersions = updateNpm(ctx, pkg)
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

			// If there are no versions, do not write any metadata.
			if len(allVersions) <= 0 {
				continue
			}

			log.Println(newVersionsToCommit)
		}
	}

	fmt.Fprint(w, "OK")
}

type APIPackage struct {
	Versions []string `json:"versions"`
}
