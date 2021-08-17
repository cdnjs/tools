package version

import (
	"bytes"
	"context"
	"log"
	"net/http"

	"github.com/cdnjs/tools/util"
)

func DownloadTar(ctx context.Context, v Version) bytes.Buffer {
	if v.Tarball == "" {
		panic("no tarball url provided for " + v.Version)
	}
	log.Printf("download %s\n", v.Tarball)

	resp, err := http.Get(v.Tarball)
	util.Check(err)
	defer resp.Body.Close()

	var buff bytes.Buffer
	_, err = buff.ReadFrom(resp.Body)
	util.Check(err)

	return buff
}
