package npm

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/cdnjs/tools/util"
)

// Registry contains metadata about a particular npm package.
type Registry struct {
	Versions   map[string]interface{} `json:"versions"`  // Versions contains metadata about each npm version.
	TimeStamps map[string]interface{} `json:"time"`      // TimeStamps contains times for each versions as well as the created/modified time.
	DistTags   map[string]string      `json:"dist-tags"` // DistTags map dist tags to string versions
}

// Version represents a version of an npm package.
type Version struct {
	Version   string
	Tarball   string
	TimeStamp time.Time
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

// Gets the protocol, either http or https.
func getProtocol() string {
	if util.HasHTTProxy() {
		return "http"
	}
	return "https"
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

// GetVersions gets all of the versions associated with an npm package,
// as well as the latest version based on the `latest`.
func GetVersions(name string) ([]Version, string) {
	resp, err := http.Get(getProtocol() + "://registry.npmjs.org/" + name)
	util.Check(err)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	util.Check(err)

	var r Registry
	util.Check(json.Unmarshal(body, &r))

	versions := make([]Version, 0)
	for k, v := range r.Versions {
		if v, ok := v.(map[string]interface{}); ok {
			if timeInt, ok := r.TimeStamps[k]; ok {
				if timeStr, ok := timeInt.(string); ok {
					// parse time.Time from time stamp
					timeStamp, err := time.Parse(time.RFC3339, timeStr)
					util.Check(err)

					dist := v["dist"].(map[string]interface{})

					versions = append(versions, Version{
						Version:   k,
						Tarball:   dist["tarball"].(string),
						TimeStamp: timeStamp,
					})
					continue
				}
			}
			panic(fmt.Sprintf("no time stamp for npm version %s", k))
		}
	}

	// get latest version according to npm
	latest, ok := r.DistTags["latest"]
	if !ok {
		panic(fmt.Sprintf("no latest tag for npm package %s", name))
	}

	return versions, latest
}
