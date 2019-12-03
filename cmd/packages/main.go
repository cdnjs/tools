package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
)

const bucketName = "cdnjs-assets"

func getCredentialsFile() string {
	home, err := os.UserHomeDir()
	check(err)

	return path.Join(home, "google_storage_cdnjs_assets.json")
}

func check(err interface{}) {
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()
	subcommand := flag.Arg(0)

	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(getCredentialsFile()))
	check(err)

	bkt := client.Bucket(bucketName)

	obj := bkt.Object("package.min.js")

	if subcommand == "set" {
		w := obj.NewWriter(ctx)
		_, err := io.Copy(w, os.Stdin)
		check(err)
		check(w.Close())
		check(obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader))
		fmt.Println("Uploaded package.min.js")
		return
	}

	panic("unknown subcommand")
}
