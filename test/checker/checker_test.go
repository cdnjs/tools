package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

func createFakeBotPath() string {
	botpath, err := ioutil.TempDir("", "test-bot-path")
	if err != nil {
		panic(err)
	}

	os.MkdirAll(path.Join(botpath, "packages", "packages", "i"), os.ModePerm)

	// create fake glob that will return all the files all the time
	{
		dir := path.Join(botpath, "glob")
		if err := os.Mkdir(dir, os.ModePerm); err != nil {
			panic(err)
		}

		content := []byte(`#!/usr/bin/env node
			const fs = require('fs');
			fs.readdirSync(process.cwd()).forEach(file => {
				console.log(file);
			});
		`)
		err := ioutil.WriteFile(path.Join(dir, "index.js"), content, 0777)
		if err != nil {
			panic(err)
		}

	}

	return botpath
}

// start a local proxy server and run the checker binary
func runChecker(fakeBotPath string, proxy string, validatePath bool, args ...string) string {
	// used to avoid validating the package's path
	if !validatePath {
		args = append([]string{"-no-path-validation"}, args...)
	}

	cmd := exec.Command("../../bin/checker", args...)
	cmd.Env = append(os.Environ(),
		"HTTP_PROXY="+proxy,
		"BOT_BASE_PATH="+fakeBotPath,
	)

	out, _ := cmd.CombinedOutput()

	return string(out)
}

func ciError(file, err string) string {
	return fmt.Sprintf("::error file=%s,line=1,col=1::%s\n", file, err)
}

func ciWarn(file, err string) string {
	return fmt.Sprintf("::warning file=%s,line=1,col=1::%s\n", file, err)
}
