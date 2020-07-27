package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/blang/semver"
	"github.com/cdnjs/tools/compress"
	"github.com/cdnjs/tools/kv"
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
	packagesPath = util.GetHumanPackagesPath()
	cdnjsPath    = util.GetCDNJSPath()

	// initialize standard debug logger
	logger = util.GetStandardLogger()

	// default context (no logger prefix)
	defaultCtx = util.ContextWithEntries(util.GetStandardEntries("", logger)...)
)

type version interface {
	Get() string             // Get the version.
	GetTimeStamp() time.Time // GetTimeStamp gets the time stamp associated with the version.
}

type newVersionToCommit struct {
	versionPath string
	newVersion  string
	pckg        *packages.Package
	timestamp   time.Time
}

// Get is used to get the new version.
func (n newVersionToCommit) Get() string {
	return n.newVersion
}

// GetTimeStamp gets the time stamp associated with the new version.
func (n newVersionToCommit) GetTimeStamp() time.Time {
	return n.timestamp
}

func main() {
	defer sentry.PanicHandler()

	var noUpdate bool
	var noPull bool
	flag.BoolVar(&noUpdate, "no-update", false, "if set, the autoupdater will not commit or push to git")
	flag.BoolVar(&noPull, "no-pull", false, "if set, the autoupdater will not pull from git")
	flag.Parse()

	if util.IsDebug() {
		fmt.Printf("Running in debug mode (no-update=%t, no-pull=%t)\n", noUpdate, noPull)
	}

	if !noPull {
		util.UpdateGitRepo(defaultCtx, cdnjsPath)
		util.UpdateGitRepo(defaultCtx, packagesPath)
	}

	// create channel to handle signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM)

	for _, f := range packages.GetHumanPackageJSONFiles(defaultCtx) {
		// create context with file path prefix, standard debug logger
		ctx := util.ContextWithEntries(util.GetStandardEntries(f, logger)...)

		select {
		case sig := <-c:
			util.Debugf(ctx, "RECEIVED SIGNAL: %s\n", sig)
			return
		default:
		}

		pckg, err := packages.ReadHumanPackageJSON(ctx, path.Join(packagesPath, f))
		util.Check(err)

		var newVersionsToCommit []newVersionToCommit
		var allVersions []version

		switch *pckg.Autoupdate.Source {
		case "npm":
			{
				util.Debugf(ctx, "running npm update")
				newVersionsToCommit, allVersions = updateNpm(ctx, pckg)
			}
		case "git":
			{
				util.Debugf(ctx, "running git update")
				newVersionsToCommit, allVersions = updateGit(ctx, pckg)
			}
		default:
			{
				panic(fmt.Sprintf("%s invalid autoupdate source: %s", *pckg.Name, *pckg.Autoupdate.Source))
			}
		}

		if !noUpdate {
			if len(newVersionsToCommit) > 0 {
				commitNewVersions(ctx, newVersionsToCommit)
				packages.GitPush(ctx, cdnjsPath)
				if !util.IsKVDisabled() {
					writeNewVersionsToKV(ctx, newVersionsToCommit)
				}
			}
			if len(allVersions) > 0 {
				latestVersion := getLatestStableVersion(allVersions)
				if latestVersion == nil {
					latestVersion = getLatestVersion(allVersions)
				}
				if latestVersion != nil {
					destpckg, err := packages.ReadHumanPackageJSON(ctx, path.Join(pckg.LibraryPath(), "package.json"))
					if err != nil || destpckg.Version == nil || *destpckg.Version != *latestVersion {
						commitPackageVersion(ctx, pckg, *latestVersion, f)
						packages.GitPush(ctx, cdnjsPath)

						// TODO:
						// Later need to change ReadPackageJSON to read the kv.Package from KV.
						// If the kv.Package does not exist, we will create one.
						// Otherwise we will update the existing one's latest version.
						// This kv.Package will then be passed to the kv.UpdateKVPackage function directly.
						pckg.Version = latestVersion
						if err := kv.UpdateKVPackage(ctx, pckg); err != nil {
							util.Debugf(ctx, "failed to update KV package metadata: %s\n", err)
						}
					}
				}
			}
		}
	}
}

