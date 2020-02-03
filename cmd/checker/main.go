package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/util"
)

var (
	// Store the number of validation errors
	validationErrorCount uint = 0
)

func main() {
	flag.Parse()
	subcommand := flag.Arg(0)

	if util.IsDebug() {
		fmt.Println("Running in debug mode")
	}

	if subcommand == "lint" {
		names := flag.Args()[1:]

		for _, name := range names {
			lintPackage(name)
		}

		if validationErrorCount > 0 {
			fmt.Printf("%d linting error(s)\n", validationErrorCount)
			os.Exit(1)
		}
		return
	}

	panic("unknown subcommand")
}

func lintPackage(name string) {
	path := path.Join(packages.PACKAGES_PATH, name, "package.json")

	ctx := util.ContextWithName(path)

	pckg, readerr := packages.ReadPackageJSON(ctx, path)
	util.Check(readerr)

	if util.IsDebug() {
		fmt.Printf("Linting %s...\n", name)
	}

	if pckg.Name == "" {
		err(ctx, shouldNotBeEmpty(".name"))
	}

	if pckg.Version == "" {
		err(ctx, shouldNotBeEmpty(".version"))
	}

	if pckg.NpmName != nil && *pckg.NpmName == "" {
		err(ctx, shouldBeEmpty(".NpmName"))
	}

	if len(pckg.NpmFileMap) > 0 {
		err(ctx, shouldBeEmpty(".NpmFileMap"))
	}

	if pckg.Autoupdate != nil {
		if pckg.Autoupdate.Source != "npm" && pckg.Autoupdate.Source != "git" {
			err(ctx, "Unsupported .autoupdate.source: "+pckg.Autoupdate.Source)
		}
	} else {
		warn(ctx, ".autoupdate should not be null. Package will never auto-update")
	}
}

func err(ctx context.Context, s string) {
	util.Printf(ctx, "error: "+s)
	validationErrorCount += 1
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
