package npm

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cdnjs/tools/util"
)

func removePackageDir(path string) string {
	if len(path) < 8 {
		return path
	}
	if path[0:8] == "package/" {
		return path[8:]
	}
	return path
}

// Untar uncompresses a tar at a destination.
func Untar(ctx context.Context, dst string, r io.Reader) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			return nil

		// return any other error
		case err != nil:
			return err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(dst, removePackageDir(header.Name))

		// While this check prevents some path walking, a path that contains
		// UTF-8 might bypass that check. For instance \u002e\u002e\u2215etc/passwd
		// will be joined to the dst directory and appear safe. However, the open
		// file syscall will interpret the UTF-8 and effectively allow path walking
		// again. In production, the bot must run in a sandboxed environment.
		if !strings.HasPrefix(target, dst) {
			util.Warnf(ctx, "Unsafe file located outside `%s` with name: `%s`", dst, header.Name)
			continue
		}

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			{
				if _, err := os.Stat(target); err != nil {
					if err := os.MkdirAll(target, 0755); err != nil {
						return err
					}
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			{
				dir := path.Dir(target)
				if _, err := os.Stat(dir); err != nil {
					if err := os.MkdirAll(dir, 0755); err != nil {
						return err
					}
				}

				f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
				if err != nil {
					return err
				}

				// copy over contents
				if _, err := io.Copy(f, tr); err != nil {
					return err
				}

				// manually close here after each file operation; defering would cause each file close
				// to wait until all operations have completed.
				f.Close()
			}
		}
	}
}

func DownloadTar(ctx context.Context, url string) bytes.Buffer {
	util.Debugf(ctx, "download %s", url)

	resp, err := http.Get(url)
	util.Check(err)
	defer resp.Body.Close()

	var buff bytes.Buffer
	buff.ReadFrom(resp.Body)

	return buff
}