// Gets the latest stable version by time stamp. A  "stable" version is
// considered to be a version that contains no pre-releases.
// If no latest stable version is found (ex. all are non-semver), a nil *string
// will be returned.
func getLatestStableVersion(versions []version) *string {
	var latest *string
	var latestTime time.Time
	for _, v := range versions {
		vStr := v.Get()
		if s, err := semver.Parse(vStr); err == nil && len(s.Pre) == 0 {
			timeStamp := v.GetTimeStamp()
			if latest == nil || timeStamp.After(latestTime) {
				latest = &vStr
				latestTime = timeStamp
			}
		}
	}
	return latest
}

// Gets the latest version by time stamp. If it does not exist, a nil *string is returned.
func getLatestVersion(versions []version) *string {
	var latest *string
	var latestTime time.Time
	for _, v := range versions {
		vStr, timeStamp := v.Get(), v.GetTimeStamp()
		if latest == nil || timeStamp.After(latestTime) {
			latest = &vStr
			latestTime = timeStamp
		}
	}
	return latest
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

	dest := path.Join(pckg.LibraryPath(), "package.json")
	file, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	util.Check(err)

	_, err = file.WriteString(string(newPackageJSONData))
	util.Check(err)
}

// Optimizes/minifies files on disk for a particular package version.
func optimizeAndMinify(ctx context.Context, version newVersionToCommit) {
	files := version.pckg.AllFiles(version.newVersion)

	for _, f := range files {
		switch path.Ext(f) {
		case ".jpg", ".jpeg":
			compress.Jpeg(ctx, path.Join(version.versionPath, f))
		case ".png":
			compress.Png(ctx, path.Join(version.versionPath, f))
		case ".js":
			compress.Js(ctx, path.Join(version.versionPath, f))
		case ".css":
			compress.CSS(ctx, path.Join(version.versionPath, f))
		}
	}
}

// write all versions to KV
func writeNewVersionsToKV(ctx context.Context, newVersionsToCommit []newVersionToCommit) {
	for _, newVersionToCommit := range newVersionsToCommit {
		pkg, version := *newVersionToCommit.pckg.Name, newVersionToCommit.newVersion

		util.Debugf(ctx, "writing version to KV %s", path.Join(pkg, version))
		if err := kv.InsertNewVersionToKV(ctx, pkg, version, newVersionToCommit.versionPath); err != nil {
			sentry.NotifyError(fmt.Errorf("kv write %s: %s", path.Join(pkg, version), err.Error()))
		}
	}
}

func commitNewVersions(ctx context.Context, newVersionsToCommit []newVersionToCommit) {
	for _, newVersionToCommit := range newVersionsToCommit {
		util.Debugf(ctx, "adding version %s", newVersionToCommit.newVersion)

		// Optimize/minifiy assets (compressing br/gz will occur later)
		optimizeAndMinify(ctx, newVersionToCommit)

		// Add to git the new version directory
		packages.GitAdd(ctx, cdnjsPath, newVersionToCommit.versionPath)

		commitMsg := fmt.Sprintf("Add %s v%s", *newVersionToCommit.pckg.Name, newVersionToCommit.newVersion)
		packages.GitCommit(ctx, cdnjsPath, commitMsg)

		metrics.ReportNewVersion(ctx)
	}
}

func commitPackageVersion(ctx context.Context, pckg *packages.Package, latestVersion, packageJSONPath string) {
	util.Debugf(ctx, "adding latest version to package.json %s", latestVersion)

	// Update package.json file
	updateVersionInCdnjs(ctx, pckg, latestVersion, packageJSONPath)

	// Add to git the updated package.json
	packages.GitAdd(ctx, cdnjsPath, path.Join(pckg.LibraryPath(), "package.json"))

	commitMsg := fmt.Sprintf("Set %s package.json (v%s)", *pckg.Name, latestVersion)
	packages.GitCommit(ctx, cdnjsPath, commitMsg)
}
