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
	case "upload":
		{
			pckgs := flag.Args()[1:]
			if len(pckgs) == 0 {
				panic("no packages specified")
			}

			kv.InsertFromDisk(logger, pckgs)
		}
	default:
		panic(fmt.Sprintf("unknown subcommand: `%s`", subcommand))
	}
}
