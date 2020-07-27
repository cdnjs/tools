package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"

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

func encodeJSON(pkgs []*packages.Package) (string, error) {
	out := struct {
		Packages []*packages.Package `json:"packages"`
	}{
		pkgs,
	}

	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(&out)
	return buffer.String(), err
}

func generatePackageWorker(jobs <-chan string, results chan<- *packages.Package) {
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
		p.Assets = p.GetAssets()
		results <- p
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
			files, err := filepath.Glob(path.Join(util.GetCDNJSLibrariesPath(), "*", "package.json"))
			util.Check(err)

			numJobs := len(files)
			if numJobs == 0 {
				panic("cannot find packages")
			}

			jobs := make(chan string, numJobs)
			results := make(chan *packages.Package, numJobs)

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
			out := make([]*packages.Package, 0)
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

func hasSRI(p *packages.Package, version string) bool {
	sriPath := path.Join(util.SRIPath, *p.Name, version+".json")
	_, statErr := os.Stat(sriPath)
	return !os.IsNotExist(statErr)
}

func writeSRIJSON(p *packages.Package, version string, content []byte) {
	sriDir := path.Join(util.SRIPath, *p.Name)
	if _, err := os.Stat(sriDir); os.IsNotExist(err) {
		util.Check(os.MkdirAll(sriDir, 0777))
	}

	sriFilename := path.Join(sriDir, version+".json")
	util.Check(ioutil.WriteFile(sriFilename, content, 0777))
}
