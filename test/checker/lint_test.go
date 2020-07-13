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
	expected []string
}

const (
	unpopularPkg   = "unpopular"
	nonexistentPkg = "nonexistent"
	normalPkg      = "normal"
	unpopularRepo  = "user/unpopularRepo"
)

// fakes the npm api and GitHub api for testing purposes
func fakeNpmGitHubHandlerLint(w http.ResponseWriter, r *http.Request) {
	switch r.Host + r.URL.Path {
	case "registry.npmjs.org/" + nonexistentPkg:
		{
			w.WriteHeader(404)
			fmt.Fprint(w, `{"error":"Not found"}`)
		}
	case "registry.npmjs.org/" + unpopularPkg:
	case "registry.npmjs.org/" + normalPkg:
		{
			fmt.Fprint(w, `{}`)
		}
	case "api.npmjs.org/downloads/point/last-month/" + unpopularPkg:
		{
			fmt.Fprintf(w, `{"downloads":3,"start":"2020-05-28","end":"2020-06-26","package":"%s"}`, unpopularPkg)
		}
	case "api.npmjs.org/downloads/point/last-month/" + normalPkg:
		{
			fmt.Fprintf(w, `{"downloads":31789789,"start":"2020-05-28","end":"2020-06-26","package":"%s"}`, normalPkg)
		}
	case "api.github.com/repos/" + unpopularRepo:
		{
			fmt.Fprintf(w, `{"stargazers_count": 123}`)
		}
	default:
		panic(fmt.Sprintf("unknown path: %s", r.Host + r.URL.Path))
	}
}

func TestCheckerLint(t *testing.T) {
	const httpTestProxy = "localhost:8667"
	const file = "/tmp/input-lint.json"

	cases := []LintTestCase{
		{
			name:  "error when invalid JSON",
			input: `{ "package":, }`,
			expected: []string{
				ciError(file, "failed to parse /tmp/input-lint.json: invalid character ',' looking for beginning of value"),
			},
		},

		{
			name:  "show required fields",
			input: `{}`,
			expected: []string{
				ciError(file, ".name should be specified") +
					ciError(file, ".repository.url should be specified") +
					ciError(file, ".autoupdate should not be null. Package will never auto-update") +
					ciError(file, "Unsupported .repository.type: "),
			},
		},

		{
			name: "version should not be specified",
			input: `{
				"name": "foo",
				"version": "v123456",
				"repository": {
					"type": "git",
					"url": "git://ff"
				},
				"autoupdate": {
					"source": "git",
					"target": "git://ff"
				}
			}`,
			expected: []string{
				ciError(file, ".version should not exist"),
			},
		},

		{
			name: "unknown autoupdate source",
			input: `{
				"name": "foo",
				"repository": {
					"type": "git",
					"url": "lol"
				},
				"autoupdate": {
					"source": "ftp",
					"target": "lol"
				}
			}`,
			expected: []string{
				ciError(file, "Unsupported .autoupdate.source: ftp"),
			},
		},

		{
			name: "unknown package on npm",
			input: `{
				"name": "foo",
				"repository": {
					"type": "git",
					"url": "git://ff"
				},
				"autoupdate": {
					"source": "npm",
					"target": "` + nonexistentPkg + `"
				}
			}`,
			expected: []string{
				ciError(file, "package doesn't exist on npm"),
			},
		},

		{
			name: "check popularity (npm)",
			input: `{
				"name": "foo",
				"repository": {
					"type": "git",
					"url": "git://ff"
				},
				"autoupdate": {
					"source": "npm",
					"target": "` + unpopularPkg + `"
				}
			}`,
			expected: []string{
				ciWarn(file, "package download per month on npm is under 800"),
			},
		},

		{
			name: "check popularity (git)",
			input: `{
				"name": "foo",
				"repository": {
					"type": "git",
					"url": "https://github.com/` + unpopularRepo + `.git"
				},
				"autoupdate": {
					"source": "git",
					"target": "https://github.com/` + unpopularRepo + `.git"
				}
			}`,
			expected: []string{
				ciWarn(file, "stars on GitHub is under 200"),
			},
		},

		{
			name: "legacy NpmName and NpmFileMap should error",
			input: `{
				"name": "foo",
				"repository": {
					"type": "git",
					"url": "git://ff"
				},
				"npmName": "` + normalPkg + `",
				"npmFileMap": [
					{
						"basePath": "",
						"files": [
							"*.css"
						]
					}
				]

			}`,
			expected: []string{
				ciError(file, "unknown field npmName") +
					ciError(file, "unknown field npmFileMap") +
					ciError(file, ".autoupdate should not be null. Package will never auto-update"),
				ciError(file, "unknown field npmFileMap") +
					ciError(file, "unknown field npmName") +
					ciError(file, ".autoupdate should not be null. Package will never auto-update"),
			},
		},
	}

	testproxy := &http.Server{
		Addr:    httpTestProxy,
		Handler: http.Handler(http.HandlerFunc(fakeNpmGitHubHandlerLint)),
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
			assert.Contains(t, tc.expected, out)

			os.Remove(file)
		})
	}

	assert.Nil(t, testproxy.Shutdown(context.Background()))
}
