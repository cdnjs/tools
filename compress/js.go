package compress

import (
	"context"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/cdnjs/tools/util"
)

// Extensions the compression handle
var (
	UGLIFYJS = "/node_modules/uglify-js/bin/uglifyjs"
	UGLIFYES = "/node_modules/uglify-es/bin/uglifyjs"
)

// Js performs a compression of the file.
func Js(ctx context.Context, file string) {
	ext := path.Ext(file)
	outfile := file[0:len(file)-len(ext)] + ".min.js"

	// compressed file already exists, ignore
	if _, err := os.Stat(outfile); err == nil {
		util.Debugf(ctx, "compressed file already exists: %s\n", outfile)
		return
	}

	// Already minified, ignore
	if strings.HasSuffix(file, ".min.js") {
		return
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
	util.Debugf(ctx, "compress: run %s\n", cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		util.Debugf(ctx, "failed with %s: %s\n", err, out)

		cmd := exec.Command(UGLIFYES, args...)
		util.Debugf(ctx, "compress: run %s\n", cmd)
		out, err := cmd.CombinedOutput()
		if err != nil {
			util.Debugf(ctx, "failed with %s: %s\n", err, out)
		}
	}
}
