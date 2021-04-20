package compress

import (
	"context"
	"path"
	"sync"

	"github.com/cdnjs/tools/packages"
)

type CompressJob struct {
	Ctx         context.Context
	File        string
	VersionPath string
}

func Worker(wg *sync.WaitGroup, jobs <-chan CompressJob, optim *packages.Optimization) {
	defaultCompress := optim == nil // compress by default if optimization not specified
	for j := range jobs {
		switch path.Ext(j.File) {
		case ".jpg", ".jpeg":
			if defaultCompress || optim.Jpg() {
				Jpeg(j.Ctx, path.Join(j.VersionPath, j.File))
			}
		case ".png":
			if defaultCompress || optim.Png() {
				Png(j.Ctx, path.Join(j.VersionPath, j.File))
			}
		case ".js":
			if defaultCompress || optim.Js() {
				Js(j.Ctx, path.Join(j.VersionPath, j.File))
			}
		case ".css":
			if defaultCompress || optim.Css() {
				CSS(j.Ctx, path.Join(j.VersionPath, j.File))
			}
		}
		wg.Done()
	}
}
