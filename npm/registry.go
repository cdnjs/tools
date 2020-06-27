package npm

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/cdnjs/tools/util"
)

// RegistryPackage contains metadata about the versions for
// a particular npm package.
type RegistryPackage struct {
	Versions map[string]interface{} `json:"versions"`
}

// Version represents a version of an npm package.
type Version struct {
	Version string
	Tarball string
}

// Get gets the version of a particular Version.
func (n Version) Get() string {
	return n.Version
}

// Download will download a particular npm version.
func (n Version) Download(args ...interface{}) string {
	ctx := args[0].(context.Context)
	return DownloadTar(ctx, n.Tarball) // return download dir
}

// Clean is used to clean up a download directory.
func (n Version) Clean(downloadDir string) {
	os.RemoveAll(downloadDir) // clean up temp tarball dir
}

// MonthlyDownload holds the number of monthly downloads
// for an npm package.
type MonthlyDownload struct {
	Downloads uint `json:"downloads"`
}

func getProtocol() string {
	if util.HasHTTProxy() {
		return "http"
	} else {
		return "https"
	}
}

// Exists determines if an npm package exists.
func Exists(name string) bool {
	resp, err := http.Get(getProtocol() + "://registry.npmjs.org/" + name)
	util.Check(err)
	return resp.StatusCode == http.StatusOK
}

// GetMonthlyDownload uses the npm API to get the MonthlyDownload
// for a particular npm package.
func GetMonthlyDownload(name string) MonthlyDownload {
	resp, err := http.Get(getProtocol() + "://api.npmjs.org/downloads/point/last-month/" + name)
	util.Check(err)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	util.Check(err)

	var counts MonthlyDownload
	util.Check(json.Unmarshal(body, &counts))
	return counts
}

// GetVersions gets all of the versions associated with an npm package.
func GetVersions(name string) []Version {
	resp, err := http.Get(getProtocol() + "://registry.npmjs.org/" + name)
	util.Check(err)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	util.Check(err)

	var npmRegistryPackage RegistryPackage
	util.Check(json.Unmarshal(body, &npmRegistryPackage))

	versions := make([]Version, 0)

	for k, v := range npmRegistryPackage.Versions {
		if v, ok := v.(map[string]interface{}); ok {
			dist := v["dist"].(map[string]interface{})

			versions = append(versions, Version{
				Version: k,
				Tarball: dist["tarball"].(string),
			})
		}
	}

	return versions
}
