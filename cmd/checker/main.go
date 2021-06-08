package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/cdnjs/tools/git"
	"github.com/cdnjs/tools/npm"
	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/sandbox"
	"github.com/cdnjs/tools/util"
	"github.com/cdnjs/tools/version"

	"github.com/pkg/errors"
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

	switch subcommand := flag.Arg(0); subcommand {
	case "lint":
		{
			for _, path := range flag.Args()[1:] {
				if err := lintPackage(path, noPathValidation); err != nil {
					log.Fatalf("failed to lint package: %s\n", err)
				}
			}

			if errCount > 0 {
				os.Exit(1)
			}
		}
	case "show-files":
		{
			if err := showFiles(flag.Arg(1), noPathValidation); err != nil {
				log.Fatalf("failed to show files: %s\n", err)
			}

			if errCount > 0 {
				os.Exit(1)
			}
		}
	default:
		panic(fmt.Sprintf("unknown subcommand: `%s`", subcommand))
	}
}

func processVersion(ctx context.Context, pckg *packages.Package, v version.Version) (string, error) {
	inDir, outDir, err := sandbox.Setup()
	if err != nil {
		return outDir, errors.Wrap(err, "failed to setup sandbox")
	}
	defer os.RemoveAll(inDir)

	buff := version.DownloadTar(ctx, v)

	dst, err := os.Create(path.Join(inDir, "new-version.tgz"))
	if err != nil {
		return outDir, errors.Wrap(err, "could not write tmp file")
	}
	defer dst.Close()
	if _, err := dst.Write(buff.Bytes()); err != nil {
		return outDir, errors.Wrap(err, "could not write new version in sandbox")
	}

	if err := writeConfig(inDir, pckg); err != nil {
		return outDir, errors.Wrap(err, "failed to write configuration")
	}

	name := fmt.Sprintf("%s_%s", *pckg.Name, v.Version)
	logs, err := sandbox.Run(ctx, name, inDir, outDir)
	if err != nil {
		return outDir, errors.Wrap(err, "failed to run sandbox")
	}
	log.Println("logs", len(logs), logs)

	return outDir, nil
}

func showFiles(pckgPath string, noPathValidation bool) error {
	// create context with file path prefix, checker logger
	ctx := util.ContextWithEntries(util.GetCheckerEntries(pckgPath, logger)...)

	// parse *Package from JSON
	pckg, err := parseHumanPackage(ctx, pckgPath, noPathValidation)
	if err != nil {
		return errors.Wrap(err, "could not parse package")
	}
	if pckg == nil {
		return nil
	}

	if err := sandbox.Init(ctx); err != nil {
		log.Fatalf("failed to init sandbox: %s", err)
	}

	// autoupdate exists, download latest versions based on source
	src := *pckg.Autoupdate.Source

	var versions []version.Version

	switch src {
	case "npm":
		{
			// get npm versions and sort
			versions, _ = npm.GetVersions(ctx, pckg.Autoupdate)
			sort.Sort(version.ByDate(versions))
		}
	case "git":
		{
			var err error
			// get git versions and sort
			versions, err = git.GetVersions(ctx, pckg.Autoupdate)
			if err != nil {
				return errors.Wrap(err, "failed to retrieve git versions")
			}
			sort.Sort(version.ByDate(versions))
		}
	default:
		{
			panic(fmt.Sprintf("unknown autoupdate source: %s", src))
		}
	}

	// download into temp dir
	if len(versions) > 0 {
		// print info for first src version
		if err := printMostRecentVersion(ctx, pckg, versions[0]); err != nil {
			return errors.Wrap(err, "could not print most recent version")
		}

		// print aggregate info for the few last src versions
		if err := printLastVersions(ctx, pckg, versions[1:]); err != nil {
			return errors.Wrap(err, "could not print most last versions")
		}
	} else {
		showErr(ctx, "no version found on "+src)
	}
	return nil
}

// Try to parse a *Package, outputting ci errors/warnings.
// If there is an issue, *Package will be nil.
func parseHumanPackage(ctx context.Context, pckgPath string, noPathValidation bool) (*packages.Package, error) {
	if !noPathValidation {
		// check package path matches regex
		matches := pckgPathRegex.FindStringSubmatch(pckgPath)
		if matches == nil {
			showErr(ctx, fmt.Sprintf("package path `%s` does not match %s", pckgPath, pckgPathRegex.String()))
			return nil, nil
		}

		// check the package is going into the correct folder
		// (ex. My-Package -> packages/m/My-Package.json)
		actualDir, pckgName := matches[1], matches[2]
		expectedDir := strings.ToLower(string(pckgName[0]))
		if actualDir != expectedDir {
			showErr(ctx, fmt.Sprintf("package `%s` must go into `%s` dir, not `%s` dir", pckgName, expectedDir, actualDir))
			return nil, nil
		}
	}

	bytes, err := ioutil.ReadFile(pckgPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read package file")
	}

	// parse package JSON
	pckg, readerr := packages.ReadHumanJSONBytes(ctx, pckgPath, bytes)
	if readerr != nil {
		if invalidHumanErr, ok := readerr.(packages.InvalidSchemaError); ok {
			// output all schema errors
			for _, resErr := range invalidHumanErr.Result.Errors() {
				showErr(ctx, resErr.String())
			}
		} else {
			showErr(ctx, readerr.Error())
		}
		return nil, nil
	}

	checkFilename(ctx, pckg)
	return pckg, nil
}

