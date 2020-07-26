package main

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
)

type SchemaTestCase struct {
	filePath    string
	valid       bool
	invalidJSON bool
	errors      []string
}

func TestSchema(t *testing.T) {
	cases := []SchemaTestCase{
		// author valid
		{
			filePath: "schema_tests/authors/valid/missing_authors.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/authors/valid/missing_email.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/authors/valid/missing_url.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/authors/valid/multiple_authors.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/authors/valid/one_author.json",
			valid:    true,
		},
		// author invalid
		{
			filePath: "schema_tests/authors/invalid/additional_property.json",
			valid:    false,
			errors:   []string{"authors.0: Additional property github is not allowed"},
		},
		{
			filePath: "schema_tests/authors/invalid/duplicate_authors.json",
			valid:    false,
			errors:   []string{"authors: array items[0,1] must be unique"},
		},
		{
			filePath: "schema_tests/authors/invalid/empty_array.json",
			valid:    false,
			errors:   []string{"authors: Array must have at least 1 items"},
		},
		{
			filePath: "schema_tests/authors/invalid/empty_author_object.json",
			valid:    false,
			errors:   []string{"authors.0: name is required"},
		},
		{
			filePath: "schema_tests/authors/invalid/empty_email.json",
			valid:    false,
			errors:   []string{"authors.0.email: String length must be greater than or equal to 1"},
		},
		{
			filePath: "schema_tests/authors/invalid/empty_name.json",
			valid:    false,
			errors:   []string{"authors.0.name: String length must be greater than or equal to 1"},
		},
		{
			filePath: "schema_tests/authors/invalid/empty_url.json",
			valid:    false,
			errors:   []string{"authors.0.url: String length must be greater than or equal to 1"},
		},
		{
			filePath: "schema_tests/authors/invalid/one_author_no_name.json",
			valid:    false,
			errors:   []string{"authors.0: name is required"},
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
				assert.Contains(t, tc.errors, resErr.String())
			}
		})
	}
}
