package compress

import (
	"context"
	"os/exec"

	"github.com/cdnjs/tools/util"
)

// Extensions the compression handle
var PngExt = map[string]bool{
	".png": true,
}

// Perform an in-place compression of the file
func CompressPng(ctx context.Context, file string) {
	args := []string{
		"--iterations=60",
		"--keepchunks=iCCP",
		"--lossy_transparent",
		"--splitting=3",
		"-my",
		file, file,
	}

	cmd := exec.Command("zopflipng", args...)
	util.Debugf(ctx, "compress: run %s\n", cmd)
	out := util.CheckCmd(cmd.CombinedOutput())
	util.Debugf(ctx, "%s\n", out)
}
