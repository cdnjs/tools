package packages

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/cdnjs/tools/util"

	"github.com/pkg/errors"
)

// ReadPackageJSON parses a JSON file into a Package.
func ReadPackageJSON(ctx context.Context, file string) (*Package, error) {
	var jsondata map[string]interface{}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read %s", file)
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
			s := value.(string)
			p.Version = &s
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
						URL:      stringInObject("url", valuemap),
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
						URL:  "",
					}
				} else if valuemap, ok := value.(map[string]interface{}); ok {
					p.License = &License{
						Name: stringInObject("name", valuemap),
						URL:  stringInObject("url", valuemap),
					}
				} else {
					return nil, errors.New(fmt.Sprintf("failed to parse %s: unsupported Autoupdate", file))
				}
			}
		case "autoupdate":
			{
				if valuemap, ok := value.(map[string]interface{}); ok {
					source, sourceok := valuemap["source"].(string)
					target, targetok := valuemap["target"].(string)
					if sourceok && targetok {
						p.Autoupdate = &Autoupdate{
							Source: source,
							Target: target,
						}
						if fileMap, ok := valuemap["fileMap"].([]interface{}); ok {
							p.NpmFileMap = make([]FileMap, 0)
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
								p.NpmFileMap = append(p.NpmFileMap, autoupdateFileMap)
							}
						}
					} else {
						return nil, errors.New(fmt.Sprintf("failed to parse %s: unsupported Autoupdate map", file))
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
