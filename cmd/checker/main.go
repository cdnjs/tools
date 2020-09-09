package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/cdnjs/tools/git"
	"github.com/cdnjs/tools/npm"
	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/util"
)

var (
	// Store the number of validation errors
	errCount uint = 0

	// initialize checker debug logger
	logger = util.GetCheckerLogger()

	// regex for path in cdnjs/packages/
	pckgPathRegex = regexp.MustCompile("^packages/([a-z0-9])/([a-zA-Z0-9._-]+).json$")
)

func main() {
	var noPathValidation bool
	flag.BoolVar(&noPathValidation, "no-path-validation", false, "If set, all package paths are accepted.")
	flag.Parse()

	if util.IsDebug() {
		fmt.Println("Running in debug mode")
	}

	switch subcommand := flag.Arg(0); subcommand {
	case "lint":
		{
			for _, path := range flag.Args()[1:] {
				lintPackage(path, noPathValidation)
			}

			if errCount > 0 {
				os.Exit(1)
			}
		}
	case "show-files":
		{
			showFiles(flag.Arg(1), noPathValidation)

			if errCount > 0 {
				os.Exit(1)
			}
		}
	default:
		panic(fmt.Sprintf("unknown subcommand: `%s`", subcommand))
	}
}

// Represents a version of a package,
// which could be a git version, npm version, etc.
type version interface {
	Get() string                    // Get the version as a string.
	Download(...interface{}) string // Download a version, returning the download dir.
	Clean(string)                   // Clean a download dir.
}

func showFiles(pckgPath string, noPathValidation bool) {
	// create context with file path prefix, checker logger
	ctx := util.ContextWithEntries(util.GetCheckerEntries(pckgPath, logger)...)

	// parse *Package from JSON
	pckg := parseHumanPackage(ctx, pckgPath, noPathValidation)
	if pckg == nil {
		return
	}

	// autoupdate exists, download latest versions based on source
	src := *pckg.Autoupdate.Source
	var versions []version
	var downloadDir, noVersionsErr string
	switch src {
	case "npm":
		{
			// get npm versions and sort
			npmVersions, _ := npm.GetVersions(ctx, *pckg.Autoupdate.Target)
			sort.Sort(npm.ByTimeStamp(npmVersions))

			// cast to interface
			for _, v := range npmVersions {
				versions = append(versions, v)
			}

			// download into temp dir
			if len(versions) > 0 {
				downloadDir = npm.DownloadTar(ctx, npmVersions[0].Tarball)
			}

			// set err string if no versions
			noVersionsErr = "no version found on npm"
		}
	case "git":
		{
			// make temp dir and clone
			packageGitDir, direrr := ioutil.TempDir("", src)
			util.Check(direrr)
			out, cloneerr := packages.GitClone(ctx, pckg, packageGitDir)
			if cloneerr != nil {
				err(ctx, fmt.Sprintf("could not clone repo: %s: %s\n", cloneerr, out))
				return
			}

			// get git versions and sort
			gitVersions, _ := git.GetVersions(ctx, packageGitDir)
			sort.Sort(git.ByTimeStamp(gitVersions))

			// cast to interface
			for _, v := range gitVersions {
				versions = append(versions, v)
			}

			// set download dir
			downloadDir = packageGitDir

			// set err string if no versions
			noVersionsErr = "no tagged version found in git"
		}
	default:
		{
			panic(fmt.Sprintf("unknown autoupdate source: %s", src))
		}
	}

	// clean up temp dir
	defer os.RemoveAll(downloadDir)

	// enforce at least one version
	if len(versions) == 0 {
		err(ctx, noVersionsErr)
		return
	}

	// limit versions
	if len(versions) > util.ImportAllMaxVersions {
		versions = versions[:util.ImportAllMaxVersions]
	}

	// print info for first src version
	printMostRecentVersion(ctx, pckg, downloadDir, versions[0])

	// print aggregate info for the few last src versions
	printLastVersions(ctx, pckg, downloadDir, versions[1:])
}

// Try to parse a *Package, outputting ci errors/warnings.
// If there is an issue, *Package will be nil.
func parseHumanPackage(ctx context.Context, pckgPath string, noPathValidation bool) *packages.Package {
	if !noPathValidation {
		// check package path matches regex
		matches := pckgPathRegex.FindStringSubmatch(pckgPath)
		if matches == nil {
			err(ctx, fmt.Sprintf("package path `%s` does not match %s", pckgPath, pckgPathRegex.String()))
			return nil
		}

		// check the package is going into the correct folder
		// (ex. My-Package -> packages/m/My-Package.json)
		actualDir, pckgName := matches[1], matches[2]
		expectedDir := strings.ToLower(string(pckgName[0]))
		if actualDir != expectedDir {
			err(ctx, fmt.Sprintf("package `%s` must go into `%s` dir, not `%s` dir", pckgName, expectedDir, actualDir))
			return nil
		}
	}

	// parse package JSON
	pckg, readerr := packages.ReadHumanJSONFile(ctx, pckgPath)
	if readerr != nil {
		if invalidHumanErr, ok := readerr.(packages.InvalidSchemaError); ok {
			// output all schema errors
			for _, resErr := range invalidHumanErr.Result.Errors() {
				err(ctx, resErr.String())
			}
		} else {
			err(ctx, readerr.Error())
		}
		return nil
	}

	checkFilename(ctx, pckg)
	return pckg
}

