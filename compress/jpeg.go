package compress

import (
	"context"
	"os/exec"

	"github.com/cdnjs/tools/util"
)

// Extensions the compression handle
var JpegExt = map[string]bool{
	".jpg":  true,
	".jpeg": true,
}

// Perform an in-place compression of the file
func CompressJpeg(ctx context.Context, file string) {
	cmd := exec.Command("jpegoptim", file)
	util.Debugf(ctx, "compress: run %s\n", cmd)
	out := util.CheckCmd(cmd.CombinedOutput())
	util.Debugf(ctx, "%s\n", out)
}
