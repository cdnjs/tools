package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/cdnjs/tools/util"

	"github.com/stretchr/testify/assert"
)

const (
	jsFilesPkg          = "jsPkg"
	oversizedFilesPkg   = "oversizePkg"
	unpublishedFieldPkg = "unpublishedPkg"
	sortByTimeStampPkg  = "sortByTimePkg"
	symlinkPkg          = "symlinkPkg"
	walkerPkg           = "walkerPkg"
	timeStamp1          = "1.0.0"
	timeStamp2          = "2.0.0"
	timeStamp3          = "3.0.0"
	timeStamp4          = "4.0.0"
	timeStamp5          = "5.0.0"
)

type ShowFilesTestCase struct {
	name         string
	input        string
	expected     string
	validatePath bool
	file         *string
}

func addTarFile(tw *tar.Writer, path string, content string) error {
	// now lets create the header as needed for this file within the tarball
	header := new(tar.Header)
	header.Name = path
	header.Size = int64(len(content))
	header.Mode = int64(0666)
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

func addTarSymlink(tw *tar.Writer, path string, dst string) error {
	// now lets create the header as needed for this file within the tarball
	header := new(tar.Header)
	header.Typeflag = tar.TypeSymlink
	header.Name = path
	header.Linkname = dst
	header.Mode = int64(0666)
	// write the header to the tarball archive
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	return nil
}

func createTar(filemap map[string]VirtualFile) (*os.File, error) {
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
	for path, virtualFile := range filemap {
		if virtualFile.LinkTo != "" {
			if err := addTarSymlink(tw, "package/"+path, virtualFile.LinkTo); err != nil {
				return nil, err
			}
		}
		if virtualFile.Content != "" {
			if err := addTarFile(tw, "package/"+path, virtualFile.Content); err != nil {
				return nil, err
			}
		}
	}

	return file, nil
}

func servePackage(w http.ResponseWriter, r *http.Request, filemap map[string]VirtualFile) {
	file, err := createTar(filemap)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	defer os.Remove(file.Name())

	http.ServeFile(w, r, file.Name())
}

type VirtualFile struct {
	Content string
	LinkTo  string
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
			"time": { "0.0.2": "2012-06-19T04:01:32.220Z" },
			"dist-tags": {
				"latest": "0.0.2"
			}
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
			"time": { "0.0.2": "2012-06-19T04:01:32.220Z" },
			"dist-tags": {
				"latest": "0.0.2"
			}
		}`)
	case "/" + unpublishedFieldPkg:
		fmt.Fprint(w, `{
			"versions": {
				"1.3.1": {
					"dist": {
						"tarball": "http://registry.npmjs.org/`+unpublishedFieldPkg+`.tgz"
					}
				}
			},
			 "time": {
    			"modified": "2014-12-30T19:39:27.425Z",
    			"created": "2014-12-30T19:39:27.425Z",
    			"1.3.1": "2014-12-30T19:39:27.425Z",
    			"unpublished": {
      				"name": "username",
      				"time": "2015-01-08T20:31:29.361Z",
      				"tags": {
        				"latest": "1.3.1"
					}
				}
			},
			"dist-tags": {
				"latest": "1.3.1"
			}
		}`)
	case "/" + sortByTimeStampPkg:
		fmt.Fprint(w, `{
			"versions": {
				"1.0.0": {
					"dist": {
						"tarball": "http://registry.npmjs.org/`+timeStamp1+`.tgz"
					}
				},
				"2.0.0": {
					"dist": {
						"tarball": "http://registry.npmjs.org/`+timeStamp2+`.tgz"
					}
				},
				"3.0.0": {
					"dist": {
						"tarball": "http://registry.npmjs.org/`+timeStamp3+`.tgz"
					}
				},
				"4.0.0": {
					"dist": {
						"tarball": "http://registry.npmjs.org/`+timeStamp4+`.tgz"
					}
				},
				"5.0.0": {
					"dist": {
						"tarball": "http://registry.npmjs.org/`+timeStamp5+`.tgz"
					}
				}
			},
			 "time": {
				"2.0.0": "2019-12-30T19:39:27.425Z",
				"3.0.0": "2019-12-30T19:38:27.425Z",
				"1.0.0": "2018-12-30T19:39:27.425Z",
				"5.0.0": "2017-12-30T19:39:27.425Z",
    			"4.0.0": "2017-11-30T19:39:27.425Z"
			},
			"dist-tags": {
				"latest": "3.0.0"
			}
		}`)
	case "/" + symlinkPkg:
		fmt.Fprint(w, `{
			"versions": {
				"0.0.2": {
					"dist": {
						"tarball": "http://registry.npmjs.org/`+symlinkPkg+`.tgz"
					}
				}
			},
			"time": { "0.0.2": "2012-06-19T04:01:32.220Z" },
			"dist-tags": {
				"latest": "0.0.2"
			}
		}`)
	case "/" + walkerPkg:
		fmt.Fprint(w, `{
			"versions": {
				"0.0.2": {
					"dist": {
						"tarball": "http://registry.npmjs.org/`+walkerPkg+`.tgz"
					}
				}
			},
			"time": { "0.0.2": "2012-06-19T04:01:32.220Z" },
			"dist-tags": {
				"latest": "0.0.2"
			}
		}`)
	case "/" + jsFilesPkg + ".tgz":
		servePackage(w, r, map[string]VirtualFile{
			"a.js": {Content: "a"},
			"b.js": {Content: "b"},
		})
	case "/" + oversizedFilesPkg + ".tgz":
		servePackage(w, r, map[string]VirtualFile{
			"a.js": {Content: strings.Repeat("a", int(util.MaxFileSize)+100)},
			"b.js": {Content: "ok"},
		})
	case "/" + unpublishedFieldPkg + ".tgz":
		servePackage(w, r, map[string]VirtualFile{
			"a.js": {Content: "a"},
			"b.js": {Content: "b"},
			"c.js": {Content: "c"},
		})
	case "/" + timeStamp2 + ".tgz":
		servePackage(w, r, map[string]VirtualFile{
			"2.js": {Content: "most recent version"},
		})
	case "/" + timeStamp3 + ".tgz":
		servePackage(w, r, map[string]VirtualFile{
			"3.js": {Content: "2nd most recent version"},
		})
	case "/" + timeStamp1 + ".tgz":
		servePackage(w, r, map[string]VirtualFile{
			"1.js": {Content: "3rd most recent version"},
		})
	case "/" + timeStamp5 + ".tgz":
		servePackage(w, r, map[string]VirtualFile{
			"5.js": {Content: "4th most recent version"},
		})
	case "/" + timeStamp4 + ".tgz":
		servePackage(w, r, map[string]VirtualFile{
			"4.js": {Content: "5th most recent version"},
		})
	case "/" + symlinkPkg + ".tgz":
		servePackage(w, r, map[string]VirtualFile{
			"a.js": {LinkTo: "/etc/issue"},
			"b.js": {LinkTo: "/dev/urandom"},
			"c.js": {Content: "/dev/urandom"},
		})
	case "/" + walkerPkg + ".tgz":
		servePackage(w, r, map[string]VirtualFile{
			"a.js":          {Content: "a"},
			"../../b.js":    {Content: "b"},
			"../../../c.js": {Content: "c"},
		})
	default:
		panic("unreachable: " + r.URL.Path)
	}
}

func TestCheckerNPMShowFiles(t *testing.T) {
	fakeBotPath := createFakeBotPath()
	defer os.RemoveAll(fakeBotPath)

	httpTestProxy := "localhost:8666"
	file := path.Join(fakeBotPath, "packages", "packages", "i", "input-show-files.json")

	cases := []ShowFilesTestCase{
		{
			name: "show files on npm",
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
		    "filename": "a.js",
		    "homepage": "https://github.com/tc80",
			"autoupdate": {
				"source": "npm",
				"target": "` + jsFilesPkg + `",
				"fileMap": [
					{ "basePath":"", "files":["*.js"] }
				]
			}
		}`,
			expected: `

most recent version: 0.0.2

` + "```" + `
a.js
b.js
` + "```" + `

0 last version(s):
`,
		},

		{
			name: "most recent version does not contain filename",
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
		    "filename": "not_included.js",
		    "homepage": "https://github.com/tc80",
			"autoupdate": {
				"source": "npm",
				"target": "` + jsFilesPkg + `",
				"fileMap": [
					{ "basePath":"", "files":["*.js"] }
				]
			}
		}`,
			expected: `

