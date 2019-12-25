package packages

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xtuc/cdnjs-go/util"

	"github.com/pkg/errors"
)

const (
	PACKAGES_PATH = "./ajax/libs"
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

func ReadPackageJSON(ctx context.Context, file string) (*Package, error) {
	var jsondata map[string]interface{}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	jsonerr := json.Unmarshal(data, &jsondata)
	if jsonerr != nil {
		return nil, errors.Wrapf(jsonerr, "failed to parse %s", file)
	}

	var p Package
	p.ctx = ctx

	for key, value := range jsondata {
		switch key {
		case "title":
			p.Title = value.(string)
		case "name":
			p.Name = value.(string)
		case "description":
			p.Description = value.(string)
		case "version":
			p.Version = value.(string)
		case "author":
			{
				if str, ok := value.(string); ok {
					p.Author = Author{
						Name:  str,
						Email: "",
					}
				} else {
					util.Check(json.Unmarshal(data, &p.Author))
				}
			}
		case "filename":
			p.Filename = value.(string)
		case "homepage":
			p.Homepage = value.(string)
		case "repository":
			{
				if valuemap, ok := value.(map[string]interface{}); ok {
					p.Repository = Repository{
						Repotype: stringInObject("type", valuemap),
						Url:      stringInObject("url", valuemap),
					}
				} else {
					return nil, errors.New(fmt.Sprintf("failed to parse %s: unsupported Repository", file))
				}
			}
		case "keywords":
			{
				if values, ok := value.([]interface{}); ok {
					p.Keywords = make([]string, 0)
					for _, value := range values {
						p.Keywords = append(p.Keywords, value.(string))
					}
				} else {
					return nil, errors.New(fmt.Sprintf("failed to parse %s: unsupported Keywords", file))
				}
			}
		case "npmName":
			{
				str := value.(string)
				p.NpmName = &str
				// the package refers to a package on npm, we can set the autoupdate
				// method to npm
				p.Autoupdate = &Autoupdate{
					Source: "npm",
					Target: str,
				}
			}
		case "npmFileMap":
			{
				if values, ok := value.([]interface{}); ok {
					p.NpmFileMap = make([]FileMap, 0)
					for _, rawValue := range values {
						value := rawValue.(map[string]interface{})
						fileMap := FileMap{
							BasePath: stringInObject("basePath", value),
							Files:    make([]string, 0),
						}

						for _, file := range value["files"].([]interface{}) {
							fileMap.Files = append(fileMap.Files, file.(string))
						}

						p.NpmFileMap = append(p.NpmFileMap, fileMap)
					}
				} else {
					return nil, errors.New(fmt.Sprintf("failed to parse %s: unsupported npmFileMap", file))
				}
			}
		case "license":
			{
				if name, ok := value.(string); ok {
					p.License = &License{
						Name: name,
						Url:  "",
					}
				} else if valuemap, ok := value.(map[string]interface{}); ok {
					p.License = &License{
						Name: stringInObject("name", valuemap),
						Url:  stringInObject("url", valuemap),
					}
				} else {
					return nil, errors.New(fmt.Sprintf("failed to parse %s: unsupported Autoupdate", file))
				}
			}
		case "autoupdate":
			{
				if valuemap, ok := value.(map[string]interface{}); ok {
					p.Autoupdate = &Autoupdate{
						Source: valuemap["source"].(string),
						Target: valuemap["target"].(string),
					}
					if fileMap, ok := valuemap["fileMap"].([]interface{}); ok {
						p.Autoupdate.FileMap = new([]FileMap)
						for _, rawvalue := range fileMap {
							value := rawvalue.(map[string]interface{})
							autoupdateFileMap := FileMap{
								BasePath: stringInObject("basePath", value),
								Files:    make([]string, 0),
							}
							for _, file := range value["files"].([]interface{}) {
								autoupdateFileMap.Files = append(autoupdateFileMap.Files, file.(string))
							}
							*p.Autoupdate.FileMap = append(*p.Autoupdate.FileMap, autoupdateFileMap)
						}
					}
				} else {
					return nil, errors.New(fmt.Sprintf("failed to parse %s: unsupported Autoupdate", file))
				}
			}
		case "main":
		case "scripts":
		case "bugs":
		case "dependencies":
		case "devDependencies":
			// ignore
		default:
			util.Printf(ctx, "unknown field %s\n", key)
		}
	}

	return &p, nil
}

func (p *Package) path() string {
	return path.Join(PACKAGES_PATH, p.Name)
}

func (p *Package) Versions() (versions []string) {
	files, err := filepath.Glob(path.Join(p.path(), "*"))
	util.Check(err)
	// filter out package.json
	for _, file := range files {
		if !strings.HasSuffix(file, "package.json") {
			parts := strings.Split(file, "/")
			versions = append(versions, parts[3])
		}
	}
	return versions
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
