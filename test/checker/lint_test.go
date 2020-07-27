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
	popularRepo    = "user/popularRepo"
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
	case "api.github.com/repos/" + popularRepo:
		{
			fmt.Fprintf(w, `{"stargazers_count": 500}`)
		}
	default:
		panic(fmt.Sprintf("unknown path: %s", r.Host+r.URL.Path))
	}
}

func TestCheckerLint(t *testing.T) {
	const httpTestProxy = "localhost:8667"
	const file = "/tmp/input-lint.json"

	cases := []LintTestCase{
		{
			name:     "error when invalid JSON",
			input:    `{ "package":, }`,
			expected: []string{ciError(file, "failed to parse /tmp/input-lint.json: invalid character ',' looking for beginning of value")},
		},

		{
			name:  "show required fields",
			input: `{}`,
			expected: []string{
				ciError(file, "(root): autoupdate is required") +
					ciError(file, "(root): description is required") +
					ciError(file, "(root): filename is required") +
					ciError(file, "(root): keywords is required") +
					ciError(file, "(root): name is required") +
					ciError(file, "(root): repository is required"),
			},
		},

		{
			name: "version should not be specified",
			input: `{
			"version": "v123456",
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
		    "filename": "happy.js",
		    "homepage": "https://github.com/tc80",
		    "autoupdate": {
		        "source": "git",
		        "target": "git://github.com/tc80/a-happy-tyler.git",
		        "fileMap": [
		            {
		                "basePath": "src",
		                "files": [
		                    "*"
		                ]
		            }
		        ]
		    }
		}`,
			expected: []string{ciError(file, "(root): Additional property version is not allowed")},
		},

		{
			name: "unknown autoupdate source",
			input: `{
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
		    "filename": "happy.js",
		    "homepage": "https://github.com/tc80",
		    "autoupdate": {
		        "source": "ftp",
		        "target": "git://github.com/tc80/a-happy-tyler.git",
		        "fileMap": [
		            {
		                "basePath": "src",
		                "files": [
		                    "*"
		                ]
		            }
		        ]
		    }
		}`,
			expected: []string{ciError(file, "autoupdate.source: Does not match pattern '"+autoupdateSourceRegex+"'")},
		},

		{
			name: "unknown package on npm",
			input: `{
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
		    "filename": "happy.js",
		    "homepage": "https://github.com/tc80",
		    "autoupdate": {
		        "source": "npm",
		        "target": "` + nonexistentPkg + `",
		        "fileMap": [
		            {
		                "basePath": "src",
		                "files": [
		                    "*"
		                ]
		            }
		        ]
		    }
		}`,
			expected: []string{ciError(file, "package doesn't exist on npm")},
		},

		{
			name: "check popularity (npm)",
			input: `{
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
		        "url": "git://github.com/` + unpopularRepo + `"
		    },
		    "filename": "happy.js",
		    "homepage": "https://github.com/tc80",
		    "autoupdate": {
		        "source": "npm",
		        "target": "` + unpopularPkg + `",
		        "fileMap": [
		            {
		                "basePath": "src",
		                "files": [
		                    "*"
		                ]
		            }
		        ]
		    }
		}`,
			expected: []string{
				ciWarn(file, "stars on GitHub is under 200") +
					ciWarn(file, "package download per month on npm is under 800"),
			},
		},

		{
			name: "check popularity (git)",
			input: `{
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
		        "url": "https://github.com/` + unpopularRepo + `.git"
		    },
		    "filename": "happy.js",
		    "homepage": "https://github.com/tc80",
		    "autoupdate": {
		        "source": "git",
		        "target": "https://github.com/` + unpopularRepo + `.git",
		        "fileMap": [
		            {
		                "basePath": "src",
		                "files": [
		                    "*"
		                ]
		            }
		        ]
		    }
		}`,
			expected: []string{ciWarn(file, "stars on GitHub is under 200")},
		},

		{
			name: "legacy NpmName and NpmFileMap should error",
			input: `{
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
		        "url": "https://github.com/` + unpopularRepo + `.git"
		    },
		    "filename": "happy.js",
		    "homepage": "https://github.com/tc80",
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
				ciError(file, "(root): autoupdate is required") +
					ciError(file, "(root): Additional property npmName is not allowed") +
					ciError(file, "(root): Additional property npmFileMap is not allowed"),
				ciError(file, "(root): autoupdate is required") +
					ciError(file, "(root): Additional property npmFileMap is not allowed") +
					ciError(file, "(root): Additional property npmName is not allowed"),
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
