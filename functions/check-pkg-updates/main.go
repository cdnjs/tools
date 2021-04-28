package check_pkg_updates

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/sentry"
	"github.com/cdnjs/tools/util"

	"github.com/pkg/errors"
)

const PACKAGES_ZIP = "https://github.com/cdnjs/packages/archive/refs/heads/master.zip"

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

	packages, err := fetchPackages()
	if err != nil {
		http.Error(w, "failed to fetch packages", 500)
		fmt.Println(err)
		return
	}

	for _, pkg := range packages {
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

func fetchPackages() ([]*packages.Package, error) {
	zipfile, err := ioutil.TempFile("", "zip")
	if err != nil {
		return nil, errors.Wrap(err, "could not create temp file")
	}
	defer os.Remove(zipfile.Name())

	resp, err := http.Get(PACKAGES_ZIP)
	if err != nil {
		return nil, errors.Wrap(err, "could not fetch packages")
	}
	defer resp.Body.Close()
	_, err = io.Copy(zipfile, resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "could not download packages zip")
	}

	packages, err := inflatePackages(zipfile)
	if err != nil {
		return nil, errors.Wrap(err, "could not inflate packages")
	}
	return packages, nil
}

func inflatePackages(src *os.File) ([]*packages.Package, error) {
	var list []*packages.Package

	r, err := zip.OpenReader(src.Name())
	if err != nil {
		return nil, err
	}
	defer r.Close()

	prefix := "packages-master/packages"
	// FIXME: pass from root
	ctx := context.Background()

	for _, f := range r.File {
		if strings.HasPrefix(f.Name, prefix) && strings.HasSuffix(f.Name, ".json") {
			reader, err := f.Open()
			if err != nil {
				return nil, errors.Wrap(err, "could open file")
			}
			bytes, err := ioutil.ReadAll(reader)
			if err != nil {
				return nil, errors.Wrap(err, "could not read file")
			}

			pkg, err := packages.ReadHumanJSONBytes(ctx, f.Name, bytes)
			if err != nil {
				return nil, errors.Wrapf(err, "could not parse Package: %s", f.Name)
			}

			list = append(list, pkg)
		}
	}
	return list, nil
}
