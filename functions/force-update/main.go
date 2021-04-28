package force_update

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path"

	"github.com/cdnjs/tools/audit"
	"github.com/cdnjs/tools/gcp"
	"github.com/cdnjs/tools/npm"
	"github.com/cdnjs/tools/packages"
)

const (
	PKG     = "jquery"
	VERSION = "3.6.0"
)

func Invoke(w http.ResponseWriter, r *http.Request) {
	list, err := packages.FetchPackages()
	if err != nil {
		http.Error(w, "failed to fetch packages", 500)
		fmt.Println(err)
		return
	}
	ctx := context.Background()

	for _, pkg := range list {
		if *pkg.Name == PKG {
			npmVersions, _ := npm.GetVersions(ctx, *pkg.Autoupdate.Target)

			var targetVersion *npm.Version
			for _, version := range npmVersions {
				if version.Version == VERSION {
					targetVersion = &version
					break
				}
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
}
