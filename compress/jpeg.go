package compress

import (
	"context"
	"log"
	"os/exec"

	"github.com/cdnjs/tools/util"
)

// Jpeg performs an in-place compression of the file.
func Jpeg(ctx context.Context, file string) {
	cmd := exec.Command("jpegoptim", file)
	log.Printf("compress: run %s\n", cmd)
	out := util.CheckCmd(cmd.CombinedOutput())
	util.Debugf(ctx, "%s\n", out)
}
