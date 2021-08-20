package force_update

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path"

	"github.com/cdnjs/tools/audit"
	"github.com/cdnjs/tools/gcp"
	"github.com/cdnjs/tools/git"
	"github.com/cdnjs/tools/npm"
	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/version"
)

type UpdateReq struct {
	Pkg     string `json:"package"`
	Version string `json:"version"`
}

func Invoke(w http.ResponseWriter, r *http.Request) {
	var d UpdateReq
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, "invalid request", 400)
		fmt.Println(err)
		return
	}

	list, err := packages.FetchPackages()
	if err != nil {
		http.Error(w, "failed to fetch packages", 500)
		fmt.Println(err)
		return
	}
	ctx := context.Background()

	for _, pkg := range list {
		if *pkg.Name == d.Pkg {
			src := *pkg.Autoupdate.Source
			var versions []version.Version
			switch src {
			case "git":
				versions, err = git.GetVersionsWithLimit(ctx, pkg.Autoupdate, 100)
				if err != nil {
					http.Error(w, "failed to fetch versions", 500)
					fmt.Println(err)
					return
				}
			case "npm":
				versions, _ = npm.GetVersions(ctx, pkg.Autoupdate)
			default:
				panic("unreachable")
			}

			var targetVersion *version.Version
			for _, version := range versions {
				if version.Version == d.Version {
					targetVersion = &version
					break
				}
			}

			if targetVersion == nil {
				var versionNames []string
				for _, version := range versions {
					versionNames = append(versionNames, version.Version)
				}
				msg := fmt.Sprintf("target version `%s` not found: %v", d.Version, versionNames)
				http.Error(w, msg, 500)
				return
			}
			tarball := version.DownloadTar(ctx, *targetVersion)
			if err := gcp.AddIncomingFile(path.Base(targetVersion.Tarball), tarball, pkg, *targetVersion); err != nil {
				log.Fatalf("could not store in GCS: %s", err)
			}
			if err := audit.NewVersionDetected(ctx, *pkg.Name, targetVersion.Version); err != nil {
				log.Fatalf("could not audit: %s", err)
			}

			return
		}
	}

	w.Write([]byte("OK"))
}
