package check_pkg_updates

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cdnjs/tools/kv"
	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/sentry"
	"github.com/cdnjs/tools/util"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
)

var (
	KV_TOKEN              = os.Getenv("KV_TOKEN")
	CF_ACCOUNT_ID         = os.Getenv("CF_ACCOUNT_ID")
	PKG_AUTOUPDATE_SOURCE = os.Getenv("PKG_AUTOUPDATE_SOURCE")
	RESTRICT_PKGS         = strings.Split(os.Getenv("RESTRICT_PKGS"), ",")
)

type APIPackage struct {
	Versions []string `json:"versions"`
}

func getExistingVersions(p *packages.Package) ([]string, error) {
	cfapi, err := cloudflare.NewWithAPIToken(KV_TOKEN, cloudflare.UsingAccount(CF_ACCOUNT_ID))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloudflare API client")
	}

	versions, err := kv.GetVersions(cfapi, *p.Name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get verions")
	}

	return versions, nil
}

func Invoke(w http.ResponseWriter, r *http.Request) {
	sentry.Init()
	defer sentry.PanicHandler()

	if PKG_AUTOUPDATE_SOURCE == "" {
		panic("PKG_AUTOUPDATE_SOURCE should be present")
	}

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
	if os.Getenv("RESTRICT_PKGS") == "" {
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

	src := *pkg.Autoupdate.Source
	if src != PKG_AUTOUPDATE_SOURCE {
		// we are not auto-updateing packages with that source; skip.
		return nil
	}

	switch src {
	case "npm", "git":
		{
			if err := updatePackage(ctx, pkg, src); err != nil {
				return errors.Wrap(err, "failed to update package via "+src)
			}
		}
	default:
		{
			return errors.Errorf("%s invalid autoupdate source: %s", *pkg.Name, src)
		}
	}
	return nil
}
