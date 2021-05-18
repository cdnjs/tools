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
	"github.com/cdnjs/tools/npm"
	"github.com/cdnjs/tools/packages"
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
			npmVersions, _ := npm.GetVersions(ctx, *pkg.Autoupdate.Target)

			var targetVersion *npm.Version
			for _, version := range npmVersions {
				if version.Version == d.Version {
					targetVersion = &version
					break
				}
			}

			if targetVersion == nil {
				http.Error(w, "target version not found", 500)
				return
			}

			tarball := npm.DownloadTar(ctx, targetVersion.Tarball)
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
