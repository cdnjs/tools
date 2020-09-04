package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/cdnjs/tools/algolia"
	"github.com/cdnjs/tools/cloudstorage"
	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/sentry"
	"github.com/cdnjs/tools/util"
)

// PackagesJSON is used to wrap around a slice of []Packages
// when JSON unmarshalling.
type PackagesJSON struct {
	Packages []packages.Package `json:"packages"`
}

func main() {
	defer sentry.PanicHandler()
	flag.Parse()

	switch subcommand := flag.Arg(0); subcommand {
	case "update":
		{
			fmt.Printf("Downloading package.min.js...")
			b := getPackagesBuffer()
			fmt.Printf("Ok\n")

			var j PackagesJSON
			util.Check(json.Unmarshal(b.Bytes(), &j))

			fmt.Printf("Building index...\n")

			algoliaClient := algolia.GetClient()
			tmpIndex := algolia.GetTmpIndex(algoliaClient)

			for _, p := range j.Packages {
				fmt.Printf("%s: ", *p.Name)
				util.Check(algolia.IndexPackage(p, tmpIndex))
				fmt.Printf("Ok\n")
			}
			fmt.Printf("Ok\n")

			fmt.Printf("Promoting index to production...")
			algolia.PromoteIndex(algoliaClient)
			fmt.Printf("Ok\n")
		}
	default:
		panic(fmt.Sprintf("unknown subcommand: `%s`", subcommand))
	}
}

func getPackagesBuffer() bytes.Buffer {
	ctx := context.Background()

	bkt, err := cloudstorage.GetAssetsBucket(ctx)
	util.Check(err)

	obj := bkt.Object("package.min.js")

	r, err := obj.NewReader(ctx)
	util.Check(err)
	defer r.Close()

	var b bytes.Buffer

	_, copyerr := io.Copy(bufio.NewWriter(&b), r)
	util.Check(copyerr)

	return b
}
