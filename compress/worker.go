package compress

import (
	"context"
	"path"
	"sync"
)

type CompressJob struct {
	Ctx  context.Context
	File string
}

func Worker(wg *sync.WaitGroup, jobs <-chan CompressJob) {
	for j := range jobs {
		switch path.Ext(j.File) {
		case ".jpg", ".jpeg":
			Jpeg(j.Ctx, j.File)
		case ".png":
			Png(j.Ctx, j.File)
		case ".js":
			Js(j.Ctx, j.File)
		case ".css":
			CSS(j.Ctx, j.File)
		}
		wg.Done()
	}
}
