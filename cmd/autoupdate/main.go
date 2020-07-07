package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"

	"github.com/cdnjs/tools/compress"
	"github.com/cdnjs/tools/metrics"
	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/sentry"
	"github.com/cdnjs/tools/util"
)

func init() {
	sentry.Init()
}

var (
	basePath     = util.GetBotBasePath()
	packagesPath = path.Join(basePath, "packages", "packages")
	cdnjsPath    = util.GetCDNJSPath()

	// initialize standard debug logger
	logger = util.GetStandardLogger()

	// default context (no logger prefix)
	defaultCtx = util.ContextWithEntries(util.GetStandardEntries("", logger)...)

	noUpdate bool
)

func getPackages(ctx context.Context) []string {
	list, err := util.ListFilesGlob(ctx, packagesPath, "*/*.json")
	util.Check(err)
	return list
}

type newVersionToCommit struct {
	commitType  string // Either "newVersion" "newLatestVersion"
	versionPath string
	packageJSONPath string
	newVersion  string
	pckg        *packages.Package
}

func worker(id int, wg *sync.WaitGroup, packagePaths <-chan string, commits chan<- newVersionToCommit) {
	defer wg.Done()
	for f := range packagePaths {
		// create context with file path prefix, standard debug logger
		// TODO: Use the worker id in the logging output...
		ctx := util.ContextWithEntries(util.GetStandardEntries(f, logger)...)

		pckg, err := packages.ReadPackageJSON(ctx, path.Join(packagesPath, f))
		util.Check(err)

		var newVersionsToCommit []newVersionToCommit
		var latestVersion *newVersionToCommit

		if pckg.Autoupdate != nil {
			if pckg.Autoupdate.Source == "npm" {
				util.Debugf(ctx, "running npm update")
				newVersionsToCommit, latestVersion = updateNpm(ctx, pckg)
			}

			if pckg.Autoupdate.Source == "git" {
				util.Debugf(ctx, "running git update")
				newVersionsToCommit, latestVersion = updateGit(ctx, pckg)
			}
		}
		if !noUpdate && len(newVersionsToCommit) > 0 {
			for _, commit := range newVersionsToCommit {
				commits <- commit
			}
			if latestVersion != nil && (pckg.Version == nil || *pckg.Version != latestVersion.newVersion) {
				latestVersion.packageJSONPath = f
				commits <- *latestVersion
			}
		}
	}
}

func main() {
	defer sentry.PanicHandler()

	var noPull bool
	var workers int
	flag.BoolVar(&noUpdate, "no-update", false, "if set, the autoupdater will not commit or push to git")
	flag.BoolVar(&noPull, "no-pull", false, "if set, the autoupdater will not pull from git")
	flag.IntVar(&workers, "workers", 5, "the number of workers (default: 5)")
	flag.Parse()

	if util.IsDebug() {
		fmt.Printf("Running in debug mode (no-update=%t, no-pull=%t, workers=%d)\n", noUpdate, noPull, workers)
	}

	if !noPull {
		util.UpdateGitRepo(defaultCtx, cdnjsPath)
		util.UpdateGitRepo(defaultCtx, packagesPath)
	}

	packagesPath := make(chan string)
	commits := make(chan newVersionToCommit, 10)

	var wg sync.WaitGroup
	for w := 1; w <= workers; w++ {
		wg.Add(1)
		go worker(w, &wg, packagesPath, commits)
	}

	go func() {
		for _, f := range getPackages(defaultCtx) {
			packagesPath <- f
		}
		close(packagesPath)
	}()

	var commitWg sync.WaitGroup
	if !noUpdate {
		go func() {
			for commit := range commits {
				if commit.commitType == "newVersion" {
					commitNewVersions(defaultCtx, commit)
				} else if commit.commitType == "newLatestVersion" {
					commitPackageVersion(defaultCtx, commit.pckg, commit.newVersion, commit.packageJSONPath)
				}
			}
		}()
	}

	wg.Wait()
	close(commits)
	commitWg.Wait()

	if !noUpdate {
		packages.GitPush(defaultCtx, cdnjsPath)
	}
}

