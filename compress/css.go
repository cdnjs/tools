package compress

import (
	"context"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/cdnjs/tools/util"
)

// Extensions the compression handle
var (
	cleanCSS = "/node_modules/clean-css-cli/bin/cleancss"
)

// CSS performs a compression of the file.
func CSS(ctx context.Context, file string) {
	ext := path.Ext(file)
	outfile := file[0:len(file)-len(ext)] + ".min.css"

	// compressed file already exists, ignore
	if _, err := os.Stat(outfile); err == nil {
		util.Debugf(ctx, "compressed file already exists: %s\n", outfile)
		return
	}

	// Already minified, ignore
	if strings.HasSuffix(file, ".min.css") {
		return
	}

	args := []string{
		"--compatibility",
		"--s0",
		"-o", outfile,
		file,
	}

	cmd := exec.Command(cleanCSS, args...)
	log.Printf("compress: run %s\n", cmd)

	if bytes, err := cmd.CombinedOutput(); err != nil {
		util.Debugf(ctx, "Failed to compress CSS: %v\n", err)
	} else {
		util.Debugf(ctx, "%s\n", bytes)
	}
}
