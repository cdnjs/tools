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
            
        },
        "filename": {},
        "homepage": {},
        "keywords": {},
        "license": {},
        "name": {},
        "repository": {}
    },
    "required": [],
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
