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
	"strings"
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

	if err := extractInput(*config.Autoupdate.Source); err != nil {
		log.Fatalf("failed to extract input: %s", err)
	}

	if err := optimizePackage(ctx, config); err != nil {
		log.Fatalf("failed to optimize files: %s", err)
	}
	log.Printf("processed %s\n", *config.Name)
}

type optimizeJob struct {
	Ctx          context.Context
	Optimization *packages.Optimization
	File         string
	Dest         string
}

func (j optimizeJob) clone() optimizeJob {
	return optimizeJob{
		Ctx:          j.Ctx,
		Optimization: j.Optimization,
		File:         j.File,
		Dest:         j.Dest,
	}
}

func (j optimizeJob) emitFromWorkspace(src string) {
	dest := path.Join(OUTPUT, j.Dest)
	if err := os.MkdirAll(path.Dir(dest), 0755); err != nil {
		log.Fatalf("could not create dest dir: %s", err)
	}

	ext := path.Ext(src)
	if _, ok := calculateSRI[ext]; ok {
		outSRI := fmt.Sprintf("%s.sri", dest)
		sri.CalculateFileSRI(src, outSRI)
		log.Printf("sri %s -> %s\n", src, outSRI)
	}

	if _, ok := doNotCompress[ext]; !ok {
		outBr := fmt.Sprintf("%s.br", dest)
		if _, err := os.Stat(outBr); err == nil {
			log.Printf("file %s already exists at the output\n", outBr)
		} else {
			compress.Brotli11CLI(j.Ctx, src, outBr)
			log.Printf("br %s -> %s\n", src, outBr)
		}

		outGz := fmt.Sprintf("%s.gz", dest)
		if _, err := os.Stat(outGz); err == nil {
			log.Printf("file %s already exists at the output\n", outGz)
		} else {
			compress.Gzip9Native(j.Ctx, src, outGz)
			log.Printf("gz %s -> %s\n", src, outGz)
		}
	} else {
		if err := copyFile(src, dest); err != nil {
			log.Fatalf("failed to copy file: %s", err)
		}
		log.Printf("copy %s -> %s\n", src, dest)
	}
}

func (j optimizeJob) emit(name string) {
	src := path.Join(WORKSPACE, name)
	j.emitFromWorkspace(src)
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

func removeFirstDir(path string) string {
	parts := strings.Split(path, "/")
	return strings.Replace(path, parts[0]+"/", "", 1)
}

func readConfig() (*packages.Package, error) {
	file := path.Join(INPUT, "config.json")
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrap(err, "could not read file")
	}
	config := new(packages.Package)
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrap(err, "could not parse config")
	}
	return config, nil
}

func extractInput(source string) error {
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

		target := header.Name
		if source == "npm" {
			// remove package folder
			target = removePackageDir(header.Name)
		}
		if source == "git" {
			// remove package folder
			target = removeFirstDir(header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// ignore dirs
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
			log.Printf(
				"ExtractTarGz: uknown type: %x in %s\n",
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
			if j.Optimization.Jpg() {
				compress.Jpeg(j.Ctx, intputFile)
			}
		case ".png":
			if j.Optimization.Png() {
				compress.Png(j.Ctx, intputFile)
			}
		case ".js":
			if j.Optimization.Js() {
				if out := compress.Js(j.Ctx, intputFile); out != nil {
					j := j.clone()
					j.Dest = strings.Replace(j.Dest, ".js", ".min.js", 1)
					j.emitFromWorkspace(*out)
				}
			}
		case ".css":
			if j.Optimization.Css() {
				if out := compress.CSS(j.Ctx, intputFile); out != nil {
					j := j.clone()
					j.Dest = strings.Replace(j.Dest, ".css", ".min.css", 1)
					j.emitFromWorkspace(*out)
				}
			}
		}

		j.emit(j.File)
		wg.Done()
	}
}

// Optimizes/minifies package's files on disk for a particular package version.
func optimizePackage(ctx context.Context, config *packages.Package) error {
	log.Printf("optimizing files (Js %t, Css %t, Png %t, Jpg %t)\n",
		config.Optimization.Js(),
		config.Optimization.Css(),
		config.Optimization.Png(),
		config.Optimization.Jpg())

	files := config.NpmFilesFrom(WORKSPACE)
	cpuCount := runtime.NumCPU()
	jobs := make(chan optimizeJob, cpuCount)

	var wg sync.WaitGroup
	wg.Add(len(files))

	for w := 1; w <= cpuCount; w++ {
		go optimizeWorker(&wg, jobs)
	}

	for _, file := range files {
		jobs <- optimizeJob{
			Ctx:          ctx,
			Optimization: config.Optimization,
			File:         file.From,
			Dest:         file.To,
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