func filewalker(basedir string, files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrap(err, "failed to walk fs")
		}
		ext := filepath.Ext(path)
		if ext == ".gz" {
			path = strings.ReplaceAll(path, ".gz", "")
			path = strings.ReplaceAll(path, basedir+"/", "")
			*files = append(*files, path)
		}
		return nil
	}
}

// Prints the files of a package version, outputting debug
// messages if no valid files are present.
func printMostRecentVersion(ctx context.Context, p *packages.Package, v version.Version) error {
	fmt.Printf("\nmost recent version: %s\n", v.Version)

	outDir, err := processVersion(ctx, p, v)
	if err != nil {
		log.Fatalf("failed to process version: %s", err)
	}
	defer os.RemoveAll(outDir)

	var files []string

	err = filepath.Walk(outDir, filewalker(outDir, &files))
	if err != nil {
		return errors.Wrap(err, "could not inspect sandbox output")
	}

	if len(files) == 0 {
		errormsg := fmt.Sprintf("No files will be published for version %s.\n", v.Version)
		showErr(ctx, errormsg)
		return nil
	}

	var filenameFound bool

	fmt.Printf("\n```\n")
	for _, file := range files {
		fmt.Printf("%s\n", file)
		if p.Filename != nil && !filenameFound && file == *p.Filename {
			filenameFound = true
		}
	}
	fmt.Printf("```\n")

	if p.Filename != nil && !filenameFound {
		showErr(ctx, fmt.Sprintf("Filename `%s` not found in most recent version `%s`.\n", *p.Filename, v.Version))
	}
	return nil
}

// Prints the matching files of a number of last versions.
// Each previous version will be downloaded and cleaned up if necessary.
// For example, a temporary directory may be downloaded and then removed later.
func printLastVersions(ctx context.Context, p *packages.Package, versions []version.Version) error {
	// limit versions
	if len(versions) > util.ImportAllMaxVersions {
		versions = versions[:util.ImportAllMaxVersions]
	}

	fmt.Printf("\n%d last version(s):\n", len(versions))
	for _, version := range versions {
		outDir, err := processVersion(ctx, p, version)
		if err != nil {
			log.Fatalf("failed to process version: %s", err)
		}

		var files []string

		err = filepath.Walk(outDir, filewalker(outDir, &files))
		if err != nil {
			return errors.Wrap(err, "could not inspect sandbox output")
		}

		fmt.Printf("- %s: %d file(s) matched", version.Version, len(files))
		if len(files) > 0 {
			fmt.Printf(" :heavy_check_mark:\n")
		} else {
			fmt.Printf(" :heavy_exclamation_mark:\n")
		}

		os.RemoveAll(outDir)
	}
	return nil
}

func checkGitHubPopularity(ctx context.Context, pckg *packages.Package) bool {
	if !strings.Contains(*pckg.Repository.URL, "github.com") {
		return false
	}

	if s := git.GetGitHubStars(*pckg.Repository.URL); s.Stars < util.MinGitHubStars {
		showWarn(ctx, fmt.Sprintf("stars on GitHub is under %d", util.MinGitHubStars))
		return false
	}
	return true
}

func checkFilename(ctx context.Context, pckg *packages.Package) {
	// warn if filename is not present
	// current, only a few packages have exceptions
	// that allow them to have missing filenames
	if pckg.Filename == nil {
		showWarn(ctx, "filename is missing")
	}
}

func lintPackage(pckgPath string, noPathValidation bool) error {
	// create context with file path prefix, checker logger
	ctx := util.ContextWithEntries(util.GetCheckerEntries(pckgPath, logger)...)

	// parse *Package from JSON
	pckg, err := parseHumanPackage(ctx, pckgPath, noPathValidation)
	if err != nil {
		return errors.Wrap(err, "could not parse package")
	}
	if pckg == nil {
		return nil
	}

	switch *pckg.Autoupdate.Source {
	case "npm":
		{
			// check that it exists
			if !npm.Exists(*pckg.Autoupdate.Target) {
				showErr(ctx, "package doesn't exist on npm")
				break
			}

			// check if it has enough downloads
			if md := npm.GetMonthlyDownload(*pckg.Autoupdate.Target); md.Downloads < util.MinNpmMonthlyDownloads {
				if !checkGitHubPopularity(ctx, pckg) {
					showWarn(ctx, fmt.Sprintf("package download per month on npm is under %d", util.MinNpmMonthlyDownloads))
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

	log.Printf("%s lint OK\n", pckgPath)
	return nil
}

// wrapper around outputting a checker error
func showErr(ctx context.Context, s string) {
	util.Errf(ctx, s)
	errCount++
}

// wrapper around outputting a checker warning
func showWarn(ctx context.Context, s string) {
	util.Warnf(ctx, s)
}

func writeConfig(dstDir string, pkg *packages.Package) error {
	config := []byte(pkg.String())
	if err := ioutil.WriteFile(path.Join(dstDir, "config.json"), config, 0644); err != nil {
		return errors.Wrap(err, "could not write config file")
	}
	return nil
}
