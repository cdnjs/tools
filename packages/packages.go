package packages

import (
	"context"
	"os"
	"path"
	"sort"

	"github.com/cdnjs/tools/openssl"
	"github.com/cdnjs/tools/util"
)

type Repository struct {
	Repotype string `json:"type"`
	Url      string `json:"url"`
}

type Author struct {
	Name  string  `json:"name"`
	Email string  `json:"email"`
	Url   *string `json:"url,omitempty"`
}

type License struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type FileMap struct {
	BasePath string   `json:"basePath"`
	Files    []string `json:"files"`
}

type Autoupdate struct {
	Source  string     `json:"source"`
	Target  string     `json:"target"`
	FileMap *[]FileMap `json:"fileMap,omitempty"`
}

type Asset struct {
	Version string   `json:"version"`
	Files   []string `json:"files"`
}

type Package struct {
	ctx context.Context
	// Cache list of versions for the package
	versions []string

	Title       string
	Name        string
	Description string
	Version     *string
	Author      Author
	Homepage    string
	Keywords    []string
	Repository  Repository
	Filename    string
	NpmName     *string
	NpmFileMap  []FileMap
	License     *License
	Autoupdate  *Autoupdate
}

func stringInObject(key string, object map[string]interface{}) string {
	value := object[key]
	if str, ok := value.(string); ok {
		return str
	} else {
		return ""
	}
}

// Location of the package in the cdnjs repo
func (p *Package) Path() string {
	return path.Join(util.GetCDNJSPackages(), p.Name)
}

func (p *Package) Versions() (versions []string) {
	if p.versions != nil {
		return p.versions
	}
	p.versions = GitListPackageVersions(p.ctx, p.Path())
	return p.versions
}

func (p *Package) CalculateVersionSris(version string) map[string]string {
	sriFileMap := make(map[string]string)

	for _, relFile := range p.AllFiles(version) {
		if path.Ext(relFile) == ".js" || path.Ext(relFile) == ".css" {
			absFile := path.Join(p.Path(), version, relFile)
			sriFileMap[relFile] = openssl.CalculateFileSri(absFile)
		}
	}

	return sriFileMap
}

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

	for _, fileMap := range p.NpmFileMap {
		for _, pattern := range fileMap.Files {
			basePath := path.Join(base, fileMap.BasePath)

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
				if size > util.MAX_FILE_SIZE {
					util.Warnf(p.ctx, "file %s ignored due to byte size (%d > %d)", f, size, util.MAX_FILE_SIZE)
					continue
				}

				// file is ok
				out = append(out, NpmFileMoveOp{
					From: path.Join(fileMap.BasePath, f),
					To:   f,
				})
			}
		}
	}

	return out
}

// List all files in the version directory
func (p *Package) AllFiles(version string) []string {
	out := make([]string, 0)

	absPath := path.Join(p.Path(), version)
	list, err := util.ListFilesInVersion(p.ctx, absPath)
	util.Check(err)
	out = append(out, list...)

	return out
}

func (p *Package) Assets() []Asset {
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
