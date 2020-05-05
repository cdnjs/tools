package npm

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"

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

func Untar(dst string, r io.Reader) error {
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

		// the following switch could also be done using fi.Mode(), not sure if there
		// a benefit of using one vs. the other.
		// fi := header.FileInfo()

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

// Extract the tarball url in a temporary location
func DownloadTar(ctx context.Context, url string) string {
	dest, err := ioutil.TempDir("", "npmtarball")
	util.Check(err)

	util.Debugf(ctx, "download %s in %s", url, dest)

	resp, err := http.Get(url)
	util.Check(err)

	defer resp.Body.Close()

	util.Check(Untar(dest, resp.Body))
	return dest
}
