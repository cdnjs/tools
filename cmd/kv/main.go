package main

import (
	"flag"
	"fmt"

	"github.com/cdnjs/tools/kv"
	"github.com/cdnjs/tools/sentry"

	"github.com/cdnjs/tools/util"
)

var (
	// initialize standard debug logger
	logger = util.GetStandardLogger()
)

func init() {
	sentry.Init()
}

func main() {
	defer sentry.PanicHandler()
	var metaOnly, srisOnly, filesOnly, ungzip, unbrotli bool
	flag.BoolVar(&metaOnly, "meta-only", false, "If set, only version metadata is uploaded to KV (no files, no SRIs).")
	flag.BoolVar(&srisOnly, "sris-only", false, "If set, only file SRIs are uploaded to KV (no files, no metadata).")
	flag.BoolVar(&filesOnly, "files-only", false, "If set, only files are uploaded to KV (no metadata, no SRIs).")
	flag.BoolVar(&ungzip, "ungzip", false, "If set, the file content will be decompressed with gzip.")
	flag.BoolVar(&unbrotli, "unbrotli", false, "If set, the file content will be decompressed with brotli.")
	flag.Parse()

	if util.IsDebug() {
		fmt.Println("Running in debug mode")
	}

	switch subcommand := flag.Arg(0); subcommand {
	case "upload":
		{
			count := 0
			for _, b := range []bool{metaOnly, srisOnly, filesOnly} {
				if b {
					count++
				}
			}
			if count > 1 {
				panic("can only set one of -meta-only, -sris-only, -files-only")
			}

			pckgs := flag.Args()[1:]
			if len(pckgs) == 0 {
				panic("no packages specified")
			}

			kv.InsertFromDisk(logger, pckgs, metaOnly, srisOnly, filesOnly)
		}
	case "upload-aggregate":
		{
			pckgs := flag.Args()[1:]
			if len(pckgs) == 0 {
				panic("no packages specified")
			}

			kv.InsertAggregateMetadataFromScratch(logger, pckgs)
		}
	case "aggregate-packages":
		{
			kv.OutputAllAggregatePackages()
		}
	case "packages":
		{
			kv.OutputAllPackages()
		}
	case "file":
		{
			if ungzip && unbrotli {
				panic("can only set one of -ungzip, -unbrotli")
			}

			file := flag.Arg(1)
			if file == "" {
				panic("no file specified")
			}

			kv.OutputFile(logger, file, ungzip, unbrotli)
		}
	case "files":
		{
			pckg := flag.Arg(1)
			if pckg == "" {
				panic("no package specified")
			}

			kv.OutputAllFiles(logger, pckg)
		}
	case "meta":
		{
			pckg := flag.Arg(1)
			if pckg == "" {
				panic("no package specified")
			}

			kv.OutputAllMeta(logger, pckg)
		}
	case "aggregate":
		{
			pckg := flag.Arg(1)
			if pckg == "" {
				panic("no package specified")
			}

			kv.OutputAggregate(pckg)
		}
	case "sris":
		{
			prefix := flag.Arg(1)
			if prefix == "" {
				panic("no prefix specified") // avoid listing all SRIs
			}

			kv.OutputSRIs(prefix)
		}
	default:
		panic(fmt.Sprintf("unknown subcommand: `%s`", subcommand))
	}
}
