package main

import (
	"context"
	"io/ioutil"
	"log"
	"path"
	"runtime"
	"strings"
	"sync"

	"github.com/cdnjs/tools/compress"
	"github.com/cdnjs/tools/packages"
	"github.com/pkg/errors"
)

// Optimizes/minifies package's files on disk for a particular package version.
func optimizePackage(ctx context.Context, config *packages.Package) error {
	log.Printf("optimizing files (Js %t, Css %t, Png %t, Jpg %t)\n",
		config.Optimization.Js(),
		config.Optimization.Css(),
		config.Optimization.Png(),
		config.Optimization.Jpg())

	files, err := ioutil.ReadDir(OUTPUT)
	if err != nil {
		return errors.Wrap(err, "failed to list output files")
	}
	cpuCount := runtime.NumCPU()
	jobs := make(chan optimizeJob, cpuCount)

	var wg sync.WaitGroup

	for w := 1; w <= cpuCount; w++ {
		go optimizeWorker(&wg, jobs)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		wg.Add(1)
		jobs <- optimizeJob{
			Ctx:          ctx,
			Optimization: config.Optimization,
			File:         file.Name(),
		}
	}
	close(jobs)

	wg.Wait()
	return nil
}

type optimizeJob struct {
	Ctx          context.Context
	Optimization *packages.Optimization
	File         string
}

func (j optimizeJob) emit(name string) {
	src := path.Join(OUTPUT, j.File)
	dest := path.Join(OUTPUT, name)
	if err := copyFile(src, dest); err != nil {
		log.Fatalf("failed to copy file: %s", err)
	}
}

func optimizeWorker(wg *sync.WaitGroup, jobs <-chan optimizeJob) {
	for j := range jobs {
		src := path.Join(OUTPUT, j.File)
		ext := path.Ext(src)
		switch ext {
		case ".jpg", ".jpeg":
			if j.Optimization.Jpg() {
				compress.Jpeg(j.Ctx, src) // replaces in-place
			}
		case ".png":
			if j.Optimization.Png() {
				compress.Png(j.Ctx, src) // replaces in-place
			}
		case ".js":
			if j.Optimization.Js() {
				if out := compress.Js(j.Ctx, src); out != nil {
					out := strings.Replace(src, ".js", ".min.js", 1)
					j.emit(out)
				}
			}
		case ".css":
			if j.Optimization.Css() {
				if out := compress.CSS(j.Ctx, src); out != nil {
					out := strings.Replace(src, ".css", ".min.css", 1)
					j.emit(out)
				}
			}
		}

		wg.Done()
	}
}
