package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
)

type SchemaTestCase struct {
	filePath    string
	valid       bool
	invalidJSON bool
	errors      []string
	content     string
}

func runSchemaTestCases(t *testing.T, schema *gojsonschema.Schema, cases []SchemaTestCase) {
	for _, tc := range cases {
		tc := tc // capture range variable

		// since all tests share the same input, this needs to run sequentially
		t.Run(tc.filePath, func(t *testing.T) {

			// validate test file against schema
			res, err := schema.Validate(gojsonschema.NewStringLoader(tc.content))
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
