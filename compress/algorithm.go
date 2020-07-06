package compress

import (
	"bytes"
	"fmt"
	"os/exec"

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

// Brotli11 returns a brotli compressed file as bytes
// at optimal compression (quality 11).
func Brotli11(filePath string) []byte {
	return runAlgorithm("brotli", "-c", "-q", "11", filePath)
}

// Gzip9 returns a gzip compressed file as bytes
// at optimal compression (level 9).
func Gzip9(filePath string) []byte {
	return runAlgorithm("gzip", "-c", "-9", filePath)
}
