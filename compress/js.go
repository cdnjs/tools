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
	ext := path.Ext(file)
	outfile := file[0:len(file)-len(ext)] + ".min.js"

	// compressed file already exists, ignore
	if _, err := os.Stat(outfile); err == nil {
		log.Printf("compressed file already exists: %s\n", outfile)
		return nil
	}

	// Already minified, ignore
	if strings.HasSuffix(file, ".min.js") {
		log.Printf("%s.min.js compressed file already exists\n", file)
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
	log.Printf("compress: run %s\n", cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("failed with %s: %s\n", err, out)

		cmd := exec.Command(UGLIFYES, args...)
		log.Printf("compress: run %s\n", cmd)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("failed with %s: %s\n", err, out)
			return nil
		}
	}
	return &outfile
}
