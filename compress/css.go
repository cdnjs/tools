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
	cleanCSS = "/node_modules/clean-css-cli/bin/cleancss"
)

// CSS performs a compression of the file.
func CSS(ctx context.Context, file string) *string {
	ext := path.Ext(file)
	outfile := file[0:len(file)-len(ext)] + ".min.css"

	// compressed file already exists, ignore
	if _, err := os.Stat(outfile); err == nil {
		log.Printf("%s already has a compressed version: %s\n", file, outfile)
		return nil
	}

	// Already minified, ignore
	if strings.HasSuffix(file, ".min.css") {
		return nil
	}

	args := []string{
		"--compatibility",
		"--s0",
		"-o", outfile,
		file,
	}

	cmd := exec.Command(cleanCSS, args...)
	log.Printf("compress: run %s\n", cmd)

	if _, err := cmd.CombinedOutput(); err != nil {
		log.Printf("Failed to compress CSS: %v\n", err)
		return nil
	}
	return &outfile
}
