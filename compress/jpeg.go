package compress

import (
	"context"
	"os/exec"
	"syscall"

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
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pgid: 0}
	util.Debugf(ctx, "compress: run %s\n", cmd)
	out := util.CheckCmd(cmd.CombinedOutput())
	util.Debugf(ctx, "%s\n", out)
}
