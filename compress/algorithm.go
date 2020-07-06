package compress

import (
	"bytes"
	"compress/gzip"

	"github.com/andybalholm/brotli"
	"github.com/cdnjs/tools/util"
)

// Brotli11 returns a brotli compressed file as bytes
// at optimal compression (quality 11).
func Brotli11(uncompressed []byte) []byte {
	var b bytes.Buffer

	w := brotli.NewWriterLevel(&b, brotli.BestCompression)

	_, err := w.Write(uncompressed)
	util.Check(err)
	util.Check(w.Close())

	return b.Bytes()
}

// Gzip9 returns a gzip compressed file as bytes
// at optimal compression (level 9).
func Gzip9(uncompressed []byte) []byte {
	var b bytes.Buffer

	w, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	util.Check(err)

	_, err = w.Write(uncompressed)
	util.Check(err)
	util.Check(w.Close())

	return b.Bytes()
}
