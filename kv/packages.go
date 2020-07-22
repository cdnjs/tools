package kv

import (
	"context"
	"encoding/json"

	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/util"
)

// PackageMetadata contains metadata for a particular package.
// This is mirroring `outputPackage` from cmd/packages/main.go,
// which produces the package.json files.
// TODO:
// 		 - eventually remove packages.min.js entirely
//		 - SIMPLIFY -- can we marshal/unmarshal into the same struct instead of packages.Package and this one?
type PackageMetadata struct {
	Name        string            `json:"name"`
	Filename    string            `json:"filename"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Repository  map[string]string `json:"repository"`
	Keywords    []string          `json:"keywords"`
	// Author is either a name or an object with name and email
	Author   interface{} `json:"author"`
	Homepage string      `json:"homepage,omitempty"`
	// TODO: is this needed?
	Autoupdate *packages.Autoupdate `json:"autoupdate,omitempty"`
	// License is either a license or a list of licenses
	License interface{} `json:"license,omitempty"`
}

// GetPackage gets the package metadata from KV.
//
// TODO:
// - currently unused, will eventually replace reading `package.json` files from disk
func GetPackage(ctx context.Context, key string) (*packages.Package, error) {
	bytes, err := Read(key, packagesNamespaceID)
	if err != nil {
		return nil, err
	}
	return packages.ReadPackageJSONBytes(ctx, key, bytes)
}

// Gets the request to update a package metadata entry in KV with a new version.
func UpdateKVPackage(ctx context.Context, p *packages.Package) error {
	out := PackageMetadata{}

	out.Name = p.Name
	out.Version = *p.Version
	out.Description = p.Description
	out.Filename = p.Filename
	out.Repository = map[string]string{
		"type": p.Repository.Repotype,
		"url":  p.Repository.URL,
	}
	out.Keywords = p.Keywords

	if p.Homepage != "" {
		out.Homepage = p.Homepage
	}

	if p.Author.Email != "" && p.Author.Name != "" {
		out.Author = map[string]string{
			"name":  p.Author.Name,
			"email": p.Author.Email,
		}
	} else if p.Author.URL != nil && p.Author.Name != "" {
		out.Author = map[string]string{
			"name": p.Author.Name,
			"url":  *p.Author.URL,
		}
	} else if p.Author.Name != "" {
		out.Author = p.Author.Name
	}

	if p.Autoupdate != nil {
		// TODO: for some reason remove FileMap
		p.Autoupdate.FileMap = nil
		out.Autoupdate = p.Autoupdate
	}

	if p.License != nil {
		if p.License.URL != "" {
			out.License = map[string]string{
				"name": p.License.Name,
				"url":  p.License.URL,
			}
		} else {
			out.License = p.License.Name
		}
	}

	v, err := json.Marshal(out)
	util.Check(err)

	req := &writeRequest{
		key:   p.Name,
		value: v,
	}

	return encodeAndWriteKVBulk(ctx, []*writeRequest{req}, packagesNamespaceID)
}
