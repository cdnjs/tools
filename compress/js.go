package compress

import (
	"context"
	"os/exec"
	"path"
	"strings"

	"github.com/cdnjs/tools/util"
)

// Extensions the compression handle
var (
	JsExt = map[string]bool{
		".js": true,
	}

	UGLIFYJS = path.Join(BASE_PATH, "tools", "node_modules/uglify-js/bin/uglifyjs")
	UGLIFYES = path.Join(BASE_PATH, "tools", "node_modules/uglify-es/bin/uglifyjs")
)

// Perform a compression of the file
func CompressJs(ctx context.Context, file string) {
	// Already minified, ignore
	if strings.HasSuffix(file, ".min.js") {
		return
	}

	args := []string{
		"--mangle",
		"--compress",
		"if_return=true",
		"-o", strings.ReplaceAll(file, ".js", ".min.js"),
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
