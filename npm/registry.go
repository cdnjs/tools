package npm

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/cdnjs/tools/util"
)

type NpmRegistryPackage struct {
	Versions map[string]interface{} `json:"versions"`
}

type NpmVersion struct {
	tarballDir string

	Version string
	Tarball string
}

// Get gets the version of a particular NpmVersion.
func (n *NpmVersion) Get() string {
	return n.Version
}

// Download will download a particular npm version.
func (n *NpmVersion) Download(args ...interface{}) {
	ctx := args[0].(context.Context)
	n.tarballDir = DownloadTar(ctx, n.Tarball)
}

// Clean is used to satisfy the checker's version interface.
func (n *NpmVersion) Clean() {
	os.RemoveAll(n.tarballDir) // clean up temp tarball dir
}

type MonthlyDownload struct {
	Downloads uint `json:"downloads"`
}

func Exists(name string) bool {
	resp, err := http.Get("https://registry.npmjs.org/" + name)
	util.Check(err)
	return resp.StatusCode == http.StatusOK
}

func GetMonthlyDownload(name string) MonthlyDownload {
	resp, err := http.Get("https://api.npmjs.org/downloads/point/last-month/" + name)
	util.Check(err)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	util.Check(err)

	var counts MonthlyDownload
	util.Check(json.Unmarshal(body, &counts))
	return counts
}

func GetVersions(name string) []NpmVersion {
	resp, err := http.Get("https://registry.npmjs.org/" + name)
	util.Check(err)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	util.Check(err)

	var npmRegistryPackage NpmRegistryPackage
	util.Check(json.Unmarshal(body, &npmRegistryPackage))

	versions := make([]NpmVersion, 0)

	for k, v := range npmRegistryPackage.Versions {
		if v, ok := v.(map[string]interface{}); ok {
			dist := v["dist"].(map[string]interface{})

			versions = append(versions, NpmVersion{
				Version: k,
				Tarball: dist["tarball"].(string),
			})
		}
	}

	return versions
}
