package compress

import (
	"context"
	"os/exec"

	"github.com/cdnjs/tools/util"
)

// JpegExt are jpeg extensions the compression handles.
var JpegExt = map[string]bool{
	".jpg":  true,
	".jpeg": true,
}

// Jpeg performs an in-place compression of the file.
func Jpeg(ctx context.Context, file string) {
	cmd := exec.Command("jpegoptim", file)
	util.Debugf(ctx, "compress: run %s\n", cmd)
	out := util.CheckCmd(cmd.CombinedOutput())
	util.Debugf(ctx, "%s\n", out)
}