// Prints the files of a package version, outputting debug
// messages if no valid files are present.
func printMostRecentVersion(ctx context.Context, p *packages.Package, dir string, v version) {
	fmt.Printf("\nmost recent version: %s\n", v.Get())
	downloadDir := v.Download(ctx, dir)
	defer v.Clean(downloadDir)
	filesToCopy := p.NpmFilesFrom(downloadDir)

	if len(filesToCopy) == 0 {
		errormsg := fmt.Sprintf("No files will be published for version %s.\n", v.Get())

		for _, fileMap := range p.Autoupdate.FileMap {
			for _, pattern := range fileMap.Files {
				errormsg += fmt.Sprintf("[Click here to debug your glob pattern `%s`](%s).\n", pattern, makeGlobDebugLink(pattern, downloadDir))
			}
		}
		err(ctx, errormsg)
		return
	}

	var filenameFound bool

	fmt.Printf("\n```\n")
	for _, file := range filesToCopy {
		fmt.Printf("%s\n", file.To)
		if p.Filename != nil && !filenameFound && file.To == *p.Filename {
			filenameFound = true
		}
	}
	fmt.Printf("```\n")

	if p.Filename != nil && !filenameFound {
		err(ctx, fmt.Sprintf("Filename `%s` not found in most recent version `%s`.\n", *p.Filename, v.Get()))
	}
}

// Prints the matching files of a number of last versions.
// Each previous version will be downloaded and cleaned up if necessary.
// For example, a temporary directory may be downloaded and then removed later.
func printLastVersions(ctx context.Context, p *packages.Package, dir string, versions []version) {
	fmt.Printf("\n%d last version(s):\n", len(versions))
	for _, version := range versions {
		downloadDir := version.Download(ctx, dir)
		defer version.Clean(downloadDir)

		filesToCopy := p.NpmFilesFrom(downloadDir)

		fmt.Printf("- %s: %d file(s) matched", version.Get(), len(filesToCopy))
		if len(filesToCopy) > 0 {
			fmt.Printf(" :heavy_check_mark:\n")
		} else {
			fmt.Printf(" :heavy_exclamation_mark:\n")
		}
	}
}

func makeGlobDebugLink(glob string, dir string) string {
	encodedGlob := url.QueryEscape(glob)
	allTests := ""

	util.Check(filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			allTests += "&tests=" + url.QueryEscape(info.Name())
		}
		return nil
	}))

	return fmt.Sprintf("https://www.digitalocean.com/community/tools/glob?comments=true&glob=%s&matches=true%s&tests=", encodedGlob, allTests)
}

func checkGitHubPopularity(ctx context.Context, pckg *packages.Package) bool {
	if !strings.Contains(*pckg.Repository.URL, "github.com") {
		return false
	}

	if s := git.GetGitHubStars(*pckg.Repository.URL); s.Stars < util.MinGitHubStars {
		warn(ctx, fmt.Sprintf("stars on GitHub is under %d", util.MinGitHubStars))
		return false
	}
	return true
}

func checkFilename(ctx context.Context, pckg *packages.Package) {
	// warn if filename is not present
	// current, only a few packages have exceptions
	// that allow them to have missing filenames
	if pckg.Filename == nil {
		warn(ctx, "filename is missing")
	}
}

func lintPackage(pckgPath string, noPathValidation bool) {
	// create context with file path prefix, checker logger
	ctx := util.ContextWithEntries(util.GetCheckerEntries(pckgPath, logger)...)

	util.Debugf(ctx, "Linting %s...\n", pckgPath)

	// parse *Package from JSON
	pckg := parseHumanPackage(ctx, pckgPath, noPathValidation)
	if pckg == nil {
		return
	}

	switch *pckg.Autoupdate.Source {
	case "npm":
		{
			// check that it exists
			if !npm.Exists(*pckg.Autoupdate.Target) {
				err(ctx, "package doesn't exist on npm")
				break
			}

			// check if it has enough downloads
			if md := npm.GetMonthlyDownload(*pckg.Autoupdate.Target); md.Downloads < util.MinNpmMonthlyDownloads {
				if !checkGitHubPopularity(ctx, pckg) {
					warn(ctx, fmt.Sprintf("package download per month on npm is under %d", util.MinNpmMonthlyDownloads))
				}
			}
		}
	case "git":
		{
			checkGitHubPopularity(ctx, pckg)
		}
	default:
		{
			// schema will enforce npm or git, so panic
			panic(fmt.Sprintf("unsupported .autoupdate.source: " + *pckg.Autoupdate.Source))
		}
	}
}

// wrapper around outputting a checker error
func err(ctx context.Context, s string) {
	util.Errf(ctx, s)
	errCount++
}

// wrapper around outputting a checker warning
func warn(ctx context.Context, s string) {
	util.Warnf(ctx, s)
}