func packageJSONToString(packageJSON map[string]interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(packageJSON)
	return buffer.Bytes(), err
}

// Copy the package.json to the cdnjs repo and update its version.
func updateVersionInCdnjs(ctx context.Context, pckg *packages.Package, newVersion, packageJSONPath string) {
	var packageJSON map[string]interface{}

	packageJSONData, err := ioutil.ReadFile(path.Join(packagesPath, packageJSONPath))
	util.Check(err)

	util.Check(json.Unmarshal(packageJSONData, &packageJSON))

	// Rewrite the version of the package.json to the latest update from the bot
	packageJSON["version"] = newVersion

	newPackageJSONData, err := packageJSONToString(packageJSON)
	util.Check(err)

	dest := path.Join(pckg.Path(), "package.json")
	file, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	util.Check(err)

	_, err = file.WriteString(string(newPackageJSONData))
	util.Check(err)
}

// Filter a list of files by extensions
func filterByExt(files []string, extensions map[string]bool) []string {
	var matches []string
	for _, file := range files {
		ext := path.Ext(file)
		if v, ok := extensions[ext]; ok && v {
			matches = append(matches, file)
		}
	}
	return matches
}

func compressNewVersion(ctx context.Context, version newVersionToCommit) {
	files := version.pckg.AllFiles(version.newVersion)

	// jpeg
	{
		files := filterByExt(files, compress.JpegExt)
		for _, file := range files {
			absfile := path.Join(version.versionPath, file)
			compress.Jpeg(ctx, absfile)
		}
	}
	// png
	{
		// if a `.donotoptimizepng` is present in the package ignore png
		// compression
		_, err := os.Stat(path.Join(version.pckg.Path(), ".donotoptimizepng"))
		if os.IsNotExist(err) {
			files := filterByExt(files, compress.PngExt)
			for _, file := range files {
				absfile := path.Join(version.versionPath, file)
				compress.Png(ctx, absfile)
			}
		}
	}
	// js
	{
		files := filterByExt(files, compress.JsExt)
		for _, file := range files {
			absfile := path.Join(version.versionPath, file)
			compress.Js(ctx, absfile)
		}
	}
	// css
	{
		files := filterByExt(files, compress.CSSExt)
		for _, file := range files {
			absfile := path.Join(version.versionPath, file)
			compress.CSS(ctx, absfile)
		}
	}
}

func commitNewVersions(ctx context.Context, newVersionToCommit newVersionToCommit) {
	util.Debugf(ctx, "adding version %s", newVersionToCommit.newVersion)

	// Compress assets
	compressNewVersion(ctx, newVersionToCommit)

	// Add to git the new version directory
	packages.GitAdd(ctx, cdnjsPath, newVersionToCommit.versionPath)

	commitMsg := fmt.Sprintf("Add %s v%s", newVersionToCommit.pckg.Name, newVersionToCommit.newVersion)
	packages.GitCommit(ctx, cdnjsPath, commitMsg)

	metrics.ReportNewVersion(ctx)
}

func commitPackageVersion(ctx context.Context, pckg *packages.Package, latestVersion, packageJSONPath string) {
	util.Debugf(ctx, "adding latest version to package.json %s", latestVersion)

	// Update package.json file
	updateVersionInCdnjs(ctx, pckg, latestVersion, packageJSONPath)

	// Add to git the updated package.json
	packages.GitAdd(ctx, cdnjsPath, path.Join(pckg.Path(), "package.json"))

	commitMsg := fmt.Sprintf("Add %s package.json (v%s)", pckg.Name, latestVersion)
	packages.GitCommit(ctx, cdnjsPath, commitMsg)
}
