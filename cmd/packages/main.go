package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/xtuc/cdnjs-go/cloudstorage"
	"github.com/xtuc/cdnjs-go/util"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
)

func main() {
	flag.Parse()
	subcommand := flag.Arg(0)

	ctx := context.Background()

	bkt, err := cloudstorage.GetBucket(ctx)
	util.Check(err)

	obj := bkt.Object("package.min.js")

	if subcommand == "set" {
		w := obj.NewWriter(ctx)
		_, err := io.Copy(w, os.Stdin)
		util.Check(err)
		util.Check(w.Close())
		util.Check(obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader))
		fmt.Println("Uploaded package.min.js")
		return
	}

	panic("unknown subcommand")
}
