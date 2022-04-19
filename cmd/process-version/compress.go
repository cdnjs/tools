package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"sync"

	"github.com/cdnjs/tools/compress"
	"github.com/cdnjs/tools/packages"
	"github.com/pkg/errors"
)

func compressPackage(ctx context.Context, config *packages.Package) error {
	files, err := ioutil.ReadDir(OUTPUT)
	if err != nil {
		return errors.Wrap(err, "failed to list output files")
	}
	cpuCount := runtime.NumCPU()
	jobs := make(chan compressionJob, cpuCount)

	var wg sync.WaitGroup

	for w := 1; w <= cpuCount; w++ {
		go compressionWorker(&wg, jobs)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		wg.Add(1)
		jobs <- compressionJob{
			Ctx:  ctx,
			File: file.Name(),
		}
	}
	close(jobs)

	wg.Wait()
	return nil
}

type compressionJob struct {
	Ctx  context.Context
	File string
}

func compressionWorker(wg *sync.WaitGroup, jobs <-chan compressionJob) {
	for j := range jobs {
		src := path.Join(OUTPUT, j.File)
		ext := path.Ext(src)

		if _, ok := doNotCompress[ext]; !ok {
			outBr := fmt.Sprintf("%s.br", src)
			if _, err := os.Stat(outBr); err == nil {
				log.Printf("file %s already exists at the output\n", outBr)
			} else {
				compress.Brotli11CLI(j.Ctx, src, outBr)
				log.Printf("br %s -> %s\n", src, outBr)
			}

			outGz := fmt.Sprintf("%s.gz", src)
			if _, err := os.Stat(outGz); err == nil {
				log.Printf("file %s already exists at the output\n", outGz)
			} else {
				compress.Gzip9Native(j.Ctx, src, outGz)
				log.Printf("gz %s -> %s\n", src, outGz)
			}

			// Original file can be removed because we keep the compressed
			// version
			os.Remove(src)
		}

		wg.Done()
	}
}
