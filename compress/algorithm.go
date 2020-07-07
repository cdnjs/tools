package compress

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"os/exec"

	"github.com/andybalholm/brotli"
	"github.com/cdnjs/tools/util"
)

// Runs an algorithm with a set of arguments,
// and returns its stdout as bytes.
// Note, this function will panic if anything is
// output to stderr.
func runAlgorithm(alg string, args ...string) []byte {
	cmd := exec.Command(alg, args...)
	var stdOut, stdErr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdOut, &stdErr

	err := cmd.Run()
	util.Check(err)

	if stdErr.Len() > 0 {
		panic(fmt.Sprintf("%s failed: %s", alg, stdErr.String()))
	}

	return stdOut.Bytes()
}

// Brotli11CLI returns a brotli compressed file as bytes
// at optimal compression (quality 11).
func Brotli11CLI(filePath string) []byte {
	return runAlgorithm("brotli", "-c", "-q", "11", filePath)
}

// Gzip9CLI returns a gzip compressed file as bytes
// at optimal compression (level 9).
func Gzip9CLI(filePath string) []byte {
	return runAlgorithm("gzip", "-c", "-9", filePath)
}

// Brotli11Native returns a brotli compressed file as bytes
// at optimal compression (quality 11).
func Brotli11Native(uncompressed []byte) []byte {
	var b bytes.Buffer

	w := brotli.NewWriterLevel(&b, brotli.BestCompression)

	_, err := w.Write(uncompressed)
	util.Check(err)
	util.Check(w.Close())

	return b.Bytes()
}

// Gzip9Native returns a gzip compressed file as bytes
// at optimal compression (level 9).
func Gzip9Native(uncompressed []byte) []byte {
	var b bytes.Buffer

	w, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	util.Check(err)

	_, err = w.Write(uncompressed)
	util.Check(err)
	util.Check(w.Close())

	return b.Bytes()
}
