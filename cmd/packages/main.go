package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/xtuc/cdnjs-go/cloudstorage"
	"github.com/xtuc/cdnjs-go/packages"
	"github.com/xtuc/cdnjs-go/util"

	"cloud.google.com/go/storage"
)

func encodeJson(packages []outputPackage) (string, error) {
	out := struct {
		Packages []outputPackage `json:"packages"`
	}{
		packages,
	}

	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(&out)
	return buffer.String(), err
}

func main() {
	flag.Parse()
	subcommand := flag.Arg(0)

	ctx := context.Background()

	bkt, err := cloudstorage.GetBucket(ctx)
	util.Check(err)

	obj := bkt.Object("package.min.js")

	if subcommand == "set" {
		w := obj.NewWriter(ctx)
		_, err := io.Copy(w, os.Stdin)
		util.Check(err)
		util.Check(w.Close())
		util.Check(obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader))
		fmt.Println("Uploaded package.min.js")
		return
	}

	if subcommand == "generate" {
		files, err := filepath.Glob(path.Join(packages.PACKAGES_PATH, "*", "package.json"))
		util.Check(err)

		out := make([]outputPackage, 0)

		for _, f := range files {
			ctx := util.ContextWithName(f)

			p, err := packages.ReadPackageJSON(ctx, f)
			util.Check(err)

			if p.Version == "" {
				util.Printf(ctx, "version is invalid\n")
				continue
			}

			out = append(out, generatePackage(ctx, p))
		}

		str, err := encodeJson(out)
		util.Check(err)
		fmt.Println(string(str))
		return
	}

	panic("unknown subcommand")
}

// Struct used to serialize packages
type outputPackage struct {
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
	Assets  interface{} `json:"assets"`
}

func generatePackage(ctx context.Context, p *packages.Package) outputPackage {
	out := outputPackage{}

	out.Name = p.Name
	out.Version = p.Version
	out.Description = p.Description
	out.Filename = p.Filename
	out.Repository = map[string]string{
		"type": p.Repository.Repotype,
		"url":  p.Repository.Url,
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
	} else if p.Author.Url != nil && p.Author.Name != "" {
		out.Author = map[string]string{
			"name": p.Author.Name,
			"url":  *p.Author.Url,
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
		if p.License.Url != "" {
			out.License = map[string]string{
				"name": p.License.Name,
				"url":  p.License.Url,
			}
		} else {
			out.License = p.License.Name
		}
	}

	out.Assets = p.Assets()

	return out
}
