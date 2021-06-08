package npm

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/util"
	"github.com/cdnjs/tools/version"
)

// Registry contains metadata about a particular npm package.
type Registry struct {
	Versions   map[string]interface{} `json:"versions"`  // Versions contains metadata about each npm version.
	TimeStamps map[string]interface{} `json:"time"`      // TimeStamps contains times for each versions as well as the created/modified time.
	DistTags   map[string]string      `json:"dist-tags"` // DistTags map dist tags to string versions
}

// MonthlyDownload holds the number of monthly downloads
// for an npm package.
type MonthlyDownload struct {
	Downloads uint `json:"downloads"`
}

// Exists determines if an npm package exists.
func Exists(name string) bool {
	resp, err := http.Get(util.GetProtocol() + "://registry.npmjs.org/" + name)
	util.Check(err)
	return resp.StatusCode == http.StatusOK
}

// GetMonthlyDownload uses the npm API to get the MonthlyDownload
// for a particular npm package.
func GetMonthlyDownload(name string) MonthlyDownload {
	resp, err := http.Get(util.GetProtocol() + "://api.npmjs.org/downloads/point/last-month/" + name)
	util.Check(err)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	util.Check(err)

	var counts MonthlyDownload
	util.Check(json.Unmarshal(body, &counts))
	return counts
}

// GetVersions gets all of the versions associated with an npm package,
// as well as the latest version based on the `latest` tag.
func GetVersions(ctx context.Context, config *packages.Autoupdate) ([]version.Version, *string) {
	name := *config.Target
	resp, err := http.Get(util.GetProtocol() + "://registry.npmjs.org/" + name)
	util.Check(err)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	util.Check(err)

	var r Registry
	util.Check(json.Unmarshal(body, &r))

	versions := make([]version.Version, 0)
	for k, v := range r.Versions {
		if v, ok := v.(map[string]interface{}); ok {
			dist := v["dist"].(map[string]interface{})
			tarball := dist["tarball"].(string)

			if timeInt, ok := r.TimeStamps[k]; ok {
				if timeStr, ok := timeInt.(string); ok {
					// parse time.Time from time stamp
					timeStamp, err := time.Parse(time.RFC3339, timeStr)
					util.Check(err)

					if !version.IsVersionIgnored(config, k) {
						versions = append(versions, version.Version{
							Version: k,
							Tarball: tarball,
							Date:    timeStamp,
							Source:  "npm",
						})
					} else {
						log.Printf("%s: version %s is ignored\n", name, k)
					}
					continue
				}
			}
			panic(fmt.Errorf("no time stamp for npm version %s/%s", name, k))
		}
	}

	// attempt to get latest version according to npm
	if latest, ok := r.DistTags["latest"]; ok {
		return versions, &latest
	}
	return versions, nil
}
