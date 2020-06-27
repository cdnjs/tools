package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sort"

	"github.com/cdnjs/tools/git"
	"github.com/cdnjs/tools/npm"
	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/util"
)

var (
	// Store the number of validation errors
	errCount uint = 0

	// initialize standard debug logger
	logger = util.GetCheckerLogger()
)

func main() {
	flag.Parse()
	subcommand := flag.Arg(0)

	if util.IsDebug() {
		fmt.Println("Running in debug mode")
	}

	if subcommand == "lint" {
		lintPackage(flag.Arg(1))

		if errCount > 0 {
			os.Exit(1)
		}
		return
	}

	if subcommand == "show-files" {
		showFiles(flag.Arg(1))

		if errCount > 0 {
			os.Exit(1)
		}
		return
	}

	panic("unknown subcommand")
}

// Represents a version of a package,
// which could be a git version, npm version, etc.
type version interface {
	Get() string                    // Get the version as a string
	Download(...interface{}) string // Download a version, returning the download dir
	Clean(string)                   // Clean a download dir
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
			npmVersions := npm.GetVersions(pckg.Autoupdate.Target)
			sort.Sort(npm.ByNpmVersion(npmVersions))

			// cast to interface
			for _, v := range npmVersions {
				versions = append(versions, v)
			}

			// download into temp dir
			downloadDir = npm.DownloadTar(ctx, npmVersions[0].Tarball)

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
			gitVersions := git.GetVersions(ctx, pckg, packageGitDir)
			sort.Sort(git.ByGitVersion(gitVersions))

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
	if len(versions) > util.IMPORT_ALL_MAX_VERSIONS {
		versions = versions[:util.IMPORT_ALL_MAX_VERSIONS]
	}

	// print info for first src version
	printCurrentVersion(ctx, pckg, downloadDir, versions[0])

	// print aggregate info for the few last src versions
	printLastVersions(ctx, pckg, downloadDir, versions[1:])
}

// Prints the files of a package version, outputting debug
// messages if no valid files are present.
func printCurrentVersion(ctx context.Context, p *packages.Package, dir string, v version) {
	filesToCopy := p.NpmFilesFrom(dir)

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
				errormsg += fmt.Sprintf("[Click here to debug your glob pattern `%s`](%s).\n", pattern, makeGlobDebugLink(pattern, dir))
			}
		}
		err(ctx, errormsg)
		return
	}

	fmt.Printf("```\n")
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
		downloadDir := version.Download(ctx, p, dir)
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

	if pckg.Version != "" {
		err(ctx, shouldBeEmpty(".version"))
	}

	// if pckg.NpmName != nil && *pckg.NpmName == "" {
	// 	err(ctx, shouldBeEmpty(".NpmName"))
	// }

	// if len(pckg.NpmFileMap) > 0 {
	// 	err(ctx, shouldBeEmpty(".NpmFileMap"))
	// }

	if pckg.Autoupdate != nil {
		switch pckg.Autoupdate.Source {
		case "npm":
			{
				// check that it exists
				if !npm.Exists(pckg.Autoupdate.Target) {
					err(ctx, "package doesn't exist on npm")
					break
				}

				// check if it has enough downloads
				if md := npm.GetMonthlyDownload(pckg.Autoupdate.Target); md.Downloads < util.MIN_NPM_MONTHLY_DOWNLOADS {
					err(ctx, fmt.Sprintf("package download per month on npm is under %d", util.MIN_NPM_MONTHLY_DOWNLOADS))
				}
			}
		case "git":
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

func shouldBeEmpty(name string) string {
	return fmt.Sprintf("%s should be empty", name)
}

func shouldNotBeEmpty(name string) string {
	return fmt.Sprintf("%s should be specified", name)
}
