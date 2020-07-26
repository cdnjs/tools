package main

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
)

type SchemaTestCase struct {
	name        string
	filePath    string
	valid       bool
	invalidJSON bool
	errors      []string
}

func TestSchema(t *testing.T) {
	cases := []SchemaTestCase{
		{
			name:     "authors/valid: missing email",
			filePath: "schema_tests/authors/valid/missing_email.json",
			valid:    true,
		},
	}

	// read schema bytes
	schemaBytes, err := ioutil.ReadFile("../../schema.json")
	assert.Nil(t, err)
	if err != nil {
		return
	}

	// parse schema
	schema, err := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(schemaBytes))
	assert.Nil(t, err)
	if err != nil {
		return
	}

	// TODO: do I need these return stmts?

	for _, tc := range cases {
		tc := tc // capture range variable

		// since all tests share the same input, this needs to run sequentially
		t.Run(tc.name, func(t *testing.T) {
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
				// expect no errors
				assert.True(t, res.Valid())
				return
			}

			// expecting errors
			resErrs := res.Errors()

			// check the number of errors
			assert.Equal(t, len(tc.errors), len(res.Errors()))

			// make sure all errors are accounted for
			for _, resErr := range resErrs {
				assert.Contains(t, resErr.String(), tc.errors)
			}
		})
	}
}
