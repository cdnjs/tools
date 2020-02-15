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

	"github.com/blang/semver"

	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/util"
)

const (
	// When no versions exist in cdnjs and we are trying to import all of them,
	// limit it to the a few last versions to avoid publishing too many outdated
	// versions
	IMPORT_ALL_MAX_VERSIONS = 10
)

var (
	BASE_PATH     = util.GetEnv("BOT_BASE_PATH")
	PACKAGES_PATH = path.Join(BASE_PATH, "packages", "packages")
	CDNJS_PATH    = path.Join(BASE_PATH, "cdnjs")
)

func getPackages() []string {
	return util.ListFilesGlob(PACKAGES_PATH, "*/*.json")
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

	for _, f := range getPackages() {
		ctx := util.ContextWithName(f)
		pckg, err := packages.ReadPackageJSON(ctx, path.Join(PACKAGES_PATH, f))
		util.Check(err)

		// Attach the autoupdate logger
		ctx = util.WithLogger(ctx)

		var newVersionsToCommit []newVersionToCommit

		if pckg.Autoupdate != nil {
			log(ctx, LogAutoupdateStarted{Source: pckg.Autoupdate.Source})

			if pckg.Autoupdate.Source == "npm" {
				newVersionsToCommit = updateNpm(ctx, pckg)
			}
		}

		commitNewVersions(ctx, newVersionsToCommit, f)
		publishAutoupdateLog(ctx, pckg.Name)
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

func commitNewVersions(ctx context.Context, newVersionsToCommit []newVersionToCommit, packageJsonPath string) {
	if len(newVersionsToCommit) == 0 {
		return
	}

	for _, newVersionToCommit := range newVersionsToCommit {
		// Add to git the new version directory
		packages.GitAdd(ctx, CDNJS_PATH, newVersionToCommit.versionPath)

		updateVersionInCdnjs(ctx, newVersionToCommit.pckg, newVersionToCommit.newVersion, packageJsonPath)
		addDoNotAddFile(ctx, newVersionToCommit.pckg)

		// Add to git the update package.json
		packages.GitAdd(ctx, CDNJS_PATH, path.Join(newVersionToCommit.pckg.Path(), "package.json"))

		commitMsg := fmt.Sprintf("Add %s v%s", newVersionToCommit.pckg.Name, newVersionToCommit.newVersion)
		packages.GitCommit(ctx, CDNJS_PATH, commitMsg)

		log(ctx, LogNewVersionCommit{Version: newVersionToCommit.newVersion})
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