most recent version: 0.0.2

` + "```" + `
a.js
b.js
` + "```" + `
` + ciError(file, "Filename `not_included.js` not found in most recent version `0.0.2`.%0A") + `
0 last version(s):
`,
		},

		{
			name: "oversized file",
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
		    "filename": "b.js",
		    "homepage": "https://github.com/tc80",
			"autoupdate": {
				"source": "npm",
				"target": "` + oversizedFilesPkg + `",
				"fileMap": [
					{ "basePath":"", "files":["*.js"] }
				]
			}
		}`,
			expected: `

most recent version: 0.0.2
` + ciWarn(file, "file a.js ignored due to byte size (26214500 > 26214400)") + `
` + "```" + `
b.js
` + "```" + `

0 last version(s):
`,
		},

		{
			name: "unpublished field",
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
		    "filename": "a.js",
		    "homepage": "https://github.com/tc80",
			"autoupdate": {
				"source": "npm",
				"target": "` + unpublishedFieldPkg + `",
				"fileMap": [
					{ "basePath":"", "files":["*.js"] }
				]
			}
		}`,
			expected: `

most recent version: 1.3.1

` + "```" + `
a.js
b.js
c.js
` + "```" + `

0 last version(s):
`,
		},

		{
			name: "sort by time stamp",
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
		    "filename": "2.js",
		    "homepage": "https://github.com/tc80",
			"autoupdate": {
				"source": "npm",
				"target": "` + sortByTimeStampPkg + `",
				"fileMap": [
					{ "basePath":"", "files":["*.js"] }
				]
			}
		}`,
			expected: `

most recent version: 2.0.0

` + "```" + `
2.js
` + "```" + `

4 last version(s):
- 3.0.0: 1 file(s) matched :heavy_check_mark:
- 1.0.0: 1 file(s) matched :heavy_check_mark:
- 5.0.0: 1 file(s) matched :heavy_check_mark:
- 4.0.0: 1 file(s) matched :heavy_check_mark:
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
			pkgFile := file
			if tc.file != nil {
				pkgFile = *tc.file
			} else {
				err := ioutil.WriteFile(pkgFile, []byte(tc.input), 0644)
				assert.Nil(t, err)
			}

			// TODO: mock sandbox
			_ = tc.expected
			_ = tc.validatePath
			//
			// out := runChecker(fakeBotPath, httpTestProxy, tc.validatePath, "show-files", pkgFile)
			// assert.Equal(t, tc.expected, "\n"+out)

			os.Remove(pkgFile)
		})
	}

	assert.Nil(t, testproxy.Shutdown(context.Background()))
}

func TestCheckerShowFilesNPMSymlink(t *testing.T) {
	fakeBotPath := createFakeBotPath()
	defer os.RemoveAll(fakeBotPath)

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
			"source": "npm",
			"target": "` + symlinkPkg + `",
			"fileMap": [
				{ "basePath":"", "files":["*.js"] }
			]
		}
	}`
	expected := `most recent version: 0.0.2

` + "```" + `
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
	//
	// out := runChecker(fakeBotPath, httpTestProxy, false, "show-files", pkgFile)
	// assert.Contains(t, out, expected)
	assert.Nil(t, testproxy.Shutdown(context.Background()))
}

func TestCheckerShowFilesTarWalker(t *testing.T) {
	fakeBotPath := createFakeBotPath()
	defer os.RemoveAll(fakeBotPath)

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
			"source": "npm",
			"target": "` + walkerPkg + `",
			"fileMap": [
				{ "basePath":"", "files":["*.js"] }
			]
		}
	}`
	expected := []string{`
` + "```" + `
a.js
` + "```" + ``,
		"Unsafe file located outside", "with name: `package/../../b.js`",
		"Unsafe file located outside", "with name: `package/../../../c.js`",
	}

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
	// for _, text := range expected {
	// 	assert.Contains(t, out, text)
	// }
	assert.Nil(t, testproxy.Shutdown(context.Background()))
}
