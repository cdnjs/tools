package packages

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"sort"

	"github.com/cdnjs/tools/openssl"
	"github.com/cdnjs/tools/util"

	"github.com/pkg/errors"
)

const (
	PACKAGES_PATH = "ajax/libs"
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

func (p *Package) path() string {
	return path.Join(PACKAGES_PATH, p.Name)
}

func (p *Package) Versions() (versions []string) {
	if p.versions != nil {
		return p.versions
	}
	p.versions = GitListPackageVersions(p.ctx, p.path())
	return p.versions
}

func (p *Package) CalculateVersionSris(version string) map[string]string {
	sriFileMap := make(map[string]string)

	for _, relFile := range p.files(version) {
		if path.Ext(relFile) == ".js" || path.Ext(relFile) == ".css" {
			absFile := path.Join(p.path(), version, relFile)
			sriFileMap[relFile] = openssl.CalculateFileSri(absFile)
		}
	}

	return sriFileMap
}

func (p *Package) files(version string) []string {
	out := make([]string, 0)

	basePath := path.Join(p.path(), version)
	out = append(out, util.ListFilesGlob(basePath, "**")...)

	return out
}

func (p *Package) Assets() []Asset {
	assets := make([]Asset, 0)

	for _, version := range p.Versions() {
		files := p.files(version)

		assets = append(assets, Asset{
			Version: version,
			Files:   files,
		})
	}

	// Sort by version
	sort.Sort(ByVersion(assets))

	return assets
}
