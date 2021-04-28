package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
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

	if err := optimizeFiles(ctx, files); err != nil {
		log.Fatalf("failed to optimize files: %s", err)
	}
}

type optimizeJob struct {
	Ctx  context.Context
	File string
}

func optimizeWorker(wg *sync.WaitGroup, jobs <-chan optimizeJob, emit func(string)) {
	for j := range jobs {
		intputFile := path.Join(INPUT, j.File)
		switch path.Ext(j.File) {
		case ".jpg", ".jpeg":
			compress.Jpeg(j.Ctx, intputFile)
		case ".png":
			compress.Png(j.Ctx, intputFile)
		case ".js":
			compress.Js(j.Ctx, intputFile)
		case ".css":
			compress.CSS(j.Ctx, intputFile)
		}

		outBr := fmt.Sprintf("%s.br", j.File)
		compress.Brotli11CLI(j.Ctx, intputFile, path.Join(INPUT, outBr))

		outGz := fmt.Sprintf("%s.gz", j.File)
		compress.Gzip9Native(j.Ctx, intputFile, path.Join(INPUT, outGz))

		emit(outBr)
		emit(outGz)
		wg.Done()
	}
}

func getInfiles() ([]string, error) {
	var files []string

	err := filepath.Walk(INPUT, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, strings.ReplaceAll(path, INPUT, ""))
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list files")
	}

	return files, nil
}

func emit(name string) {
	src := path.Join(INPUT, name)
	dest := path.Join(OUTPUT, name)
	log.Printf("%s -> %s\n", src, dest)

	if err := copyFile(src, dest); err != nil {
		log.Fatalf("failed to copy file: %s", err)
	}
}

// Optimizes/minifies files on disk for a particular package version.
func optimizeFiles(ctx context.Context, files []string) error {
	cpuCount := runtime.NumCPU()
	jobs := make(chan optimizeJob, cpuCount)

	var wg sync.WaitGroup
	wg.Add(len(files))

	log.Printf("spanwing %d workers\n", cpuCount)
	for w := 1; w <= cpuCount; w++ {
		go optimizeWorker(&wg, jobs, emit)
	}

	for _, file := range files {
		jobs <- optimizeJob{
			Ctx:  ctx,
			File: file,
		}
	}
	close(jobs)

	wg.Wait()
	return nil
}

func copyFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return errors.Wrap(err, "could not open source file")
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return errors.Wrap(err, "could not open dest file")
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile) // check first var for number of bytes copied
	if err != nil {
		return errors.Wrap(err, "could not copy")
	}

	err = destFile.Sync()
	if err != nil {
		return errors.Wrap(err, "could not sync")
	}
	return nil
}
