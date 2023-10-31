package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cdnjs/tools/packages"

	"github.com/pkg/errors"
)

var (
	// these file extensions will be uploaded to KV
	// but not compressed
	doNotCompress = map[string]bool{
		".woff2": true,
		".sri":   true, // internal SRI hash file
	}
	// we calculate SRIs for these file extensions
	calculateSRI = map[string]bool{
		".js":  true,
		".css": true,
	}
)

const (
	INPUT   = "/input"
	OUTPUT  = "/output"
	PACKAGE = "/tmp/pkg"
)

func main() {
	ctx := context.Background()

	config, err := readConfig()
	if err != nil {
		log.Fatalf("could not read config: %s", err)
	}

	if err := os.MkdirAll(PACKAGE, 0700); err != nil {
		log.Fatalf("could not create PACKAGE: %s", err)
	}

	if err := extractInput(*config.Autoupdate.Source); err != nil {
		log.Fatalf("failed to extract input: %s", err)
	}

	// Step 1. copy all package files to their destination according to the
	// fileMap configuration.
	if err := copyPackage(ctx, config); err != nil {
		log.Fatalf("failed to optimize files: %s", err)
	}
	// Step 2. iterate over the last output and minify files
	if err := optimizePackage(ctx, config); err != nil {
		log.Fatalf("failed to optimize files: %s", err)
	}
	// Step 3. iterate over the last output and calculate SRIs for each files
	if err := calcSriPackage(ctx, config); err != nil {
		log.Fatalf("failed to optimize files: %s", err)
	}
	// Step 4. iterate over the last output and compress all files
	if err := compressPackage(ctx, config); err != nil {
		log.Fatalf("failed to optimize files: %s", err)
	}
	log.Printf("processed %s\n", *config.Name)
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
			if err := os.MkdirAll(path.Join(PACKAGE, filepath.Dir(target)), 0755); err != nil {
				return errors.Wrap(err, "ExtractTarGz: Mkdir() failed")
			}
			outFile, err := os.Create(path.Join(PACKAGE, target))
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

// Copy the files package to their inteded location
func copyPackage(ctx context.Context, config *packages.Package) error {
	files := config.NpmFilesFrom(PACKAGE)

	for _, file := range files {
		src := path.Join(PACKAGE, file.From)
		dest := path.Join(OUTPUT, file.To)

		if err := copyFile(src, dest); err != nil {
			log.Fatalf("failed to copy file: %s", err)
		}
		log.Printf("copy %s -> %s\n", src, dest)
	}

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
