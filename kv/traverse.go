package main

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

func getRoot(key string) Root {
	bytes, err := readKV(key)
	util.Check(err)
	var r Root
	util.Check(json.Unmarshal(bytes, &r))
	return r
}

func getPackage(key string) Package {
	bytes, err := readKV(key)
	util.Check(err)
	var p Package
	util.Check(json.Unmarshal(bytes, &p))
	return p
}

func getVersion(key string) Version {
	bytes, err := readKV(key)
	util.Check(err)
	var v Version
	util.Check(json.Unmarshal(bytes, &v))
	return v
}

func printFile(key string, f File) {
	// sri, name
	bytes, err := readKV(key)
	util.Check(err)
	fmt.Println("--------------------------")
	fmt.Printf("\nCurrent path: %s\n\n", key)
	fmt.Printf("Name: %s\n", f.Name)
	fmt.Printf("SRI: %s\n", f.SRI)
	fmt.Printf("Bytes: %d\n\n", len(bytes))
}

// lists options, returns selected option as well as
// if selected to go back a directory
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

// assume root exists
// enter '..' to go up a directory
// enter 'q' to quit
func traverse() {
	var p string
	var packagePath string
root:
	p = rootKey
	choice, _, back := listOptions(p, getRoot(p).Packages)
	if back {
		goto root
	}
	p = choice
	packagePath = p
versions:
	choice, _, back = listOptions(p, getPackage(p).Versions)
	if back {
		goto root
	}

	p = path.Join(p, choice)
	files := getVersion(p).Files
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
