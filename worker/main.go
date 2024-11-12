package main

import (
	"context"
	b64 "encoding/base64"

	algolia_pump "github.com/cdnjs/tools/functions/algolia-pump"
)

func main() {}

//export RunAlgoliaPump
func RunAlgoliaPump() {
	ctx := context.TODO()
	bucket := "foo"
	file := "foo"
	pkgName := "foo"
	currVersion := "foo"
	config := b64.URLEncoding.EncodeToString([]byte("{\n  \"autoupdate\": {\n    \"fileMap\": [\n      {\n        \"basePath\": \"\",\n        \"files\": [\n          \"*\"\n        ]\n      }\n    ],\n    \"source\": \"npm\",\n    \"target\": \"hi-sven\"\n  },\n  \"description\": \"Say hi to Sven\",\n  \"filename\": \"index.js\",\n  \"keywords\": [\n    \"hello\"\n  ],\n  \"license\": \"MIT\",\n  \"name\": \"hi-sven\",\n  \"repository\": {\n    \"type\": \"git\",\n    \"url\": \"git+https://github.com/xtuc/hi-sven.git\"\n  },\n  \"authors\": [\n    {\n      \"name\": \"Sven Sauleau\",\n      \"email\": \"sven@sauleau.com\"\n    }\n  ]\n}"))

	panic(algolia_pump.Run(ctx, bucket, file, pkgName, currVersion, config))
}
