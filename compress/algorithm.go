package compress

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"

	"github.com/cdnjs/tools/util"
)

// Runs an algorithm with a set of arguments,
// and returns its stdout as bytes.
// Note, this function will panic if anything is
// output to stderr.
func runAlgorithm(ctx context.Context, alg string, args ...string) []byte {
	cmd := exec.Command(alg, args...)
	var stdOut, stdErr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdOut, &stdErr

	log.Println(cmd)
	util.Debugf(ctx, "algorithm: run %s\n", cmd)
	err := cmd.Run()
	// log.Println(string(stdOut.Bytes()))
	// log.Println(string(stdErr.Bytes()))
	util.Check(err)

	if stdErr.Len() > 0 {
		panic(fmt.Sprintf("%s failed: %s", alg, stdErr.String()))
	}

	return stdOut.Bytes()
}

// Brotli11CLI returns a brotli compressed file as bytes
// at optimal compression (quality 11).
func Brotli11CLI(ctx context.Context, src string, out string) {
	runAlgorithm(ctx, "brotli", "--quality=11", "--output="+out, src)
}

// UnBrotliCLI returns a brotli compressed file as bytes
// at optimal compression (quality 11).
func UnBrotliCLI(ctx context.Context, filePath string) []byte {
	return runAlgorithm(ctx, "brotli", "--decompress", "--input", filePath)
}

// Gzip9Native returns a gzip compressed file as bytes
// at optimal compression (level 9).
func Gzip9Native(ctx context.Context, src string, out string) error {
	uncompressed, err := ioutil.ReadFile(src)
	util.Check(err)

	bytes := Gzip9Bytes(uncompressed)

	err = ioutil.WriteFile(out, bytes, 0644)
	util.Check(err)
	return nil
}

func Gzip9Bytes(uncompressed []byte) []byte {
	var b bytes.Buffer

	w, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	util.Check(err)

	_, err = w.Write(uncompressed)
	util.Check(err)
	util.Check(w.Close())

	return b.Bytes()
}

// UnGzip uncompresses a gzip file as bytes.
func UnGzip(compressed []byte) []byte {
	b := bytes.NewBuffer(compressed)

	r, err := gzip.NewReader(b)
	util.Check(err)

	var res bytes.Buffer
	_, err = res.ReadFrom(r)
	util.Check(err)

	return res.Bytes()
}
