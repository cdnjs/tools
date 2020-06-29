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

const (
	unpopularPkg   = "unpopular"
	nonexistentPkg = "nonexistent"
	httpTestProxy  = "localhost:8666"
	tmpFile        = "/tmp/input.json"
)

// fakes the npm api for testing purposes
func fakeNpmHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/" + nonexistentPkg:
		{
			w.WriteHeader(404)
			fmt.Fprint(w, `{"error":"Not found"}`)
		}
	case "/" + unpopularPkg:
		{
			fmt.Fprint(w, `{}`)
		}
	case "/downloads/point/last-month/" + unpopularPkg:
		{
			fmt.Fprintf(w, `{"downloads":3,"start":"2020-05-28","end":"2020-06-26","package":"%s"}`, unpopularPkg)
		}
	default:
		panic(fmt.Sprintf("unknown path: %s", r.URL.Path))
	}
}

// start a local proxy server and run the checker binary
func runChecker(args ...string) string {
	cmd := exec.Command("../bin/checker", args...)
	cmd.Env = append(os.Environ(),
		"HTTP_PROXY="+httpTestProxy,
	)

	out, _ := cmd.CombinedOutput()

	return string(out)
}

func CIError(err string) string {
	return fmt.Sprintf("::error file=%s,line=1,col=1::%s\n", tmpFile, err)
}

func TestCheckerLint(t *testing.T) {
	if os.Getenv("DEBUG") != "" {
		panic("DEBUG mode must be unset")
	}
	cases := []TestCase{
		{
			name:     "error when invalid JSON",
			input:    `{ "package":, }`,
			expected: CIError(fmt.Sprintf("failed to parse %s: invalid character ',' looking for beginning of value", tmpFile)),
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
					"target": "` + nonexistentPkg + `"
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
					"target": "` + unpopularPkg + `"
				}
			}`,
			expected: CIError("package download per month on npm is under 800"),
		},
	}

	testproxy := &http.Server{
		Addr:    httpTestProxy,
		Handler: http.Handler(http.HandlerFunc(fakeNpmHandler)),
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

			err := ioutil.WriteFile(tmpFile, []byte(tc.input), 0644)
			assert.Nil(t, err)

			out := runChecker("lint", tmpFile)
			assert.Equal(t, tc.expected, out)

			os.Remove(tmpFile)
		})
	}

	testproxy.Shutdown(context.Background())
}
