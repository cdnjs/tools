package main

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
)

const (
	autoupdateSourceRegex = "^git|npm$"
	licenseRegex          = "^(\\(.+ OR .+\\)|[a-zA-Z0-9].*)$"
	nameRegex             = "^[a-zA-Z0-9._-]+$"
	repositoryTypeRegex   = "^git$"
	repositoryURLRegex    = "github\\.com[/|:]([\\w\\.-]+)/([\\w\\.-]+)/?"
)

type SchemaTestCase struct {
	filePath    string
	valid       bool
	invalidJSON bool
	errors      []string
}

func runSchemaTestCases(t *testing.T, schema *gojsonschema.Schema, cases []SchemaTestCase) {
	for _, tc := range cases {
		tc := tc // capture range variable

		// since all tests share the same input, this needs to run sequentially
		t.Run(tc.filePath, func(t *testing.T) {
			// read bytes of test file
			testBytes, err := ioutil.ReadFile(tc.filePath)
			assert.Nil(t, err)
			if err != nil {
				return
			}

			// validate test file against schema
			res, err := schema.Validate(gojsonschema.NewBytesLoader(testBytes))
			if tc.invalidJSON {
				// error will be non-nil if JSON loading fails
				assert.NotNil(t, err)
				return
			}

			// JSON should load successfully
			assert.Nil(t, err)
			if err != nil {
				return
			}

			if tc.valid {
				// expecting no errors
				assert.True(t, res.Valid())
				// don't return here, since we want all errors to be outputted
				// in the case this assertion fails
			}

			// expecting errors
			resErrs := res.Errors()

			// check the number of errors
			assert.Equal(t, len(tc.errors), len(res.Errors()))

			// make sure all errors are accounted for
			for _, resErr := range resErrs {
				assert.Contains(t, tc.errors, resErr.String())
			}
		})
	}
}
