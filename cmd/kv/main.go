package main

import (
	"flag"
	"fmt"

	"github.com/cdnjs/tools/kv"

	"github.com/cdnjs/tools/util"
)

var (
	// initialize standard debug logger
	_ = util.GetStandardLogger()
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
	default:
		panic(fmt.Sprintf("unknown subcommand: `%s`", subcommand))
	}
}
