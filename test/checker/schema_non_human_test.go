package main

import (
	"testing"

	"github.com/cdnjs/tools/packages"
)

func TestNonHumanReadableSchema(t *testing.T) {
	cases := []SchemaTestCase{
		// author valid
		{
			filePath: "schema_tests/author/valid/missing_author.json",
			valid:    true,
			content: `{
    "name": "a-happy-tyler",
    "description": "Tyler is happy. Be like Tyler.",
    "keywords": [
        "tyler",
        "happy"
    ],
    "version": "123",
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
			filePath: "schema_tests/author/valid/valid_author.json",
			valid:    true,
			content: `{
    "name": "a-happy-tyler",
    "description": "Tyler is happy. Be like Tyler.",
    "keywords": [
        "tyler",
        "happy"
    ],
    "version": "123",
    "author": "Tyler Caslin",
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
		// author invalid
		{
			filePath: "schema_tests/author/invalid/empty_author.json",
			errors:   []string{"author: String length must be greater than or equal to 1"},
			content: `{
    "name": "a-happy-tyler",
    "description": "Tyler is happy. Be like Tyler.",
    "keywords": [
        "tyler",
        "happy"
    ],
    "version": "123",
    "author": "",
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
		// version valid
		{
			filePath: "schema_tests/version/valid/valid_version.json",
			valid:    true,
			content: `{
    "name": "a-happy-tyler",
    "description": "Tyler is happy. Be like Tyler.",
    "keywords": [
        "tyler",
        "happy"
    ],
    "version": "123",
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
		// version invalid
		{
			filePath: "schema_tests/version/invalid/empty_version.json",
			errors:   []string{"version: String length must be greater than or equal to 1"},
			content: `{
    "name": "a-happy-tyler",
    "description": "Tyler is happy. Be like Tyler.",
    "keywords": [
        "tyler",
        "happy"
    ],
    "version": "",
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
			filePath: "schema_tests/version/invalid/missing_version.json",
			errors:   []string{"(root): version is required"},
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
	}

	runSchemaTestCases(t, packages.NonHumanReadableSchema, cases)
}
