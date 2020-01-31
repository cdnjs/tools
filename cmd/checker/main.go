package main

import (
	"context"
	"flag"
	"fmt"
	"path"

	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/util"
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
		return
	}

	panic("unknown subcommand")
}

func lintPackage(name string) {
	path := path.Join(packages.PACKAGES_PATH, name, "package.json")

	ctx := util.ContextWithName(path)

	pckg, err := packages.ReadPackageJSON(ctx, path)
	util.Check(err)

	checkNotEmpty(ctx, ".name", pckg.Name)
	checkNotEmpty(ctx, ".version", pckg.Version)
}

func checkNotEmpty(ctx context.Context, name string, v string) {
	if v == "" {
		util.Printf(ctx, name+" is empty\n")
	}
}

func checkEmpty(ctx context.Context, name string, v string) {
	if v != "" {
		util.Printf(ctx, name+" is specified\n")
	}
}
