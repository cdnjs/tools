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
	// compress everything by default
	jpg, png, js, css := true, true, true, true
	if optim != nil {
		jpg = optim.JPG == nil || *optim.JPG
		png = optim.PNG == nil || *optim.PNG
		js = optim.JS == nil || *optim.JS
		css = optim.CSS == nil || *optim.CSS
	}
	for j := range jobs {
		switch path.Ext(j.File) {
		case ".jpg", ".jpeg":
			if jpg {
				Jpeg(j.Ctx, path.Join(j.VersionPath, j.File))
			}
		case ".png":
			if png {
				Png(j.Ctx, path.Join(j.VersionPath, j.File))
			}
		case ".js":
			if js {
				Js(j.Ctx, path.Join(j.VersionPath, j.File))
			}
		case ".css":
			if css {
				CSS(j.Ctx, path.Join(j.VersionPath, j.File))
			}
		}
		wg.Done()
	}
}
