package main

import (
	"flag"
	"fmt"

	"github.com/cdnjs/tools/kv"

	"github.com/cdnjs/tools/util"
)

var (
	// initialize standard debug logger
	logger = util.GetStandardLogger()
)

func main() {
	flag.Parse()

	if util.IsDebug() {
		fmt.Println("Running in debug mode")
	}

	switch subcommand := flag.Arg(0); subcommand {
	case "traverse":
		{
			kv.Traverse()
		}
	case "test":
		{
			// create context with file path prefix, checker logger
			ctx := util.ContextWithEntries(util.GetCheckerEntries("", logger)...)
			const maxPkgs = 3
			kv.TestInsertingPkgs(ctx, maxPkgs)
		}
	default:
		panic(fmt.Sprintf("unknown subcommand: `%s`", subcommand))
	}
}
