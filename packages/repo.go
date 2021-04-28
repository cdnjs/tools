package packages

import (
	"archive/zip"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/pkg/errors"
)

const PACKAGES_ZIP = "https://github.com/cdnjs/packages/archive/refs/heads/master.zip"

func FetchPackages() ([]*Package, error) {
	zipfile, err := ioutil.TempFile("", "zip")
	if err != nil {
		return nil, errors.Wrap(err, "could not create temp file")
	}
	defer os.Remove(zipfile.Name())

	resp, err := http.Get(PACKAGES_ZIP)
	if err != nil {
		return nil, errors.Wrap(err, "could not fetch packages")
	}
	defer resp.Body.Close()
	_, err = io.Copy(zipfile, resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "could not download packages zip")
	}

	packages, err := inflatePackages(zipfile)
	if err != nil {
		return nil, errors.Wrap(err, "could not inflate packages")
	}
	return packages, nil
}

func GetRepoPackage(name string) (*Package, error) {
	packages, err := FetchPackages()
	if err != nil {
		return nil, errors.Wrap(err, "could not fetch packages")
	}
	for _, pkg := range packages {
		if *pkg.Name == name {
			return pkg, nil
		}
	}
	return nil, errors.New("package config not found")
}

func inflatePackages(src *os.File) ([]*Package, error) {
	var list []*Package

	r, err := zip.OpenReader(src.Name())
	if err != nil {
		return nil, err
	}
	defer r.Close()

	prefix := "packages-master/packages"
	// FIXME: pass from root
	ctx := context.Background()

	for _, f := range r.File {
		if strings.HasPrefix(f.Name, prefix) && strings.HasSuffix(f.Name, ".json") {
			reader, err := f.Open()
			if err != nil {
				return nil, errors.Wrap(err, "could open file")
			}
			bytes, err := ioutil.ReadAll(reader)
			if err != nil {
				return nil, errors.Wrap(err, "could not read file")
			}

			pkg, err := ReadHumanJSONBytes(ctx, f.Name, bytes)
			if err != nil {
				return nil, errors.Wrapf(err, "could not parse Package: %s", f.Name)
			}

			list = append(list, pkg)
		}
	}
	return list, nil
}
