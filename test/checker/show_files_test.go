package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/cdnjs/tools/util"

	"github.com/stretchr/testify/assert"
)

const (
	jsFilesPkg        = "jsfilespackage"
	oversizedFilesPkg = "oversizedfilespackage"
)

type ShowFilesTestCase struct {
	name     string
	input    string
	expected string
}

func addTarFile(tw *tar.Writer, path string, content string) error {
	// now lets create the header as needed for this file within the tarball
	header := new(tar.Header)
	header.Name = path
	header.Size = int64(len(content))
	header.Mode = int64(0666)
	// header.ModTime = stat.ModTime()
	// write the header to the tarball archive
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	// copy the file data to the tarball
	if _, err := tw.Write([]byte(content)); err != nil {
		return err
	}
	return nil
}

func createTar(filemap map[string]string) (*os.File, error) {
	file, err := os.Create("/tmp/test.tgz")
	if err != nil {
		return nil, err
	}
	// set up the gzip writer
	gw := gzip.NewWriter(file)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()
	// add each file as needed into the current tar archive
	for path, content := range filemap {
		if err := addTarFile(tw, "package/"+path, content); err != nil {
			return nil, err
		}
	}

	return file, nil
}

func servePackage(w http.ResponseWriter, r *http.Request, filemap map[string]string) {
	file, err := createTar(filemap)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	http.ServeFile(w, r, file.Name())
	os.Remove(file.Name())
}

// fakes the npm api for testing purposes
func fakeNpmHandlerShowFiles(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/" + jsFilesPkg:
		fmt.Fprint(w, `{
			"versions": {
				"0.0.2": {
					"dist": {
						"tarball": "http://registry.npmjs.org/`+jsFilesPkg+`.tgz"
					}
				}
			},
			"time": { "0.0.2": "2012-06-19T04:01:32.220Z" }
		}`)
	case "/" + oversizedFilesPkg:
		fmt.Fprint(w, `{
			"versions": {
				"0.0.2": {
					"dist": {
						"tarball": "http://registry.npmjs.org/`+oversizedFilesPkg+`.tgz"
					}
				}
			},
			"time": { "0.0.2": "2012-06-19T04:01:32.220Z" }
		}`)
	case "/" + jsFilesPkg + ".tgz":
		servePackage(w, r, map[string]string{
			"a.js": "a",
			"b.js": "b",
		})
	case "/" + oversizedFilesPkg + ".tgz":
		servePackage(w, r, map[string]string{
			"a.js": strings.Repeat("a", int(util.MAX_FILE_SIZE)+100),
			"b.js": "ok",
		})
	default:
		panic("unreachable: " + r.URL.Path)
	}
}

func TestCheckerShowFiles(t *testing.T) {
	const (
		httpTestProxy = "localhost:8666"
		file          = "/tmp/input-show-files.json"
	)

	cases := []ShowFilesTestCase{
		{
			name: "show files on npm",
			input: `{
				"name": "foo",
				"repository": {
					"type": "git"
				},
				"autoupdate": {
					"source": "npm",
					"target": "` + jsFilesPkg + `",
					"fileMap": [
						{ "basePath":"", "files":["*.js"] }
					]
				}
			}`,
			expected: `

current version: 0.0.2

` + "```" + `
a.js
b.js
` + "```" + `

0 last version(s):
`,
		},

		{
			name: "oversized file",
			input: `{
				"name": "foo",
				"repository": {
					"type": "git"
				},
				"autoupdate": {
					"source": "npm",
					"target": "` + oversizedFilesPkg + `",
					"fileMap": [
						{ "basePath":"", "files":["*.js"] }
					]
				}
			}`,
			expected: `

current version: 0.0.2
` + ciWarn(file, "file a.js ignored due to byte size (10485860 > 10485760)") + `
` + "```" + `
b.js
` + "```" + `

0 last version(s):
`,
		},
	}

	testproxy := &http.Server{
		Addr:    httpTestProxy,
		Handler: http.Handler(http.HandlerFunc(fakeNpmHandlerShowFiles)),
	}

	go func() {
		if err := testproxy.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	for _, tc := range cases {
		tc := tc // capture range variable

		// since all tests share the same input, this needs to run sequentially
		t.Run(tc.name, func(t *testing.T) {
			err := ioutil.WriteFile(file, []byte(tc.input), 0644)
			assert.Nil(t, err)

			out := runChecker(httpTestProxy, "show-files", file)
			assert.Equal(t, tc.expected, "\n"+out)

			os.Remove(file)
		})
	}

	testproxy.Shutdown(context.Background())
}
