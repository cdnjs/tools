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
func runChecker(proxy string, args ...string) string {
	fakeBotPath := createFakeBotPath()

	cmd := exec.Command("../../bin/checker", args...)
	cmd.Env = append(os.Environ(),
		"HTTP_PROXY="+proxy,
		"BOT_BASE_PATH="+fakeBotPath,
	)

	out, _ := cmd.CombinedOutput()

	os.RemoveAll(fakeBotPath)

	return string(out)
}

func ciError(file, err string) string {
	return fmt.Sprintf("::error file=%s,line=1,col=1::%s\n", file, err)
}

func ciWarn(file, err string) string {
	return fmt.Sprintf("::warning file=%s,line=1,col=1::%s\n", file, err)
}
