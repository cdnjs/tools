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
	"runtime"
	"sync"

	"github.com/cdnjs/tools/compress"
	"github.com/cdnjs/tools/sri"
	"github.com/cdnjs/tools/util"

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

// FIXME: share with process-version-host
type Config struct {
	Filemap []FileMap `json:"filemap"`
}

// FIXME: share with packages
type FileMap struct {
	BasePath *string  `json:"basePath"` // can be empty
	Files    []string `json:"files,omitempty"`
}

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

	files, err := getInfiles(ctx, config.Filemap)
	if err != nil {
		log.Fatalf("could not list input files: %s", err)
	}
	log.Println("input files", files)

	if err := optimizeFiles(ctx, files); err != nil {
		log.Fatalf("failed to optimize files: %s", err)
	}
}

type optimizeJob struct {
	Ctx  context.Context
	File string
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

func readConfig() (*Config, error) {
	file := path.Join(INPUT, "config.json")
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrap(err, "could not read file")
	}
	log.Println("config", string(data))
	config := new(Config)
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

		// remove p
		target := removePackageDir(header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(path.Join(WORKSPACE, target), 0755); err != nil {
				log.Fatalf("ExtractTarGz: Mkdir() failed: %s", err.Error())
			}
		case tar.TypeReg:
			outFile, err := os.Create(path.Join(WORKSPACE, target))
			if err != nil {
				log.Fatalf("ExtractTarGz: Create() failed: %s", err.Error())
			}
			defer outFile.Close()
			if _, err := io.Copy(outFile, tarReader); err != nil {
				log.Fatalf("ExtractTarGz: Copy() failed: %s", err.Error())
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

func optimizeWorker(wg *sync.WaitGroup, jobs <-chan optimizeJob, emit func(string)) {
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

		if _, ok := doNotCompress[ext]; !ok {
			outBr := fmt.Sprintf("%s.br", j.File)
			compress.Brotli11CLI(j.Ctx, intputFile, path.Join(WORKSPACE, outBr))

			outGz := fmt.Sprintf("%s.gz", j.File)
			compress.Gzip9Native(j.Ctx, intputFile, path.Join(WORKSPACE, outGz))

			outSRI := fmt.Sprintf("%s.sri", j.File)
			sri.CalculateFileSRI(intputFile, path.Join(WORKSPACE, outSRI))

			emit(outBr)
			emit(outGz)
			emit(outSRI)
		} else {
			emit(j.File)
		}
		wg.Done()
	}
}

func getInfiles(ctx context.Context, filemap []FileMap) ([]string, error) {
	files := make([]string, 0)

	// map used to determine if a file path has already been processed
	seen := make(map[string]bool)

	for _, fileMap := range filemap {
		for _, pattern := range fileMap.Files {
			basePath := path.Join(WORKSPACE, *fileMap.BasePath)

			// find files that match glob
			list, err := util.ListFilesGlob(ctx, basePath, pattern)
			util.Check(err) // should have already run before in checker so panic if glob invalid

			for _, f := range list {
				fp := path.Join(basePath, f)

				// check if file has been processed before
				if _, ok := seen[fp]; ok {
					continue
				}
				seen[fp] = true

				info, staterr := os.Stat(fp)
				if staterr != nil {
					log.Println("stat: " + staterr.Error())
					continue
				}

				// warn for files with sizes exceeding max file size
				size := info.Size()
				if size > util.MaxFileSize {
					log.Printf("file %s ignored due to byte size (%d > %d)\n", f, size, util.MaxFileSize)
					continue
				}

				files = append(files, f)
			}
		}
	}

	return files, nil
}

func emit(name string) {
	src := path.Join(WORKSPACE, name)
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
