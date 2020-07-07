package kv

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/cdnjs/tools/util"
)

const (
	rootKey = "/packages"
)

var (
	reader = bufio.NewReader(os.Stdin)
)

// Root contains the list of all packages.
type Root struct {
	Packages []string `json:"packages"`
}

// Package contains the list of versions
// for a particular package.
//
// TODO:
// Add package-level metadata,
// which is currently stored on disk
// in package.json files (ex. latest version).
type Package struct {
	Versions []string `json:"versions"`
}

// Version contains the list of Files for a
// particular version.
//
// TODO:
// Determine what version-level metadata
// needs to be maintained (ex. time stamp).
type Version struct {
	Files []File `json:"files"`
}

// File represents a file and its
// calculated SRI when uncompressed.
type File struct {
	Name string `json:"name"`
	SRI  string `json:"sri"`
}

// GetRoot gets the root node in KV containing the list of packages.
func GetRoot(key string) (Root, error) {
	var r Root
	bytes, err := readKV(key)
	if err != nil {
		return r, err
	}
	err = json.Unmarshal(bytes, &r)
	return r, err
}

// GetPackage gets the package metadata from KV.
func GetPackage(key string) (Package, error) {
	var p Package
	bytes, err := readKV(key)
	if err != nil {
		return p, err
	}
	err = json.Unmarshal(bytes, &p)
	return p, err
}

// GetVersion gets the version metadata from KV.
func GetVersion(key string) (Version, error) {
	var v Version
	bytes, err := readKV(key)
	if err != nil {
		return v, err
	}
	err = json.Unmarshal(bytes, &v)
	return v, err
}

// Prints metadata for a file in KV, panicking if it does not exist.
func printFile(key string, f File) {
	bytes, err := readKV(key)
	util.Check(err)
	fmt.Println("--------------------------")
	fmt.Printf("\nCurrent path: %s\n\n", key)
	fmt.Printf("Name: %s\n", f.Name)
	fmt.Printf("SRI: %s\n", f.SRI)
	fmt.Printf("Bytes: %d\n\n", len(bytes))
}

// Lists options for a user to select.
// Returns the selected option, its index, and whether the
// user decided to return to that last set of options.
func listOptions(p string, opts []string) (string, int, bool) {
list:
	fmt.Println("--------------------------")
	fmt.Printf("\nCurrent path: %s\n\n", p)
	for i, s := range opts {
		fmt.Printf("%d - %s\n", i, s)
	}
	fmt.Print("\nEnter ID: ")
	text, err := reader.ReadString('\n')
	util.Check(err)
	text = strings.TrimSpace(text)
	if text == "q" || text == "quit" || text == "exit" {
		os.Exit(0)
	}
	if text == ".." {
		return "", 0, true
	}
	index, err := strconv.Atoi(text)
	if err != nil || index < 0 || index >= len(opts) {
		goto list
	}
	return opts[index], index, false
}

// Traverse is used for debugging the KV namespace.
// It is a basic CLI that interacts via stdin/stdout.
// It assumes the root entry `/packages` exists.
// To "go up" a directory, type `..`. To quit, type `q`.
func Traverse() {
	var p string
	var packagePath string
root:
	p = rootKey
	root, err := GetRoot(p)
	util.Check(err)
	choice, _, back := listOptions(p, root.Packages)
	if back {
		goto root
	}

	p = choice
	packagePath = p
versions:
	pkg, err := GetPackage(p)
	util.Check(err)
	choice, _, back = listOptions(p, pkg.Versions)
	if back {
		goto root
	}

	p = path.Join(p, choice)
	version, err := GetVersion(p)
	files := version.Files

	names := make([]string, len(files))
	for i, f := range files {
		names[i] = f.Name
	}
files:
	choice, i, back := listOptions(p, names)
	if back {
		p = packagePath
		goto versions
	}

	printFile(path.Join(p, choice), files[i])
	goto files
}
