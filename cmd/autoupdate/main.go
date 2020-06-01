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
	"sort"

	"github.com/blang/semver"

	"github.com/cdnjs/tools/compress"
	"github.com/cdnjs/tools/metrics"
	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/util"
)

var (
	BASE_PATH     = util.GetEnv("BOT_BASE_PATH")
	PACKAGES_PATH = path.Join(BASE_PATH, "packages", "packages")
	CDNJS_PATH    = path.Join(BASE_PATH, "cdnjs")
)

func getPackages(ctx context.Context) []string {
	return util.ListFilesGlob(ctx, PACKAGES_PATH, "*/*.json")
}

type newVersionToCommit struct {
	versionPath string
	newVersion  string
	pckg        *packages.Package
}

func main() {
	flag.Parse()

	if util.IsDebug() {
		fmt.Println("Running in debug mode")
	}

	util.UpdateGitRepo(context.Background(), CDNJS_PATH)
	util.UpdateGitRepo(context.Background(), PACKAGES_PATH)

	for _, f := range getPackages(context.Background()) {
		ctx := util.ContextWithName(f)
		pckg, err := packages.ReadPackageJSON(ctx, path.Join(PACKAGES_PATH, f))
		util.Check(err)

		var newVersionsToCommit []newVersionToCommit

		if pckg.Autoupdate != nil {
			if pckg.Autoupdate.Source == "npm" {
				util.Debugf(ctx, "running npm update")
				newVersionsToCommit = updateNpm(ctx, pckg)
			}

			if pckg.Autoupdate.Source == "git" {
				util.Debugf(ctx, "running git update")
				newVersionsToCommit = updateGit(ctx, pckg)
			}
		}

		commitNewVersions(ctx, newVersionsToCommit, f)
	}

	packages.GitPush(context.Background(), CDNJS_PATH)
}

func packageJsonToString(packageJson map[string]interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(packageJson)
	return buffer.Bytes(), err
}

// During the bot migration, to ensure that both bots are not running at the
// same time on the same package we rely on a file
func addDoNotAddFile(ctx context.Context, pckg *packages.Package) {
	dest := path.Join(pckg.Path(), ".do_not_update")
	util.Debugf(ctx, "create %s\n", dest)

	f, err := os.Create(dest)
	util.Check(err)

	f.Close()

	// Add .do_not_update to git and it will be commited later
	packages.GitAdd(ctx, CDNJS_PATH, dest)
}

// Copy the package.json to the cdnjs repo and update its version
// TODO: this probaly needs ordering the versions to make sure to not
// accidentally put an older version of a package in the json
func updateVersionInCdnjs(ctx context.Context, pckg *packages.Package, newVersion string, packageJsonPath string) {
	var packageJson map[string]interface{}

	packageJsonData, err := ioutil.ReadFile(path.Join(PACKAGES_PATH, packageJsonPath))
	util.Check(err)

	util.Check(json.Unmarshal(packageJsonData, &packageJson))

	// Rewrite the version of the package.json to the latest update from the bot
	packageJson["version"] = newVersion

	newPackageJsonData, err := packageJsonToString(packageJson)
	util.Check(err)

	dest := path.Join(pckg.Path(), "package.json")
	file, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	util.Check(err)

	_, err = file.WriteString(string(newPackageJsonData))
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
			compress.CompressJpeg(ctx, absfile)
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
				compress.CompressPng(ctx, absfile)
			}
		}
	}
	// js
	{
		files := filterByExt(files, compress.JsExt)
		for _, file := range files {
			absfile := path.Join(version.versionPath, file)
			compress.CompressJs(ctx, absfile)
		}
	}
	// css
	{
		files := filterByExt(files, compress.CssExt)
		for _, file := range files {
			absfile := path.Join(version.versionPath, file)
			compress.CompressCss(ctx, absfile)
		}
	}
}

func commitNewVersions(ctx context.Context, newVersionsToCommit []newVersionToCommit, packageJsonPath string) {
	if len(newVersionsToCommit) == 0 {
		return
	}

	for _, newVersionToCommit := range newVersionsToCommit {
		util.Debugf(ctx, "adding version %s", newVersionToCommit.newVersion)

		// Compress assets
		compressNewVersion(ctx, newVersionToCommit)

		// Add to git the new version directory
		packages.GitAdd(ctx, CDNJS_PATH, newVersionToCommit.versionPath)

		updateVersionInCdnjs(ctx, newVersionToCommit.pckg, newVersionToCommit.newVersion, packageJsonPath)
		addDoNotAddFile(ctx, newVersionToCommit.pckg)

		// Add to git the update package.json
		packages.GitAdd(ctx, CDNJS_PATH, path.Join(newVersionToCommit.pckg.Path(), "package.json"))

		commitMsg := fmt.Sprintf("Add %s v%s", newVersionToCommit.pckg.Name, newVersionToCommit.newVersion)
		packages.GitCommit(ctx, CDNJS_PATH, commitMsg)

		metrics.ReportNewVersion()
	}
}

func getSemverOnly(versions []string) []string {
	newVersions := make([]string, 0)

	for _, v := range versions {
		_, err := semver.Make(v)
		if err == nil {
			newVersions = append(newVersions, v)
		}
	}

	return newVersions
}

func getLatestExistingVersion(existingVersionSet []string) *semver.Version {
	if len(existingVersionSet) == 0 {
		return nil
	}

	sort.Sort(packages.ByVersionString(existingVersionSet))

	// get the first semver version
	for i := 0; i < len(existingVersionSet); i++ {
		v, err := semver.Make(existingVersionSet[i])
		if err == nil {
			return &v
		}
	}

	// No semver version exist
	return nil
}
