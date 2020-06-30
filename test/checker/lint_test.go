package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type LintTestCase struct {
	name     string
	input    string
	expected string
}

const (
	unpopularPkg   = "unpopular"
	nonexistentPkg = "nonexistent"
)

// fakes the npm api for testing purposes
func fakeNpmHandlerLint(w http.ResponseWriter, r *http.Request) {
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

func TestCheckerLint(t *testing.T) {
	const httpTestProxy = "localhost:8667"
	const file = "/tmp/input-lint.json"

	cases := []LintTestCase{
		{
			name:     "error when invalid JSON",
			input:    `{ "package":, }`,
			expected: ciError(file, "failed to parse /tmp/input-lint.json: invalid character ',' looking for beginning of value"),
		},

		{
			name:  "show required fields",
			input: `{}`,
			expected: ciError(file, ".name should be specified") +
				ciError(file, ".autoupdate should not be null. Package will never auto-update") +
				ciError(file, "Unsupported .repository.type: "),
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
			expected: ciError(file, ".version should not exist"),
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
			expected: ciError(file, "Unsupported .autoupdate.source: ftp"),
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
			expected: ciError(file, "package doesn't exist on npm"),
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
			expected: ciError(file, "package download per month on npm is under 800"),
		},
	}

	testproxy := &http.Server{
		Addr:    httpTestProxy,
		Handler: http.Handler(http.HandlerFunc(fakeNpmHandlerLint)),
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

			out := runChecker(httpTestProxy, "lint", file)
			assert.Equal(t, tc.expected, out)

			os.Remove(file)
		})
	}

	assert.Nil(t, testproxy.Shutdown(context.Background()))
}
