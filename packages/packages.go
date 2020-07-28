package packages

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path"
	"sort"

	"github.com/cdnjs/tools/sri"
	"github.com/cdnjs/tools/util"
)

// Author represents an author.
type Author struct {
	Name  *string `json:"name"`
	Email *string `json:"email"`
	URL   *string `json:"url"`
}

// Autoupdate is used to update particular files from
// a source type located at a target destination.
type Autoupdate struct {
	Source  *string   `json:"source"`
	Target  *string   `json:"target"`
	FileMap []FileMap `json:"fileMap"`
}

// FileMap represents a number of files located
// under a base path.
type FileMap struct {
	BasePath *string  `json:"basePath"`
	Files    []string `json:"files"`
}

// Repository represents a repository.
type Repository struct {
	Type *string `json:"type"`
	URL  *string `json:"url"`
}

// Package holds metadata about a package.
// Its human-readable properties come from cdnjs/packages.
// The additional properties are used to manage the package.
// Any legacy properties are used to avoid any breaking changes in the API.
type Package struct {
	ctx      context.Context // context
	versions []string        // cache list of versions

	// human-readable properties
	Authors     []Author    `json:"authors"`
	Autoupdate  Autoupdate  `json:"autoupdate"`
	Description *string     `json:"description"`
	Filename    *string     `json:"filename"`
	Homepage    *string     `json:"homepage"`
	Keywords    []string    `json:"keywords"`
	License     *string     `json:"license"`
	Name        *string     `json:"name"`
	Repository  *Repository `json:"repository"`

	// additional properties
	Version *string `json:"version"`

	// legacy
	Author *string `json:"author"`
	// TODO: Remove this when we remove package.min.js generation
	Assets []Asset `json:"assets,omitempty"`
}

// Marshal marshals the package into JSON, not escaping HTML characters.
func (p *Package) Marshal() ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(p); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Path returns the location of the package in the cdnjs repo.
func (p *Package) LibraryPath() string {
	return path.Join(util.GetCDNJSLibrariesPath(), *p.Name)
}

// Versions gets the versions from git for a particular package.
func (p *Package) Versions() (versions []string) {
	if p.versions != nil {
		return p.versions
	}
	p.versions = GitListPackageVersions(p.ctx, p.LibraryPath())
	return p.versions
}

// CalculateVersionSRIs calculates SRIs for the files in
// a particular package version.
func (p *Package) CalculateVersionSRIs(version string) map[string]string {
	sriFileMap := make(map[string]string)

	for _, relFile := range p.AllFiles(version) {
		if path.Ext(relFile) == ".js" || path.Ext(relFile) == ".css" {
			absFile := path.Join(p.LibraryPath(), version, relFile)
			sriFileMap[relFile] = sri.CalculateFileSRI(absFile)
		}
	}

	return sriFileMap
}

// TODO: Remove when no longer writing files to disk, and all files can
// be removed after temporarily processed and uploaded to KV.

// NpmFileMoveOp represents an operation to move files
// from a source destination to a target destination.
type NpmFileMoveOp struct {
	From string
	To   string
}

// NpmFilesFrom lists files that match the npm glob pattern in the `base` directory
// Returns a struct that represent the move semantics
func (p *Package) NpmFilesFrom(base string) []NpmFileMoveOp {
	out := make([]NpmFileMoveOp, 0)

	// map used to determine if a file path has already been processed
	seen := make(map[string]bool)

	for _, fileMap := range p.Autoupdate.FileMap {
		for _, pattern := range fileMap.Files {
			basePath := path.Join(base, *fileMap.BasePath)

			// find files that match glob
			list, err := util.ListFilesGlob(p.ctx, basePath, pattern)
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
					util.Warnf(p.ctx, "stat: "+staterr.Error())
					continue
				}

				// warn for files with sizes exceeding max file size
				size := info.Size()
				if size > util.MaxFileSize {
					util.Warnf(p.ctx, "file %s ignored due to byte size (%d > %d)", f, size, util.MaxFileSize)
					continue
				}

				// file is ok
				out = append(out, NpmFileMoveOp{
					From: path.Join(*fileMap.BasePath, f),
					To:   f,
				})
			}
		}
	}

	return out
}

// AllFiles lists all files in the version directory.
func (p *Package) AllFiles(version string) []string {
	out := make([]string, 0)

	absPath := path.Join(p.LibraryPath(), version)
	list, err := util.ListFilesInVersion(p.ctx, absPath)
	util.Check(err)
	out = append(out, list...)

	return out
}

// Asset associates a number of files as strings
// with a version.
type Asset struct {
	Version string   `json:"version"`
	Files   []string `json:"files"`
}

// GetAssets gets all the assets for a particular package.
func (p *Package) GetAssets() []Asset {
	assets := make([]Asset, 0)

	for _, version := range p.Versions() {
		files := p.AllFiles(version)

		assets = append(assets, Asset{
			Version: version,
			Files:   files,
		})
	}

	// Sort by version
	sort.Sort(ByVersionAsset(assets))

	return assets
}
