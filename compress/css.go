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
	CSSExt = map[string]bool{
		".css": true,
	}

	basePath = util.GetEnv("BOT_BASE_PATH")
	cleanCSS = path.Join(basePath, "tools", "node_modules/clean-css-cli/bin/cleancss")
)

// CSS performs a compression of the file.
func CSS(ctx context.Context, file string) {
	ext := path.Ext(file)
	outfile := file[0:len(file)-len(ext)] + ".min.css"

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

	cmd := exec.Command(cleanCSS, args...)
	util.Debugf(ctx, "compress: run %s\n", cmd)
	out := util.CheckCmd(cmd.CombinedOutput())
	util.Debugf(ctx, "%s\n", out)
}
