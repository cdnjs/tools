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

var (
	reader = bufio.NewReader(os.Stdin)
)

// Gets the root node in KV containing the list of packages.
func getRoot(key string) (Root, error) {
	var r Root
	bytes, err := readKV(key)
	if err != nil {
		return r, err
	}
	util.Check(json.Unmarshal(bytes, &r))
	return r, nil
}

// Gets package metadata from KV.
func getPackage(key string) (Package, error) {
	var p Package
	bytes, err := readKV(key)
	if err != nil {
		return p, err
	}
	util.Check(json.Unmarshal(bytes, &p))
	return p, nil
}

// Gets version metadata from KV.
func getVersion(key string) (Version, error) {
	var v Version
	bytes, err := readKV(key)
	if err != nil {
		return v, err
	}
	util.Check(json.Unmarshal(bytes, &v))
	return v, nil
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
// It is a basic CLI that prints to stdout.
// It assumes the root entry `/packages` exists.
// To "go up" a directory, type `..`. To quit, type `q`.
func Traverse() {
	var p string
	var packagePath string
root:
	p = rootKey
	root, err := getRoot(p)
	util.Check(err)
	choice, _, back := listOptions(p, root.Packages)
	if back {
		goto root
	}

	p = choice
	packagePath = p
versions:
	pkg, err := getPackage(p)
	util.Check(err)
	choice, _, back = listOptions(p, pkg.Versions)
	if back {
		goto root
	}

	p = path.Join(p, choice)
	version, err := getVersion(p)
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
