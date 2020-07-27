package main

import (
	"testing"

	"github.com/cdnjs/tools/packages"
)

func TestHumanReadableSchema(t *testing.T) {
	cases := []SchemaTestCase{
		// (root) valid
		{
			filePath: "schema_tests/(root)/valid/all_properties.json",
			valid:    true,
			content: `{
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
		},
		{
			filePath: "schema_tests/(root)/valid/only_required_properties.json",
			valid:    true,
			content: `{
    "name": "a-happy-tyler",
    "description": "Tyler is happy. Be like Tyler.",
    "keywords": [
        "tyler"
    ],
    "filename": "happy.js",
    "homepage": "https://github.com/tc80",
    "repository": {
        "type": "git",
        "url": "git://github.com/tc80/a-happy-tyler.git"
    },
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
		},
		// (root) invalid
		{
			filePath: "schema_tests/(root)/invalid/additional_properties.json",
			valid:    false,
			errors: []string{
				"(root): Additional property licenses is not allowed",
				"(root): Additional property author is not allowed",
				"(root): Additional property my_custom_property is not allowed",
			},
			content: `{
    "name": "a-happy-tyler",
    "description": "Tyler is happy. Be like Tyler.",
    "keywords": [
        "tyler",
        "happy"
    ],
    "author": "Tyler Caslin",
    "authors": [
        {
            "name": "Tyler Caslin",
            "email": "tylercaslin47@gmail.com",
            "url": "https://github.com/tc80"
        }
    ],
    "license": "MIT",
    "licenses": [
        {
            "type": "My License 1"
        },
        {
            "type": "My License 2"
        }
    ],
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
    },
    "my_custom_property": "hello"
}`,
		},
		{
			filePath:    "schema_tests/(root)/invalid/invalid_json.txt",
			invalidJSON: true,
			content:     `{This is invalid JSON. Note that this is a .txt file to avoid VSCode from freaking out.}`,
		},
		// author valid
		{
			filePath: "schema_tests/authors/valid/missing_authors.json",
			valid:    true,
			content: `{
    "name": "a-happy-tyler",
    "description": "Tyler is happy. Be like Tyler.",
    "keywords": [
        "tyler",
        "happy"
    ],
    "license": "MIT",
    "repository": {
        "type": "git",
        "url": "git://github.com/tc80/a-happy-tyler.git"
    },
    "filename": "happy.js",
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
		},
		{
			filePath: "schema_tests/authors/valid/missing_email.json",
			valid:    true,
			content: `{
    "name": "a-happy-tyler",
    "description": "Tyler is happy. Be like Tyler.",
    "keywords": [
        "tyler",
        "happy"
    ],
    "authors": [
        {
            "name": "Tyler Caslin",
            "url": "https://github.com/tc80"
        }
    ],
    "license": "MIT",
    "repository": {
        "type": "git",
        "url": "git://github.com/tc80/a-happy-tyler.git"
    },
    "filename": "happy.js",
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
		},
		{
			filePath: "schema_tests/authors/valid/missing_url.json",
			valid:    true,
			content: `{
    "name": "a-happy-tyler",
    "description": "Tyler is happy. Be like Tyler.",
    "keywords": [
        "tyler",
        "happy"
    ],
    "authors": [
        {
            "name": "Tyler Caslin",
            "email": "tylercaslin47@gmail.com"
        }
    ],
    "license": "MIT",
    "repository": {
        "type": "git",
        "url": "git://github.com/tc80/a-happy-tyler.git"
    },
    "filename": "happy.js",
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
		},
		{
			filePath: "schema_tests/authors/valid/multiple_authors.json",
			valid:    true,
			content: `{
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
        },
        {
            "name": "Tyler Caslin 2",
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
		},
		{
			filePath: "schema_tests/authors/valid/one_author.json",
			valid:    true,
			content: `{
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
		},
		// author invalid
		{
			filePath: "schema_tests/authors/invalid/additional_property.json",
			valid:    false,
			errors:   []string{"authors.0: Additional property github is not allowed"},
			content: `{
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
            "url": "https://github.com/tc80",
            "github": "tc80"
        }
    ],
    "license": "MIT",
    "repository": {
        "type": "git",
        "url": "git://github.com/tc80/a-happy-tyler.git"
    },
    "filename": "happy.js",
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
		},
		{
			filePath: "schema_tests/authors/invalid/duplicate_authors.json",
			valid:    false,
			errors:   []string{"authors: array items[0,1] must be unique"},
			content: `{
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
        },
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
		},
		{
			filePath: "schema_tests/authors/invalid/empty_array.json",
			valid:    false,
			errors:   []string{"authors: Array must have at least 1 items"},
			content: `{
    "name": "a-happy-tyler",
    "description": "Tyler is happy. Be like Tyler.",
    "keywords": [
        "tyler",
        "happy"
    ],
    "authors": [],
    "license": "MIT",
    "repository": {
        "type": "git",
        "url": "git://github.com/tc80/a-happy-tyler.git"
    },
    "filename": "happy.js",
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
		},
		{
			filePath: "schema_tests/authors/invalid/empty_author_object.json",
			valid:    false,
			errors:   []string{"authors.0: name is required"},
			content: `{
    "name": "a-happy-tyler",
    "description": "Tyler is happy. Be like Tyler.",
    "keywords": [
        "tyler",
        "happy"
    ],
    "authors": [
        {}
    ],
    "license": "MIT",
    "repository": {
        "type": "git",
        "url": "git://github.com/tc80/a-happy-tyler.git"
    },
    "filename": "happy.js",
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
		},
		{
			filePath: "schema_tests/authors/invalid/empty_email.json",
			valid:    false,
			errors:   []string{"authors.0.email: String length must be greater than or equal to 1"},
			content: `{
    "name": "a-happy-tyler",
    "description": "Tyler is happy. Be like Tyler.",
    "keywords": [
        "tyler",
        "happy"
    ],
    "authors": [
        {
            "name": "Tyler Caslin",
            "email": "",
            "url": "https://github.com/tc80"
        }
    ],
    "license": "MIT",
    "repository": {
        "type": "git",
        "url": "git://github.com/tc80/a-happy-tyler.git"
    },
    "filename": "happy.js",
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
		},
		{
			filePath: "schema_tests/authors/invalid/empty_name.json",
			valid:    false,
			errors:   []string{"authors.0.name: String length must be greater than or equal to 1"},
			content: `{
    "name": "a-happy-tyler",
    "description": "Tyler is happy. Be like Tyler.",
    "keywords": [
        "tyler",
        "happy"
    ],
    "authors": [
        {
            "name": "",
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
		},
		{
			filePath: "schema_tests/authors/invalid/empty_url.json",
			valid:    false,
			errors:   []string{"authors.0.url: String length must be greater than or equal to 1"},
			content: `{
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
            "url": ""
        }
    ],
    "license": "MIT",
    "repository": {
        "type": "git",
        "url": "git://github.com/tc80/a-happy-tyler.git"
    },
    "filename": "happy.js",
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
		},
		{
			filePath: "schema_tests/authors/invalid/one_author_no_name.json",
			valid:    false,
			errors:   []string{"authors.0: name is required"},
			content: `{
    "name": "a-happy-tyler",
    "description": "Tyler is happy. Be like Tyler.",
    "keywords": [
        "tyler",
        "happy"
    ],
    "authors": [
        {
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
		},
		// autoupdate valid
		{
			filePath: "schema_tests/autoupdate/valid/empty_basepath.json",
			valid:    true,
			content: `{
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
    "autoupdate": {
        "source": "git",
        "target": "git://github.com/tc80/a-happy-tyler.git",
        "fileMap": [
            {
                "basePath": "",
                "files": [
                    "*"
                ]
            }
        ]
    }
}`,
		},
		{
			filePath: "schema_tests/autoupdate/valid/multiple_filemaps.json",
			valid:    true,
			content: `{
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
    "autoupdate": {
        "source": "git",
        "target": "git://github.com/tc80/a-happy-tyler.git",
        "fileMap": [
            {
                "basePath": "base1",
                "files": [
                    "*"
                ]
            },
            {
                "basePath": "base2",
                "files": [
                    "*"
                ]
            }
        ]
    }
}`,
		},
		{
			filePath: "schema_tests/autoupdate/valid/multiple_files.json",
			valid:    true,
			content: `{
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
    "autoupdate": {
        "source": "git",
        "target": "git://github.com/tc80/a-happy-tyler.git",
        "fileMap": [
            {
                "basePath": "src",
                "files": [
                    "*.js",
                    "*.css"
                ]
            }
        ]
    }
}`,
		},
		{
			filePath: "schema_tests/autoupdate/valid/source_git.json",
			valid:    true,
			content: `{
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
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		{
			filePath: "schema_tests/autoupdate/valid/source_npm.json",
			valid:    true,
			content: `{
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
    "autoupdate": {
        "source": "npm",
        "target": "a-happy-tyler",
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
		},
		// autoupdate invalid
		{
			filePath: "schema_tests/autoupdate/invalid/additional_properties.json",
			valid:    false,
			errors: []string{
				"autoupdate: Additional property repo is not allowed",
				"autoupdate.fileMap.0: Additional property directory is not allowed",
			},
			content: `{
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
    "autoupdate": {
        "repo": "github",
        "source": "git",
        "target": "a-happy-tyler",
        "fileMap": [
            {
                "directory": "happy",
                "basePath": "src",
                "files": [
                    "*"
                ]
            }
        ]
    }
}`,
		},
		{
			filePath: "schema_tests/autoupdate/invalid/duplicate_filemap.json",
			valid:    false,
			errors:   []string{"autoupdate.fileMap: array items[0,1] must be unique"},
			content: `{
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
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
        "fileMap": [
            {
                "basePath": "src",
                "files": [
                    "*"
                ]
            },
            {
                "basePath": "src",
                "files": [
                    "*"
                ]
            }
        ]
    }
}`,
		},
		{
			filePath: "schema_tests/autoupdate/invalid/duplicate_files.json",
			valid:    false,
			errors:   []string{"autoupdate.fileMap.0.files: array items[0,1] must be unique"},
			content: `{
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
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
        "fileMap": [
            {
                "basePath": "src",
                "files": [
                    "*",
                    "*"
                ]
            }
        ]
    }
}`,
		},
		{
			filePath: "schema_tests/autoupdate/invalid/empty_file.json",
			valid:    false,
			errors:   []string{"autoupdate.fileMap.0.files.0: String length must be greater than or equal to 1"},
			content: `{
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
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
        "fileMap": [
            {
                "basePath": "src",
                "files": [
                    ""
                ]
            }
        ]
    }
}`,
		},
		{
			filePath: "schema_tests/autoupdate/invalid/empty_filemap.json",
			valid:    false,
			errors:   []string{"autoupdate.fileMap: Array must have at least 1 items"},
			content: `{
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
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
        "fileMap": []
    }
}`,
		},
		{
			filePath: "schema_tests/autoupdate/invalid/empty_files.json",
			valid:    false,
			errors:   []string{"autoupdate.fileMap.0.files: Array must have at least 1 items"},
			content: `{
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
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
        "fileMap": [
            {
                "basePath": "src",
                "files": []
            }
        ]
    }
}`,
		},
		{
			filePath: "schema_tests/autoupdate/invalid/empty_source.json",
			valid:    false,
			errors:   []string{"autoupdate.source: Does not match pattern '" + autoupdateSourceRegex + "'"},
			content: `{
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
    "autoupdate": {
        "source": "",
        "target": "a-happy-tyler",
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
		},
		{
			filePath: "schema_tests/autoupdate/invalid/empty_target.json",
			valid:    false,
			errors:   []string{"autoupdate.target: String length must be greater than or equal to 1"},
			content: `{
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
    "autoupdate": {
        "source": "git",
        "target": "",
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
		},
		{
			filePath: "schema_tests/autoupdate/invalid/missing_autoupdate.json",
			valid:    false,
			errors:   []string{"(root): autoupdate is required"},
			content: `{
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
    "filename": "happy.js"
}`,
		},
		{
			filePath: "schema_tests/autoupdate/invalid/missing_basepath.json",
			valid:    false,
			errors:   []string{"autoupdate.fileMap.0: basePath is required"},
			content: `{
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
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
        "fileMap": [
            {
                "files": [
                    "*"
                ]
            }
        ]
    }
}`,
		},
		{
			filePath: "schema_tests/autoupdate/invalid/missing_filemap.json",
			valid:    false,
			errors:   []string{"autoupdate: fileMap is required"},
			content: `{
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
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler"
    }
}`,
		},
		{
			filePath: "schema_tests/autoupdate/invalid/missing_files.json",
			valid:    false,
			errors:   []string{"autoupdate.fileMap.0: files is required"},
			content: `{
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
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
        "fileMap": [
            {
                "basePath": "src"
            }
        ]
    }
}`,
		},
		{
			filePath: "schema_tests/autoupdate/invalid/missing_source.json",
			valid:    false,
			errors:   []string{"autoupdate: source is required"},
			content: `{
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
    "autoupdate": {
        "target": "a-happy-tyler",
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
		},
		{
			filePath: "schema_tests/autoupdate/invalid/missing_target.json",
			valid:    false,
			errors:   []string{"autoupdate: target is required"},
			content: `{
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
    "autoupdate": {
        "source": "git",
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
		},
		{
			filePath: "schema_tests/autoupdate/invalid/source_svn.json",
			valid:    false,
			errors:   []string{"autoupdate.source: Does not match pattern '" + autoupdateSourceRegex + "'"},
			content: `{
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
    "autoupdate": {
        "source": "svn",
        "target": "a-happy-tyler",
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
		},
		// description valid
		{
			filePath: "schema_tests/description/valid/valid_description.json",
			valid:    true,
			content: `{
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
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		// description invalid
		{
			filePath: "schema_tests/description/invalid/empty_description.json",
			errors:   []string{"description: String length must be greater than or equal to 1"},
			content: `{
    "name": "a-happy-tyler",
    "description": "",
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
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		{
			filePath: "schema_tests/description/invalid/missing_description.json",
			errors:   []string{"(root): description is required"},
			content: `{
    "name": "a-happy-tyler",
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
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		// filename valid
		{
			filePath: "schema_tests/filename/valid/valid_filename.json",
			valid:    true,
			content: `{
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
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		// filename invalid
		{
			filePath: "schema_tests/filename/invalid/empty_filename.json",
			errors:   []string{"filename: String length must be greater than or equal to 1"},
			content: `{
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
    "filename": "",
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		{
			filePath: "schema_tests/filename/invalid/missing_filename.json",
			errors:   []string{"(root): filename is required"},
			content: `{
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
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		// homepage valid
		{
			filePath: "schema_tests/homepage/valid/missing_homepage.json",
			valid:    true,
			content: `{
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
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		{
			filePath: "schema_tests/homepage/valid/valid_homepage.json",
			valid:    true,
			content: `{
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
    "homepage": "https://github.com/tc80",
    "license": "MIT",
    "repository": {
        "type": "git",
        "url": "git://github.com/tc80/a-happy-tyler.git"
    },
    "filename": "happy.js",
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		// homepage invalid
		{
			filePath: "schema_tests/homepage/invalid/empty_homepage.json",
			errors:   []string{"homepage: String length must be greater than or equal to 1"},
			content: `{
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
    "homepage": "",
    "license": "MIT",
    "repository": {
        "type": "git",
        "url": "git://github.com/tc80/a-happy-tyler.git"
    },
    "filename": "happy.js",
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		// keywords valid
		{
			filePath: "schema_tests/keywords/valid/multiple_keywords.json",
			valid:    true,
			content: `{
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
		},
		{
			filePath: "schema_tests/keywords/valid/one_keyword.json",
			valid:    true,
			content: `{
    "name": "a-happy-tyler",
    "description": "Tyler is happy. Be like Tyler.",
    "keywords": [
        "tyler"
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
		},
		// keywords invalid
		{
			filePath: "schema_tests/keywords/invalid/duplicate_keywords.json",
			errors:   []string{"keywords: array items[0,1] must be unique"},
			content: `{
    "name": "a-happy-tyler",
    "description": "Tyler is happy. Be like Tyler.",
    "keywords": [
        "tyler",
        "tyler"
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
		},
		{
			filePath: "schema_tests/keywords/invalid/empty_keyword.json",
			errors:   []string{"keywords.0: String length must be greater than or equal to 1"},
			content: `{
    "name": "a-happy-tyler",
    "description": "Tyler is happy. Be like Tyler.",
    "keywords": [
        "",
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
		},
		{
			filePath: "schema_tests/keywords/invalid/empty_keywords.json",
			errors:   []string{"keywords: Array must have at least 1 items"},
			content: `{
    "name": "a-happy-tyler",
    "description": "Tyler is happy. Be like Tyler.",
    "keywords": [],
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
		},
		{
			filePath: "schema_tests/keywords/invalid/missing_keywords.json",
			errors:   []string{"(root): keywords is required"},
			content: `{
    "name": "a-happy-tyler",
    "description": "Tyler is happy. Be like Tyler.",
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
		},
		// license valid
		{
			filePath: "schema_tests/license/valid/many_licenses.json",
			valid:    true,
			content: `{
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
    "homepage": "https://github.com/tc80",
    "license": "(MIT1 OR MIT2 OR MIT3 OR MIT4 OR MIT5)",
    "repository": {
        "type": "git",
        "url": "git://github.com/tc80/a-happy-tyler.git"
    },
    "filename": "happy.js",
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		{
			filePath: "schema_tests/license/valid/missing_license.json",
			valid:    true,
			content: `{
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
    "homepage": "https://github.com/tc80",
    "repository": {
        "type": "git",
        "url": "git://github.com/tc80/a-happy-tyler.git"
    },
    "filename": "happy.js",
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		{
			filePath: "schema_tests/license/valid/single_license.json",
			valid:    true,
			content: `{
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
    "homepage": "https://github.com/tc80",
    "license": "MIT",
    "repository": {
        "type": "git",
        "url": "git://github.com/tc80/a-happy-tyler.git"
    },
    "filename": "happy.js",
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		{
			filePath: "schema_tests/license/valid/two_licenses.json",
			valid:    true,
			content: `{
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
    "homepage": "https://github.com/tc80",
    "license": "(MIT1 OR MIT2)",
    "repository": {
        "type": "git",
        "url": "git://github.com/tc80/a-happy-tyler.git"
    },
    "filename": "happy.js",
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		// license invalid
		{
			filePath: "schema_tests/license/invalid/empty_license.json",
			errors:   []string{"license: Does not match pattern '" + licenseRegex + "'"},
			content: `{
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
    "homepage": "https://github.com/tc80",
    "license": "",
    "repository": {
        "type": "git",
        "url": "git://github.com/tc80/a-happy-tyler.git"
    },
    "filename": "happy.js",
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		{
			filePath: "schema_tests/license/invalid/invalid_multiple_licenses.json",
			errors:   []string{"license: Does not match pattern '" + licenseRegex + "'"},
			content: `{
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
    "homepage": "https://github.com/tc80",
    "license": "(MIT OR )",
    "repository": {
        "type": "git",
        "url": "git://github.com/tc80/a-happy-tyler.git"
    },
    "filename": "happy.js",
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		{
			filePath: "schema_tests/license/invalid/invalid_single_license.json",
			errors:   []string{"license: Does not match pattern '" + licenseRegex + "'"},
			content: `{
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
    "homepage": "https://github.com/tc80",
    "license": "(MIT",
    "repository": {
        "type": "git",
        "url": "git://github.com/tc80/a-happy-tyler.git"
    },
    "filename": "happy.js",
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		// name valid
		{
			filePath: "schema_tests/name/valid/valid_name.json",
			valid:    true,
			content: `{
    "name": "a_happy-Tyler123",
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
    "homepage": "https://github.com/tc80",
    "license": "MIT",
    "repository": {
        "type": "git",
        "url": "git://github.com/tc80/a-happy-tyler.git"
    },
    "filename": "happy.js",
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		// name invalid
		{
			filePath: "schema_tests/name/invalid/empty_name.json",
			errors:   []string{"name: Does not match pattern '" + nameRegex + "'"},
			content: `{
    "name": "",
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
    "homepage": "https://github.com/tc80",
    "license": "MIT",
    "repository": {
        "type": "git",
        "url": "git://github.com/tc80/a-happy-tyler.git"
    },
    "filename": "happy.js",
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		{
			filePath: "schema_tests/name/invalid/invalid_name.json",
			errors:   []string{"name: Does not match pattern '" + nameRegex + "'"},
			content: `{
    "name": "an/invalid/name",
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
    "homepage": "https://github.com/tc80",
    "license": "MIT",
    "repository": {
        "type": "git",
        "url": "git://github.com/tc80/a-happy-tyler.git"
    },
    "filename": "happy.js",
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		{
			filePath: "schema_tests/name/invalid/missing_name.json",
			errors:   []string{"(root): name is required"},
			content: `{
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
    "homepage": "https://github.com/tc80",
    "license": "MIT",
    "repository": {
        "type": "git",
        "url": "git://github.com/tc80/a-happy-tyler.git"
    },
    "filename": "happy.js",
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		// repository valid
		{
			filePath: "schema_tests/repository/valid/type_git.json",
			valid:    true,
			content: `{
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
    "homepage": "https://github.com/tc80",
    "license": "MIT",
    "repository": {
        "type": "git",
        "url": "git://github.com/tc80/a-happy-tyler.git"
    },
    "filename": "happy.js",
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		// repository invalid
		{
			filePath: "schema_tests/repository/invalid/additional_property.json",
			errors:   []string{"repository: Additional property custom_type is not allowed"},
			content: `{
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
    "homepage": "https://github.com/tc80",
    "license": "MIT",
    "repository": {
        "type": "git",
        "url": "git://github.com/tc80/a-happy-tyler.git",
        "custom_type": "custom"
    },
    "filename": "happy.js",
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		{
			filePath: "schema_tests/repository/invalid/empty_repository.json",
			errors: []string{
				"repository: type is required",
				"repository: url is required",
			},
			content: `{
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
    "homepage": "https://github.com/tc80",
    "license": "MIT",
    "repository": {},
    "filename": "happy.js",
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		{
			filePath: "schema_tests/repository/invalid/empty_type.json",
			errors:   []string{"repository.type: Does not match pattern '" + repositoryTypeRegex + "'"},
			content: `{
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
    "homepage": "https://github.com/tc80",
    "license": "MIT",
    "repository": {
        "type": "",
        "url": "git://github.com/tc80/a-happy-tyler.git"
    },
    "filename": "happy.js",
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		{
			filePath: "schema_tests/repository/invalid/empty_url.json",
			errors:   []string{"repository.url: Does not match pattern '" + repositoryURLRegex + "'"},
			content: `{
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
    "homepage": "https://github.com/tc80",
    "license": "MIT",
    "repository": {
        "type": "git",
        "url": ""
    },
    "filename": "happy.js",
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		{
			filePath: "schema_tests/repository/invalid/missing_repository.json",
			errors:   []string{"(root): repository is required"},
			content: `{
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
    "homepage": "https://github.com/tc80",
    "license": "MIT",
    "filename": "happy.js",
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		{
			filePath: "schema_tests/repository/invalid/missing_type.json",
			errors:   []string{"repository: type is required"},
			content: `{
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
    "homepage": "https://github.com/tc80",
    "license": "MIT",
    "repository": {
        "url": "git://github.com/tc80/a-happy-tyler.git"
    },
    "filename": "happy.js",
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		{
			filePath: "schema_tests/repository/invalid/missing_url.json",
			errors:   []string{"repository: url is required"},
			content: `{
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
    "homepage": "https://github.com/tc80",
    "license": "MIT",
    "repository": {
        "type": "git"
    },
    "filename": "happy.js",
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		{
			filePath: "schema_tests/repository/invalid/type_git_invalid_url.json",
			errors:   []string{"repository.url: Does not match pattern '" + repositoryURLRegex + "'"},
			content: `{
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
    "homepage": "https://github.com/tc80",
    "license": "MIT",
    "repository": {
        "type": "git",
        "url": "git://git.com/tc80/a-happy-tyler.git"
    },
    "filename": "happy.js",
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		{
			filePath: "schema_tests/repository/invalid/type_npm.json",
			errors: []string{
				"repository.type: Does not match pattern '" + repositoryTypeRegex + "'",
				"repository.url: Does not match pattern '" + repositoryURLRegex + "'",
			},
			content: `{
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
    "homepage": "https://github.com/tc80",
    "license": "MIT",
    "repository": {
        "type": "npm",
        "url": "https://www.npmjs.com/package/a-happy-tyler"
    },
    "filename": "happy.js",
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
		{
			filePath: "schema_tests/repository/invalid/type_svn.json",
			errors: []string{
				"repository.type: Does not match pattern '" + repositoryTypeRegex + "'",
				"repository.url: Does not match pattern '" + repositoryURLRegex + "'",
			},
			content: `{
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
    "homepage": "https://github.com/tc80",
    "license": "MIT",
    "repository": {
        "type": "svn",
        "url": "https://www.svn-not-a-valid-link.com/package/a-happy-tyler"
    },
    "filename": "happy.js",
    "autoupdate": {
        "source": "git",
        "target": "a-happy-tyler",
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
		},
	}

	runSchemaTestCases(t, packages.HumanReadableSchema, cases)
}
