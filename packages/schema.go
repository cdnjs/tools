package packages

import (
	"github.com/cdnjs/tools/util"
	"github.com/xeipuuv/gojsonschema"
)

var (
	// HumanReadableSchema is the human-readable package schema used for
	// JSON files in cdnjs/packages.
	HumanReadableSchema = initHumanReadableSchema()

	// NonHumanReadableSchema is the non-human-readable package schema used for
	// storing metadata into KV.
	// This depends on HumanReadableSchema so that the unit tests can be shared.
	NonHumanReadableSchema = initNonHumanReadableSchema()
)

func initHumanReadableSchema() *gojsonschema.Schema {
	s, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(HumanReadableSchemaString))
	util.Check(err)
	return s
}
func initNonHumanReadableSchema() *gojsonschema.Schema {
	s, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(NonHumanReadableSchemaString))
	util.Check(err)
	return s
}

// HumanReadableSchemaString is the stringified human-readable package schema used for
// JSON files in cdnjs/packages.
const HumanReadableSchemaString = `{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "type": "object",
    "properties": {` + humanReadableProperties + `
    },
    "required": [
        "autoupdate",
        "description",
        "filename",
        "keywords",
        "name",
        "repository"
    ],
    "additionalProperties": false
}`

// NonHumanReadableSchemaString is the stringified non-human-readable package schema used for
// storing metadata into KV.
const NonHumanReadableSchemaString = `{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "type": "object",
    "properties": {` + humanReadableProperties + `,` + nonHumanReadableProperties + `
    },
    "required": [
        "description",
        "filename",
        "keywords",
        "name",
        "repository",
        "version"
    ],
    "additionalProperties": false
}`

const humanReadableProperties = `"authors": {
            "description": "The attributed author for the library, as defined in the cdnjs package JSON file for this library.",
            "type": "array",
            "minItems": 1,
            "uniqueItems": true,
            "items": {
                "type": "object",
                "properties": {
                    "email": {
                        "type": "string",
                        "minLength": 1
                    },
                    "name": {
                        "type": "string",
                        "minLength": 1
                    },
                    "url": {
                        "type": "string",
                        "minLength": 1
                    }
                },
                "additionalProperties": false,
                "required": [
                    "name"
                ]
            }
        },
        "autoupdate": {
            "description": "Subscribes the package to an autoupdating service when a new version is released.",
            "type": "object",
            "properties": {
                "fileMap": {
                    "type": "array",
                    "minItems": 1,
                    "uniqueItems": true,
                    "items": {
                        "type": "object",
                        "properties": {
                            "basePath": {
                                "type": "string"
                            },
                            "files": {
                                "type": "array",
                                "minItems": 1,
                                "uniqueItems": true,
                                "items": {
                                    "type": "string",
                                    "minLength": 1
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
                "source": {
                    "type": "string",
                    "pattern": "^git|npm$"
                },
                "target": {
                    "type": "string",
                    "minLength": 1
                }
            },
            "required": [
                "fileMap",
                "source",
                "target"
            ],
            "additionalProperties": false
        },
        "description": {
            "description": "The description of the library if it has been provided in the cdnjs package JSON file.",
            "type": "string",
            "minLength": 1
        },
        "filename": {
            "description": "This will be the name of the default file for the library.",
            "type": "string",
            "minLength": 1
        },
        "homepage": {
            "description": "A link to the homepage of the package, if one is defined in the cdnjs package JSON file. Normally, this is either the package repository or the package website.",
            "type": "string",
            "minLength": 1
        },
        "keywords": {
            "description": "An array of keywords provided in the cdnjs package JSON for the library.",
            "type": "array",
            "minItems": 1,
            "uniqueItems": true,
            "items": {
                "type": "string",
                "minLength": 1
            }
        },
        "license": {
            "description": "The license defined for the library on cdnjs, as a string. If the library has a custom license, it may not be shown here.",
            "type": "string",
            "pattern": "^(\\(.+ OR .+\\)|[a-zA-Z0-9].*)$"
        },
        "name": {
            "description": "This will be the full name of the library, as stored on cdnjs.",
            "type": "string",
            "pattern": "^[a-zA-Z0-9._-]+$"
        },
        "repository": {
            "description": "The repository for the library, if known, in standard repository format.",
            "type": "object",
            "properties": {
                "type": {
                    "type": "string",
                    "pattern": "^git$"
                },
                "url": {
                    "type": "string",
                    "pattern": "github\\.com[/|:]([\\w\\.-]+)\/([\\w\\.-]+)\/?"
                }
            },
            "required": [
                "type",
                "url"
            ],
            "additionalProperties": false
        }`

const nonHumanReadableProperties = `
        "author": {
            "type": "string",
            "minLength": 1
        },
        "version": {
            "type": "string",
            "minLength": 1
        }`
