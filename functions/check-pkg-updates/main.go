package check_pkg_updates

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/sentry"
	"github.com/cdnjs/tools/util"

	"github.com/pkg/errors"
)

var httpclient = &http.Client{
	Timeout: 10 * time.Second,
}

type version interface {
	Get() string             // Get the version.
	GetTimeStamp() time.Time // GetTimeStamp gets the time stamp associated with the version.
}

var (
	RESTRICT_PKGS = strings.Split(os.Getenv("RESTRICT_PKGS"), ",")
)

type APIPackage struct {
	Versions []string `json:"versions"`
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

func getExistingVersions(p *packages.Package) ([]string, error) {
	resp, err := httpclient.Get(fmt.Sprintf("%s/libraries/%s?fields=versions", util.GetCdnjsAPI(), *p.Name))
	if err != nil {
		return nil, errors.Wrap(err, "could not get existing versions")
	}

	defer resp.Body.Close()

	// package is not known by the API yet
	if resp.StatusCode == 404 {
		return make([]string, 0), nil
	}

	if resp.StatusCode != 200 {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "could not read response body")
		}
		return nil, errors.Errorf("API returned %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var target APIPackage
	if err := json.NewDecoder(resp.Body).Decode(&target); err != nil {
		return nil, errors.Wrap(err, "could not parse API response")
	}
	return target.Versions, nil
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

	// shuffle package order
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(list), func(i, j int) { list[i], list[j] = list[j], list[i] })

	for _, pkg := range list {
		if err := checkPackage(pkg); err != nil {
			log.Printf("failed to update package %s: %s", *pkg.Name, err)
		}
	}

	fmt.Fprint(w, "OK")
}

func isAllowed(pkg string) bool {
	if len(RESTRICT_PKGS) == 0 {
		return true
	}
	for _, n := range RESTRICT_PKGS {
		if pkg == n {
			return true
		}
	}
	return false
}

func checkPackage(pkg *packages.Package) error {
	if !isAllowed(*pkg.Name) {
		return nil
	}
	logger := util.GetStandardLogger()
	ctx := util.ContextWithEntries(
		util.GetStandardEntries(*pkg.Name, logger)...)

	if pkg.Autoupdate == nil {
		// package not configured to auto update; skip.
		return nil
	}

	switch *pkg.Autoupdate.Source {
	case "npm":
		{
			if err := updateNpm(ctx, pkg); err != nil {
				return errors.Wrap(err, "failed to update package via npm")
			}
		}
	// case "git":
	// 	{
	// 		if err := updateGit(ctx, pkg); err != nil {
	// 			return errors.Wrap(err, "failed to update package via git")
	// 		}
	// 	}
	default:
		{
			return errors.Errorf("%s invalid autoupdate source: %s", *pkg.Name, *pkg.Autoupdate.Source)
		}
	}
	return nil
}
