package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestCase struct {
	name     string
	input    string
	expected string
}

const NOT_POPULAR_PACKAGE = "notpopular"
const NOT_EXISTING_PACKAGE = "noexistingpackage"
const HTTP_TEST_PROXY = "localhost:8666"

// fakes the npm api for testing purposes
func fakeNpmHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/"+NOT_EXISTING_PACKAGE {
		w.WriteHeader(404)
		fmt.Fprint(w, `{"error":"Not found"}`)
		return
	}
	if r.URL.Path == "/"+NOT_POPULAR_PACKAGE {
		fmt.Fprint(w, `{}`)
		return
	}
	if r.URL.Path == "/downloads/point/last-month/"+NOT_POPULAR_PACKAGE {
		fmt.Fprintf(w, `{"downloads":3,"start":"2020-05-28","end":"2020-06-26","package":"%s"}`, NOT_POPULAR_PACKAGE)
		return
	}

	panic("unreachable")
}

// start a local proxy server and run the checker binary
func runChecker(args ...string) string {
	cmd := exec.Command("../bin/checker", args...)
	cmd.Env = append(os.Environ(),
		"HTTP_PROXY="+HTTP_TEST_PROXY,
	)

	out, _ := cmd.CombinedOutput()

	return string(out)
}

func CIError(err string) string {
	return fmt.Sprintf("::error file=/tmp/input.json,line=1,col=1::%s\n", err)
}

func TestCheckerLint(t *testing.T) {
	cases := []TestCase{
		{
			name:     "error when invalid JSON",
			input:    `{ "package":, }`,
			expected: CIError("failed to parse /tmp/input.json: invalid character ',' looking for beginning of value"),
		},

		{
			name:  "show required fields",
			input: `{}`,
			expected: CIError(".name should be specified") +
				CIError(".autoupdate should not be null. Package will never auto-update") +
				CIError("Unsupported .repository.type: "),
		},

		{
			name: "version should not be specified",
			input: `{
				"name": "foo",
				"version": "v123456",
				"repository": {
					"type": "git"
				},
				"autoupdate": {
					"source": "git",
					"target": "git://ff"
				}
			}`,
			expected: CIError(".version should be empty"),
		},

		{
			name: "unknown autoupdate source",
			input: `{
				"name": "foo",
				"repository": {
					"type": "git"
				},
				"autoupdate": {
					"source": "ftp",
					"target": "lol"
				}
			}`,
			expected: CIError("Unsupported .autoupdate.source: ftp"),
		},

		{
			name: "unknown package on npm",
			input: `{
				"name": "foo",
				"repository": {
					"type": "git"
				},
				"autoupdate": {
					"source": "npm",
					"target": "` + NOT_EXISTING_PACKAGE + `"
				}
			}`,
			expected: CIError("package doesn't exist on npm"),
		},

		{
			name: "check popularity",
			input: `{
				"name": "foo",
				"repository": {
					"type": "git"
				},
				"autoupdate": {
					"source": "npm",
					"target": "` + NOT_POPULAR_PACKAGE + `"
				}
			}`,
			expected: CIError("package download per month on npm is under 800"),
		},
	}

	testproxy := &http.Server{
		Addr:    HTTP_TEST_PROXY,
		Handler: http.Handler(http.HandlerFunc(fakeNpmHandler)),
	}

	go func() {
		if err := testproxy.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	for _, tc := range cases {
		tc := tc // capture range variable

		// since all tests share the same input, this needs to run sequentially
		t.Run(tc.name, func(t *testing.T) {
			tmpfile := "/tmp/input.json"

			err := ioutil.WriteFile(tmpfile, []byte(tc.input), 0644)
			assert.Nil(t, err)

			out := runChecker("lint", tmpfile)
			assert.Equal(t, tc.expected, out)

			os.Remove(tmpfile)
		})
	}

	testproxy.Shutdown(context.Background())
}
