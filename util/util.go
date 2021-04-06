package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	cdnjsLibsPath     = GetCDNJSLibrariesPath()
	humanPackagesPath = GetHumanPackagesPath()
	srisPath          = GetSRIsPath()
)

// Assert is used to enforce a condition is true.
func Assert(cond bool) {
	if !cond {
		panic("assertion failure")
	}
}

// MoveFile moves a file from a source path to destination path.
func MoveFile(sourcePath, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("Couldn't open source file: %s", err)
	}
	outputFile, err := os.Create(destPath)
	if err != nil {
		inputFile.Close()
		return fmt.Errorf("Couldn't open dest file: %s", err)
	}
	defer outputFile.Close()
	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("Writing to output file failed: %s", err)
	}
	// The copy was successful, so now delete the original file
	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("Failed removing original file: %s", err)
	}
	return nil
}

// ReadSRISafely reads a cdnjs/sris file safely.
func ReadSRISafely(file string) ([]byte, error) {
	return ReadFileSafely(file, srisPath)
}

// ReadHumanPackageSafely reads a cdnjs/packages file safely.
func ReadHumanPackageSafely(file string) ([]byte, error) {
	return ReadFileSafely(file, humanPackagesPath)
}

// ReadLibFileSafely reads a cdnjs/cdnjs file safely.
func ReadLibFileSafely(file string) ([]byte, error) {
	return ReadFileSafely(file, cdnjsLibsPath)
}

// ReadFileSafely reads a cdnjs file from disk safely, checking that
// it is located under the correct directory.
func ReadFileSafely(file, underPath string) ([]byte, error) {
	// evaluate any symlinks, finding the full path to the target file
	target, err := filepath.EvalSymlinks(file)
	if err != nil {
		return nil, err
	}
	// check that the target file is located under a particular directory
	if !strings.HasPrefix(target, underPath) {
		return nil, fmt.Errorf("Unsafe file located outside `%s` with path: `%s`", cdnjsLibsPath, target)
	}
	return ioutil.ReadFile(target)
}
