package main

import (
	"context"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/cdnjs/tools/compress"

	"github.com/pkg/errors"
)

const (
	INPUT  = "/input"
	OUTPUT = "/output"
)

func main() {
	ctx := context.Background()

	files, err := getInfiles()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(files)
	optimizeAndMinify(ctx, files)
	if err := outputFiles(files); err != nil {
		log.Fatalf("failed to output files: %s", err)
	}
}

func outputFiles(files []string) error {
	for _, file := range files {
		src := path.Join(INPUT, file)
		dest := path.Join(OUTPUT, file)
		log.Printf("%s -> %s\n", src, dest)

		if err := os.Rename(src, dest); err != nil {
			return errors.Wrap(err, "failed to rename")
		}
	}
	return nil
}

func getInfiles() ([]string, error) {
	var files []string

	err := filepath.Walk(INPUT, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list files")
	}

	return files, nil
}

// Optimizes/minifies files on disk for a particular package version.
func optimizeAndMinify(ctx context.Context, files []string) {
	cpuCount := runtime.NumCPU()
	jobs := make(chan compress.CompressJob, cpuCount)

	var wg sync.WaitGroup
	wg.Add(len(files))

	for w := 1; w <= cpuCount; w++ {
		go compress.Worker(&wg, jobs)
	}

	for _, file := range files {
		jobs <- compress.CompressJob{
			Ctx:  ctx,
			File: file,
		}
	}
	close(jobs)

	wg.Wait()
}
