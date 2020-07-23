package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
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
)

func main() {
	flag.Parse()

	if util.IsDebug() {
		fmt.Println("Running in debug mode")
	}

	switch subcommand := flag.Arg(0); subcommand {
	case "lint":
		{
			for _, path := range flag.Args()[1:] {
				lintPackage(path)
			}

			if errCount > 0 {
				os.Exit(1)
			}
		}
	case "show-files":
		{
			showFiles(flag.Arg(1))

			if errCount > 0 {
				os.Exit(1)
			}
		}
	case "meta":
		{
			fields := flag.Args()[1:]
			if len(fields) == 0 {
				panic("no fields specified")
			}

			printMeta(fields)
		}
	default:
		panic(fmt.Sprintf("unknown subcommand: `%s`", subcommand))
	}
}

func printMeta(nestedFields []string) {
	mainField := nestedFields[0]

	ctx := util.ContextWithEntries(util.GetStandardEntries(mainField, logger)...)
	packagesPath := util.GetPackagesPath()

	missingTypes := make(map[string]int)
	types := make(map[string]int)

	for _, f := range packages.GetPackagesJSONFiles(ctx) {
		ctx := util.ContextWithEntries(util.GetStandardEntries(f, logger)...)

		bytes, err := ioutil.ReadFile(path.Join(packagesPath, f))
		util.Check(err)

		var unknown interface{}
		util.Check(json.Unmarshal(bytes, &unknown))

		var cur string

		for i := 0; i <= len(nestedFields); i++ {
			u := unknown
			if i < len(nestedFields) {
				cur += "." + nestedFields[i]
				switch u.(type) {
				case string:
				case map[string]interface{}:
					if res, ok := unknown.(map[string]interface{})[nestedFields[i]]; ok {
						unknown = res
						continue
					}
				default:
					// assuming only strings and keys that map to strings
					panic(fmt.Sprintf("(%s) - unexpected type: %s", f, reflect.TypeOf(unknown)))
				}
			}

			t := reflect.TypeOf(unknown)
			if i == len(nestedFields) {
				util.Infof(ctx, "SUCCESS %s\n", cur)
				types[t.String()]++
				continue
			}

			util.Infof(ctx, "FAIL %s\n", cur)
			missingTypes[cur]++
			break
		}
	}

	util.Infof(ctx, "\n\nSummary of Types\n")
	for k, v := range types {
		util.Infof(ctx, "SUCCESS (%d): %s\n", v, k)
	}
	for k, v := range missingTypes {
		util.Infof(ctx, "MISSING (%d): %s\n", v, k)
	}
}

// Represents a version of a package,
// which could be a git version, npm version, etc.
type version interface {
	Get() string                    // Get the version as a string.
	Download(...interface{}) string // Download a version, returning the download dir.
	Clean(string)                   // Clean a download dir.
}

func showFiles(pckgPath string) {
	// create context with file path prefix, checker logger
	ctx := util.ContextWithEntries(util.GetCheckerEntries(pckgPath, logger)...)

	// parse package JSON
	pckg, readerr := packages.ReadPackageJSON(ctx, pckgPath)
	if readerr != nil {
		err(ctx, readerr.Error())
		return
	}

	// check for autoupdate
	if pckg.Autoupdate == nil {
		err(ctx, "autoupdate not found")
		return
	}

	// autoupdate exists, download latest versions based on source
	src := pckg.Autoupdate.Source
	var versions []version
	var downloadDir, noVersionsErr string
	switch src {
	case "npm":
		{
			// get npm versions and sort
			npmVersions, _ := npm.GetVersions(ctx, pckg.Autoupdate.Target)
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
			gitVersions, _ := git.GetVersions(ctx, pckg, packageGitDir)
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
			err(ctx, fmt.Sprintf("unknown autoupdate source: %s", src))
			return
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

// Prints the files of a package version, outputting debug
// messages if no valid files are present.
func printMostRecentVersion(ctx context.Context, p *packages.Package, dir string, v version) {
	fmt.Printf("\nmost recent version: %s\n", v.Get())
	downloadDir := v.Download(ctx, dir)
	defer v.Clean(downloadDir)
	filesToCopy := p.NpmFilesFrom(downloadDir)

	if len(filesToCopy) == 0 {
		errormsg := ""
		errormsg += fmt.Sprintf("No files will be published for version %s.\n", v.Get())

		// determine if a pattern has been seen before
		seen := make(map[string]bool)

		for _, filemap := range p.NpmFileMap {
			for _, pattern := range filemap.Files {
				if _, ok := seen[pattern]; ok {
					continue // skip duplicate pattern
				}
				seen[pattern] = true
				errormsg += fmt.Sprintf("[Click here to debug your glob pattern `%s`](%s).\n", pattern, makeGlobDebugLink(pattern, downloadDir))
			}
		}
		err(ctx, errormsg)
		return
	}

	fmt.Printf("\n```\n")
	for _, file := range filesToCopy {
		fmt.Printf("%s\n", file.To)
	}
	fmt.Printf("```\n")
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
	if !strings.Contains(pckg.Repository.URL, "github.com") {
		return false
	}

	if s := git.GetGitHubStars(pckg.Repository.URL); s.Stars < util.MinGitHubStars {
		warn(ctx, fmt.Sprintf("stars on GitHub is under %d", util.MinGitHubStars))
		return false
	}
	return true
}

func lintPackage(pckgPath string) {
	// create context with file path prefix, checker logger
	ctx := util.ContextWithEntries(util.GetCheckerEntries(pckgPath, logger)...)

	util.Debugf(ctx, "Linting %s...\n", pckgPath)

	pckg, readerr := packages.ReadPackageJSON(ctx, pckgPath)
	if readerr != nil {
		err(ctx, readerr.Error())
		return
	}

	if pckg.Name == "" {
		err(ctx, shouldNotBeEmpty(".name"))
	}

	if pckg.Version != nil {
		err(ctx, shouldNotExist(".version"))
	}

	if pckg.Repository.URL == "" {
		err(ctx, shouldNotBeEmpty(".repository.url"))
	}

	if pckg.Autoupdate != nil {
		if pckg.Autoupdate.Source == "git" && pckg.Autoupdate.Target != pckg.Repository.URL {
			err(ctx, ".autoupdate.target and .repository.url must not differ")
		}

		switch pckg.Autoupdate.Source {
		case "npm":
			{
				// check that it exists
				if !npm.Exists(pckg.Autoupdate.Target) {
					err(ctx, "package doesn't exist on npm")
					break
				}

				// check if it has enough downloads
				if md := npm.GetMonthlyDownload(pckg.Autoupdate.Target); md.Downloads < util.MinNpmMonthlyDownloads {
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
				err(ctx, "Unsupported .autoupdate.source: "+pckg.Autoupdate.Source)
			}
		}
	} else {
		err(ctx, ".autoupdate should not be null. Package will never auto-update")
	}

	// ensure repo type is git
	if pckg.Repository.Repotype != "git" {
		err(ctx, "Unsupported .repository.type: "+pckg.Repository.Repotype)
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

func shouldNotBeEmpty(name string) string {
	return fmt.Sprintf("%s should be specified", name)
}

func shouldNotExist(name string) string {
	return fmt.Sprintf("%s should not exist", name)
}
