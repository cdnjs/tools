package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/cdnjs/tools/compress"
	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/sri"

	"github.com/pkg/errors"
)

var (
	// these file extensions will be uploaded to KV
	// but not compressed
	doNotCompress = map[string]bool{
		".woff2": true,
	}
	// we calculate SRIs for these file extensions
	calculateSRI = map[string]bool{
		".js":  true,
		".css": true,
	}
)

const (
	INPUT     = "/input"
	OUTPUT    = "/output"
	WORKSPACE = "/tmp/work"
)

func main() {
	ctx := context.Background()

	config, err := readConfig()
	if err != nil {
		log.Fatalf("could not read config: %s", err)
	}

	if err := os.MkdirAll(WORKSPACE, 0700); err != nil {
		log.Fatalf("could not create workspace: %s", err)
	}

	if err := extractInput(); err != nil {
		log.Fatalf("failed to extract input: %s", err)
	}

	files := config.NpmFilesFrom(WORKSPACE)
	log.Println("input files", files)

	if err := optimizeFiles(ctx, files); err != nil {
		log.Fatalf("failed to optimize files: %s", err)
	}
	log.Printf("processed %s\n", *config.Name)
}

type optimizeJob struct {
	Ctx  context.Context
	File string
	Dest string
}

func (j optimizeJob) emit(name string) {
	src := path.Join(WORKSPACE, name)
	dest := path.Join(OUTPUT, j.Dest)
	log.Printf("%s -> %s\n", src, dest)

	if err := copyFile(src, dest); err != nil {
		log.Fatalf("failed to copy file: %s", err)
	}
}

func removePackageDir(path string) string {
	if len(path) < 8 {
		return path
	}
	if path[0:8] == "package/" {
		return path[8:]
	}
	return path
}

func readConfig() (*packages.Package, error) {
	file := path.Join(INPUT, "config.json")
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrap(err, "could not read file")
	}
	log.Println("config sandbox", string(data))
	config := new(packages.Package)
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrap(err, "could not parse config")
	}
	return config, nil
}

func extractInput() error {
	gzipStream, err := os.Open(path.Join(INPUT, "new-version.tgz"))
	if err != nil {
		return errors.Wrap(err, "could not open input")
	}

	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return errors.Wrap(err, "could not create reader")
	}

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("ExtractTarGz: Next() failed: %s", err.Error())
		}

		// remove package folder
		target := removePackageDir(header.Name)

		switch header.Typeflag {
		case tar.TypeReg:
			if err := os.MkdirAll(path.Join(WORKSPACE, filepath.Dir(target)), 0755); err != nil {
				return errors.Wrap(err, "ExtractTarGz: Mkdir() failed")
			}
			outFile, err := os.Create(path.Join(WORKSPACE, target))
			if err != nil {
				return errors.Wrap(err, "ExtractTarGz: Create() failed")
			}
			defer outFile.Close()
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return errors.Wrap(err, "ExtractTarGz: Copy() failed")
			}
		default:
			return errors.Errorf(
				"ExtractTarGz: uknown type: %x in %s",
				header.Typeflag,
				header.Name)
		}
	}
	return nil
}

func optimizeWorker(wg *sync.WaitGroup, jobs <-chan optimizeJob) {
	for j := range jobs {
		intputFile := path.Join(WORKSPACE, j.File)
		ext := path.Ext(j.File)
		switch ext {
		case ".jpg", ".jpeg":
			compress.Jpeg(j.Ctx, intputFile)
		case ".png":
			compress.Png(j.Ctx, intputFile)
		case ".js":
			compress.Js(j.Ctx, intputFile)
		case ".css":
			compress.CSS(j.Ctx, intputFile)
		}

		if _, ok := calculateSRI[ext]; ok {
			outSRI := fmt.Sprintf("%s.sri", j.File)
			sri.CalculateFileSRI(intputFile, path.Join(WORKSPACE, outSRI))
			j.emit(outSRI)
		}

		if _, ok := doNotCompress[ext]; !ok {
			outBr := fmt.Sprintf("%s.br", j.File)
			compress.Brotli11CLI(j.Ctx, intputFile, path.Join(WORKSPACE, outBr))

			outGz := fmt.Sprintf("%s.gz", j.File)
			compress.Gzip9Native(j.Ctx, intputFile, path.Join(WORKSPACE, outGz))

			j.emit(outBr)
			j.emit(outGz)
		} else {
			j.emit(j.File)
		}
		wg.Done()
	}
}

// Optimizes/minifies files on disk for a particular package version.
func optimizeFiles(ctx context.Context, files []packages.NpmFileMoveOp) error {
	cpuCount := runtime.NumCPU()
	jobs := make(chan optimizeJob, cpuCount)

	var wg sync.WaitGroup
	wg.Add(len(files))

	log.Printf("spanwing %d workers\n", cpuCount)
	for w := 1; w <= cpuCount; w++ {
		go optimizeWorker(&wg, jobs)
	}

	for _, file := range files {
		jobs <- optimizeJob{
			Ctx:  ctx,
			File: file.From,
			Dest: file.To,
		}
	}
	close(jobs)

	wg.Wait()
	return nil
}

func copyFile(src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return errors.Wrap(err, "could not create dir")
	}

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
