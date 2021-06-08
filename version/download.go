package version

import (
	"bytes"
	"context"
	"log"
	"net/http"

	"github.com/cdnjs/tools/util"
)

func DownloadTar(ctx context.Context, v Version) bytes.Buffer {
	log.Printf("download %s\n", v.Tarball)

	resp, err := http.Get(v.Tarball)
	util.Check(err)
	defer resp.Body.Close()

	var buff bytes.Buffer
	buff.ReadFrom(resp.Body)

	return buff
}
