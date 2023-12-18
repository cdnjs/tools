package compress

import (
	"context"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

// Extensions the compression handle
var (
	UGLIFYJS = "/node_modules/uglify-js/bin/uglifyjs"
	UGLIFYES = "/node_modules/uglify-es/bin/uglifyjs"
)

// Js performs a compression of the file.
func Js(ctx context.Context, file string) *string {
	if strings.HasSuffix(file, ".min.js") {
		log.Printf("%s is already compressed\n", file)
		return nil
	}

	ext := path.Ext(file)
	outfile := file[0:len(file)-len(ext)] + ".min.js"

	if _, err := os.Stat(outfile); err == nil {
		log.Printf("%s already has corresponding compressed file\n", outfile)
		return nil
	}

	args := []string{
		"--mangle",
		"--compress",
		"if_return=true",
		"-o", outfile,
		file,
	}

	// try with uglifyjs, if it fails retry with uglifyes
	cmd := exec.Command(UGLIFYJS, args...)
	version := getNpmVersion("uglify-js")
	log.Printf("compress: run %s (%s) %s\n", UGLIFYJS, version, args)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("failed with %s: %s\n", err, out)

		cmd := exec.Command(UGLIFYES, args...)
		version := getNpmVersion("uglify-es")
		log.Printf("compress: run %s (%s) %s\n", UGLIFYES, version, args)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("failed with %s: %s\n", err, out)
			return nil
		}
	}
	return &outfile
}
