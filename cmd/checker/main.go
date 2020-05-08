package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"

	"github.com/cdnjs/tools/npm"
	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/util"
)

var (
	// Store the number of validation errors
	errCount uint = 0
)

func main() {
	flag.Parse()
	subcommand := flag.Arg(0)

	if util.IsDebug() {
		fmt.Println("Running in debug mode")
	}

	// change output for readability in CI
	util.SetLoggerFlag(0)

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

func showFiles(path string) {
	ctx := util.ContextWithName(path)
	pckg, readerr := packages.ReadPackageJSON(ctx, path)
	if readerr != nil {
		err(ctx, readerr.Error())
		return
	}

	fmt.Printf("Preview for `%s`\n", path)

	if pckg.Autoupdate == nil || pckg.Autoupdate.Source != "npm" {
		fmt.Println(pckg.Autoupdate)
		err(ctx, "unsupported autoupdate")
		return
	}

	npmVersions := npm.GetVersions(pckg.Autoupdate.Target)
	if len(npmVersions) == 0 {
		err(ctx, "no version found on npm")
		return
	}

	sort.Sort(npm.ByNpmVersion(npmVersions))

	if len(npmVersions) > util.IMPORT_ALL_MAX_VERSIONS {
		npmVersions = npmVersions[:util.IMPORT_ALL_MAX_VERSIONS]
	}

	// print info for the first version
	firstNpmVersion := npmVersions[0]
	fmt.Printf("Last version (%s):\n", firstNpmVersion.Version)
	{
		tarballDir := npm.DownloadTar(ctx, firstNpmVersion.Tarball)
		filesToCopy := pckg.NpmFilesFrom(tarballDir)

		if len(filesToCopy) == 0 {
			err(ctx, "No files will be published for this version; you can debug using")

			for _, filemap := range pckg.NpmFileMap {
				for _, pattern := range filemap.Files {
					fmt.Printf("[Click here to debug your glob pattern `%s`](%s)\n.", pattern, makeGlobDebugLink(pattern, tarballDir))
				}
			}
			return
		}

		fmt.Printf("```\n")
		for _, file := range filesToCopy {
			fmt.Printf("%s\n", file.To)
		}
		fmt.Printf("```\n")
	}

	// aggregate info for the few last version
	fmt.Printf("%d last versions:\n", util.IMPORT_ALL_MAX_VERSIONS)
	{
		for _, version := range npmVersions {
			tarballDir := npm.DownloadTar(ctx, version.Tarball)
			filesToCopy := pckg.NpmFilesFrom(tarballDir)

			fmt.Printf("- %s: %d file(s) matched", version.Version, len(filesToCopy))
			if len(filesToCopy) > 0 {
				fmt.Printf(":heavy_check_mark:\n")
			} else {
				fmt.Printf(":heavy_exclamation_mark:\n")
			}
		}
	}
}

func makeGlobDebugLink(glob string, dir string) string {
	encodedGlob := url.QueryEscape(glob)
	allTests := ""

	fmt.Printf("dir %s\n", dir)

	util.Check(filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			fmt.Printf("f %s\n", info.Name())
			allTests += "&tests=" + url.QueryEscape(info.Name())
		}
		return nil
	}))

	return fmt.Sprintf("https://www.digitalocean.com/community/tools/glob?comments=true&glob=%s&matches=true%s&tests=", encodedGlob, allTests)
}

func lintPackage(path string) {
	ctx := util.ContextWithName(path)

	util.Debugf(ctx, "Linting %s...\n", path)

	pckg, readerr := packages.ReadPackageJSON(ctx, path)
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
		if pckg.Autoupdate.Source != "npm" && pckg.Autoupdate.Source != "git" {
			err(ctx, "Unsupported .autoupdate.source: "+pckg.Autoupdate.Source)
		}
	} else {
		warn(ctx, ".autoupdate should not be null. Package will never auto-update")
	}

	if pckg.Repository.Repotype != "git" {
		err(ctx, "Unsupported .repository.type: "+pckg.Repository.Repotype)
	}

	if pckg.Autoupdate != nil && pckg.Autoupdate.Source == "npm" {
		if !npm.Exists(pckg.Autoupdate.Target) {
			err(ctx, "package doesn't exists on npm")
		} else {
			counts := npm.GetMonthlyDownload(pckg.Autoupdate.Target)
			if counts.Downloads < 800 {
				err(ctx, "package download per month on npm is under 800")
			}
		}
	}

}

func err(ctx context.Context, s string) {
	util.Printf(ctx, "error: "+s)
	errCount += 1
}

func warn(ctx context.Context, s string) {
	util.Printf(ctx, "warning: "+s)
}

func shouldBeEmpty(name string) string {
	return fmt.Sprintf("%s should be empty\n", name)
}

func shouldNotBeEmpty(name string) string {
	return fmt.Sprintf("%s should be specified\n", name)
}
