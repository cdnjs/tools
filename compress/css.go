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
	CssExt = map[string]bool{
		".css": true,
	}

	BASE_PATH = util.GetEnv("BOT_BASE_PATH")
	CLEANCSS  = path.Join(BASE_PATH, "tools", "node_modules/clean-css-cli/bin/cleancss")
)

// Perform a compression of the file
func CompressCss(ctx context.Context, file string) {
	outfile := strings.ReplaceAll(file, ".css", ".min.css")

	// compressed file already exists, ignore
	if _, err := os.Stat(outfile); !os.IsNotExist(err) {
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

	cmd := exec.Command(CLEANCSS, args...)
	util.Debugf(ctx, "compress: run %s\n", cmd)
	out := util.CheckCmd(cmd.CombinedOutput())
	util.Debugf(ctx, "%s\n", out)
}
