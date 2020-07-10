package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/agnivade/levenshtein"
	"github.com/cdnjs/tools/cloudstorage"
	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/sentry"
	"github.com/cdnjs/tools/util"

	"cloud.google.com/go/storage"
)

var (
	// initialize standard debug logger
	logger = util.GetStandardLogger()

	// default context (no logger prefix)
	defaultCtx = util.ContextWithEntries(util.GetStandardEntries("", logger)...)
)

func init() {
	sentry.Init()
}

func encodeJSON(packages []*outputPackage) (string, error) {
	out := struct {
		Packages []*outputPackage `json:"packages"`
	}{
		packages,
	}

	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(&out)
	return buffer.String(), err
}

func generatePackageWorker(jobs <-chan string, results chan<- *outputPackage) {
	for f := range jobs {
		// create context with file path prefix, standard debug logger
		ctx := util.ContextWithEntries(util.GetStandardEntries(f, logger)...)

		p, err := packages.ReadPackageJSON(ctx, f)
		if err != nil {
			util.Printf(ctx, "error while processing package: %s\n", err)
			results <- nil
			return
		}

		if p.Version == nil || *p.Version == "" {
			util.Printf(ctx, "version is invalid\n")
			results <- nil
			return
		}

		for _, version := range p.Versions() {
			if !hasSRI(p, version) {
				util.Printf(ctx, "version %s needs SRI calculation\n", version)

				sriFileMap := p.CalculateVersionSRIs(version)
				bytes, jsonErr := json.Marshal(sriFileMap)
				util.Check(jsonErr)

				writeSRIJSON(p, version, bytes)
			}
		}

		util.Printf(ctx, "OK\n")
		results <- generatePackage(ctx, p)
	}
}

func main() {
	defer sentry.PanicHandler()
	flag.Parse()

	if util.IsDebug() {
		fmt.Println("Running in debug mode")
	}

	switch subcommand := flag.Arg(0); subcommand {
	case "set":
		{
			ctx := defaultCtx
			bkt, err := cloudstorage.GetAssetsBucket(ctx)
			util.Check(err)
			obj := bkt.Object("package.min.js")

			w := obj.NewWriter(ctx)
			_, err = io.Copy(w, os.Stdin)
			util.Check(err)
			util.Check(w.Close())
			util.Check(obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader))
			fmt.Println("Uploaded package.min.js")
		}
	case "generate":
		{
			files, err := filepath.Glob(path.Join(util.GetCDNJSPackages(), "*", "package.json"))
			util.Check(err)

			numJobs := len(files)
			if numJobs == 0 {
				panic("cannot find packages")
			}

			jobs := make(chan string, numJobs)
			results := make(chan *outputPackage, numJobs)

			// spawn workers
			for w := 1; w <= runtime.NumCPU()*10; w++ {
				go generatePackageWorker(jobs, results)
			}

			// submit jobs; packages to encode
			for _, f := range files {
				jobs <- f
			}
			close(jobs)

			// collect results
			out := make([]*outputPackage, 0)
			for i := 1; i <= numJobs; i++ {
				if res := <-results; res != nil {
					out = append(out, res)
				}
			}

			str, err := encodeJSON(out)
			util.Check(err)
			fmt.Println(string(str))
		}
	default:
		panic(fmt.Sprintf("unknown subcommand: `%s`", subcommand))
	}
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

func generatePackage(ctx context.Context, p *packages.Package) *outputPackage {
	out := outputPackage{}

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

	assets := p.Assets()
	out.Assets = assets

	if files, ok := versionContainsFile(out.Name, out.Version, out.Filename, assets); !ok {
		if len(files) == 0 {
			// no files in this version, filename will be 404 on website
			sentry.NotifyError(fmt.Errorf("version contains no files: %s/%s", out.Name, out.Version))
		} else {
			p.Filename = getMostSimilarFilename(p.Filename, files) // get "best" alternative
		}
	}

	return &out
}

// Gets the most similar filename to a target filename.
// The []string of alternatives must have at least one element.
func getMostSimilarFilename(target string, filenames []string) string {
	var mostSimilar string
	var minDist int = math.MaxInt32
	for _, f := range filenames {
		if dist := levenshtein.ComputeDistance(target, f); dist < minDist {
			mostSimilar = f
			minDist = dist
		}
	}
	return mostSimilar
}

// Determines if a version contains a file.
// If not found, returns the version's files.
func versionContainsFile(name, version, filename string, assets []packages.Asset) ([]string, bool) {
	for _, a := range assets {
		if a.Version == version {
			for _, f := range a.Files {
				if f == filename {
					return nil, true // target version contains file
				}
			}
			// target version does not contain file, return files
			return a.Files, false
		}
	}
	// version not found, filename will be 404 on website
	sentry.NotifyError(fmt.Errorf("version not found in assets: %s/%s", name, version))
	return nil, true // send true to avoid processing this issue further
}

func hasSRI(p *packages.Package, version string) bool {
	sriPath := path.Join(util.SRIPath, p.Name, version+".json")
	_, statErr := os.Stat(sriPath)
	return !os.IsNotExist(statErr)
}

func writeSRIJSON(p *packages.Package, version string, content []byte) {
	sriDir := path.Join(util.SRIPath, p.Name)
	if _, err := os.Stat(sriDir); os.IsNotExist(err) {
		util.Check(os.MkdirAll(sriDir, 0777))
	}

	sriFilename := path.Join(sriDir, version+".json")
	util.Check(ioutil.WriteFile(sriFilename, content, 0777))
}
