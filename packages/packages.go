package packages

import (
	"context"
	"path"
	"sort"

	"github.com/cdnjs/tools/openssl"
	"github.com/cdnjs/tools/util"
)

var (
	BASE_PATH           = util.GetEnv("BOT_BASE_PATH")
	CDNJS_PATH          = path.Join(BASE_PATH, "cdnjs")
	CDNJS_PACKAGES_PATH = path.Join(BASE_PATH, "cdnjs", "ajax", "libs")
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
	Version     string
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
	return path.Join(CDNJS_PACKAGES_PATH, p.Name)
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

// List files that match the npm glob pattern in the `base` directory
// Returns a struct that represent the move semantics
func (p *Package) NpmFilesFrom(base string) []NpmFileMoveOp {
	out := make([]NpmFileMoveOp, 0)

	for _, fileMap := range p.NpmFileMap {
		for _, pattern := range fileMap.Files {
			basePath := path.Join(base, fileMap.BasePath)

			for _, f := range util.ListFilesGlob(basePath, pattern) {
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

	absPath := path.Join(CDNJS_PATH, version)
	out = append(out, util.ListFilesGlob(absPath, "**")...)

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
