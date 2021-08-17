package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
)

func createGit(t *testing.T, filemap map[string]VirtualFile) string {
	dir, err := ioutil.TempDir("", "git")
	assert.Nil(t, err)

	repo, err := git.PlainInit(dir, false)
	assert.Nil(t, err)
	tree, err := repo.Worktree()
	assert.Nil(t, err)

	for name, vfile := range filemap {
		absfile := path.Join(dir, name)
		if vfile.Content != "" {
			err := ioutil.WriteFile(absfile, []byte(vfile.Content), 0644)
			assert.Nil(t, err)
			_, err = tree.Add(name)
			assert.Nil(t, err)
		}
		if vfile.LinkTo != "" {
			err := os.Symlink(vfile.LinkTo, absfile)
			assert.Nil(t, err)
		}
	}

	user := object.Signature{
		Name:  "Name",
		Email: "Email",
	}

	hash, err := tree.Commit("add files", &git.CommitOptions{
		Author: &user,
	})
	assert.Nil(t, err)
	_, err = repo.CreateTag("v0.0.1", hash, &git.CreateTagOptions{
		Tagger:  &user,
		Message: "v0.0.1",
	})
	assert.Nil(t, err)

	return dir
}

func TestCheckerShowFilesGitSymlink(t *testing.T) {
	fakeBotPath := createFakeBotPath()
	defer os.RemoveAll(fakeBotPath)

	symbolicGit := createGit(t, map[string]VirtualFile{
		"a.js": {LinkTo: "/etc/issue"},
		"b.js": {LinkTo: "/dev/urandom"},
		"c.js": {Content: "/dev/urandom"},
	})
	defer os.RemoveAll(symbolicGit)

	httpTestProxy := "localhost:8666"
	pkgFile := path.Join(fakeBotPath, "packages", "packages", "i", "input-show-files.json")
	input := `{
		"name": "a-happy-tyler",
		"description": "Tyler is happy. Be like Tyler.",
		"keywords": [
			"tyler",
			"happy"
		],
		"authors": [
			{
				"name": "Tyler Caslin",
				"email": "tylercaslin47@gmail.com",
				"url": "https://github.com/tc80"
			}
		],
		"license": "MIT",
		"repository": {
			"type": "git",
			"url": "git://github.com/tc80/a-happy-tyler.git"
		},
		"filename": "a.js",
		"homepage": "https://github.com/tc80",
		"autoupdate": {
			"source": "git",
			"target": "` + symbolicGit + `",
			"fileMap": [
				{ "basePath":"", "files":["*.js"] }
			]
		}
	}`
	expected := `most recent version: 0.0.1

` + "```" + `
.git
c.js
` + "```" + ``

	err := ioutil.WriteFile(pkgFile, []byte(input), 0644)
	assert.Nil(t, err)
	defer os.Remove(pkgFile)

	testproxy := &http.Server{
		Addr:    httpTestProxy,
		Handler: http.Handler(http.HandlerFunc(fakeNpmHandlerShowFiles)),
	}

	go func() {
		if err := testproxy.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	// TODO: mock sandbox
	_ = expected
	// out := runChecker(fakeBotPath, httpTestProxy, false, "show-files", pkgFile)
	// assert.Contains(t, out, expected)
	assert.Nil(t, testproxy.Shutdown(context.Background()))
}
