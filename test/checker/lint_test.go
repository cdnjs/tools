package main

import (
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

const NOT_POPULAR_PACKAGE = "notpopular"
const NOT_EXISTING_PACKAGE = "noexistingpackage"

// fakes the npm api for testing purposes
func fakeNpmHandlerLint(w http.ResponseWriter, r *http.Request) {
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

func TestCheckerLint(t *testing.T) {
	const HTTP_TEST_PROXY = "localhost:8667"
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
			expected: ciError(file, ".version should be empty"),
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
					"target": "` + NOT_EXISTING_PACKAGE + `"
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
					"target": "` + NOT_POPULAR_PACKAGE + `"
				}
			}`,
			expected: ciError(file, "package download per month on npm is under 800"),
		},
	}

	testproxy := &http.Server{
		Addr:    HTTP_TEST_PROXY,
		Handler: http.Handler(http.HandlerFunc(fakeNpmHandlerLint)),
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
			err := ioutil.WriteFile(file, []byte(tc.input), 0644)
			assert.Nil(t, err)

			out := runChecker(HTTP_TEST_PROXY, "lint", file)
			assert.Equal(t, tc.expected, out)

			os.Remove(file)
		})
	}
}
