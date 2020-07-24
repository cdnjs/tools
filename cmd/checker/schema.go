package main

import (
	"fmt"

	"github.com/cdnjs/tools/util"
	"github.com/xeipuuv/gojsonschema"
)

const (
	validateAgainst = `
	{
		"author": {
    		"name": "Tyler Caslin",
    		"email": "tylercaslin47@gmail.com"
  		}
	}
	`
	schema = `
{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "type": "object",
    "properties": {
        "author": {
            "description": "The attributed author for the library, as defined in the cdnjs package JSON file for this library.",
            "type": "object",
            "properties": {
                "email": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "phone": {
                    "type": "string"
                },
                "twitter": {
                    "type": "string"
                },
                "url": {
                    "type": "string"
                }
            },
            "additionalProperties": false
        },
        "autoupdate": {
            "description": "Subscribes the package to an autoupdating service when a new version is released.",
            "type": "object",
            "properties": {
                "source": {
                    "type": "string"
                },
                "target": {
                    "type": "string"
                },
                "fileMap": {
                    "type": "array",
                    "properties": {
                        "basePath": {
                            "type": "string"
                        },
                        "files": {
                            "type": "array",
                            "minItems": 1,
                            "uniqueItems": true,
                            "items": {
                                "type": "string"
                            }
                        }
                    },
                    "required": [
                        "basePath",
                        "files"
                    ],
                    "additionalProperties": false
                }
            },
            "required": [
                "source",
                "target",
                "fileMap"
            ],
            "additionalProperties": false
        },
        "description": {
            "description": "The description of the library if it has been provided in the cdnjs package JSON file.",
            "type": "string"
        },
        "filename": {
            "description": "This will be the name of the default file for the library.",
            "type": "string"
        },
        "homepage": {
            "description": "A link to the homepage of the package, if one is defined in the cdnjs package JSON file. Normally, this is either the package repository or the package website.",
            "type": "string"
        },
        "keywords": {
            "description": "An array of keywords provided in the cdnjs package JSON for the library.",
            "type": "array",
            "items": {
                "type": "string"
            }
        },
        "license": {
            "description": "The license defined for the library on cdnjs, as a string. If the library has a custom license, it may not be shown here.",
            "type": "string"
        },
        "name": {
            "description": "This will be the full name of the library, as stored on cdnjs.",
            "type": "string"
        },
        "repository": {
            "description": "The repository for the library, if known, in standard repository format.",
            "type": "object",
            "properties": {
                "type": {
                    "type": "string"
                },
                "url": {
                    "type": "string"
                },
                "docs": {
                    "type": "string"
                }
            },
            "required": [
                "type",
                "url"
            ],
            "additionalProperties": false
        }
    },
    "required": [
        "filename",
        "name"
    ],
    "additionalProperties": false
}`
)

func test() {
	// ensure license is valid spdx

	s, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(schema))
	util.Check(err)

	res, err := s.Validate(gojsonschema.NewStringLoader(validateAgainst))
	util.Check(err)

	fmt.Println(res.Valid(), res.Errors())

	// input := gojsonschema.NewStringLoader(validateAgainst)

	// res, err := gojsonschema.Validate(s, input)
	// fmt.Println(res, err)
	// s, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(schema))
	// util.Check(err)

	// gojsonschema.FormatCheckers.Add()
	// res, err := s.Validate(`something`)
	// util.Check(err)
}
