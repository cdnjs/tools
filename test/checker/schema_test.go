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
		// top-level valid
		{
			filePath: "schema_tests/top_level/valid/all_properties.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/top_level/valid/only_required_properties.json",
			valid:    true,
		},
		// top-level invalid
		{
			filePath: "schema_tests/top_level/invalid/additional_properties.json",
			valid:    false,
			errors: []string{
				"(root): Additional property licenses is not allowed",
				"(root): Additional property author is not allowed",
				"(root): Additional property my_custom_property is not allowed",
			},
		},
		{
			filePath:    "schema_tests/top_level/invalid/invalid_json.txt",
			invalidJSON: true,
		},
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
		// autoupdate valid
		{
			filePath: "schema_tests/autoupdate/valid/empty_basepath.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/autoupdate/valid/multiple_filemaps.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/autoupdate/valid/multiple_files.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/autoupdate/valid/source_git.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/autoupdate/valid/source_npm.json",
			valid:    true,
		},
		// autoupdate invalid
		{
			filePath: "schema_tests/autoupdate/invalid/additional_properties.json",
			valid:    false,
			errors: []string{
				"autoupdate: Additional property repo is not allowed",
				"autoupdate.fileMap.0: Additional property directory is not allowed",
			},
		},
		{
			filePath: "schema_tests/autoupdate/invalid/duplicate_filemap.json",
			valid:    false,
			errors:   []string{"autoupdate.fileMap: array items[0,1] must be unique"},
		},
		{
			filePath: "schema_tests/autoupdate/invalid/duplicate_files.json",
			valid:    false,
			errors:   []string{"autoupdate.fileMap.0.files: array items[0,1] must be unique"},
		},
		{
			filePath: "schema_tests/autoupdate/invalid/empty_file.json",
			valid:    false,
			errors:   []string{"autoupdate.fileMap.0.files.0: String length must be greater than or equal to 1"},
		},
		{
			filePath: "schema_tests/autoupdate/invalid/empty_filemap.json",
			valid:    false,
			errors:   []string{"autoupdate.fileMap: Array must have at least 1 items"},
		},
		{
			filePath: "schema_tests/autoupdate/invalid/empty_files.json",
			valid:    false,
			errors:   []string{"autoupdate.fileMap.0.files: Array must have at least 1 items"},
		},
		{
			filePath: "schema_tests/autoupdate/invalid/empty_source.json",
			valid:    false,
			errors:   []string{"autoupdate.source: Does not match pattern '^(git|npm)$'"},
		},
		{
			filePath: "schema_tests/autoupdate/invalid/empty_target.json",
			valid:    false,
			errors:   []string{"autoupdate.target: String length must be greater than or equal to 1"},
		},
		{
			filePath: "schema_tests/autoupdate/invalid/missing_autoupdate.json",
			valid:    false,
			errors:   []string{"(root): autoupdate is required"},
		},
		{
			filePath: "schema_tests/autoupdate/invalid/missing_basepath.json",
			valid:    false,
			errors:   []string{"autoupdate.fileMap.0: basePath is required"},
		},
		{
			filePath: "schema_tests/autoupdate/invalid/missing_filemap.json",
			valid:    false,
			errors:   []string{"autoupdate: fileMap is required"},
		},
		{
			filePath: "schema_tests/autoupdate/invalid/missing_files.json",
			valid:    false,
			errors:   []string{"autoupdate.fileMap.0: files is required"},
		},
		{
			filePath: "schema_tests/autoupdate/invalid/missing_source.json",
			valid:    false,
			errors:   []string{"autoupdate: source is required"},
		},
		{
			filePath: "schema_tests/autoupdate/invalid/missing_target.json",
			valid:    false,
			errors:   []string{"autoupdate: target is required"},
		},
		{
			filePath: "schema_tests/autoupdate/invalid/source_svn.json",
			valid:    false,
			errors:   []string{"autoupdate.source: Does not match pattern '^(git|npm)$'"},
		},
		// description valid
		{
			filePath: "schema_tests/description/valid/valid_description.json",
			valid:    true,
		},
		// description invalid
		{
			filePath: "schema_tests/description/invalid/empty_description.json",
			errors:   []string{"description: String length must be greater than or equal to 1"},
		},
		{
			filePath: "schema_tests/description/invalid/missing_description.json",
			errors:   []string{"(root): description is required"},
		},
		// filename valid
		{
			filePath: "schema_tests/filename/valid/valid_filename.json",
			valid:    true,
		},
		// filename invalid
		{
			filePath: "schema_tests/filename/invalid/empty_filename.json",
			errors:   []string{"filename: String length must be greater than or equal to 1"},
		},
		{
			filePath: "schema_tests/filename/invalid/missing_filename.json",
			errors:   []string{"(root): filename is required"},
		},
		// homepage valid
		{
			filePath: "schema_tests/homepage/valid/valid_homepage.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/homepage/valid/missing_homepage.json",
			valid:    true,
		},
		// homepage invalid
		{
			filePath: "schema_tests/homepage/invalid/empty_homepage.json",
			errors:   []string{"homepage: String length must be greater than or equal to 1"},
		},
		// keywords valid
		{
			filePath: "schema_tests/keywords/valid/multiple_keywords.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/keywords/valid/one_keyword.json",
			valid:    true,
		},
		// keywords invalid
		{
			filePath: "schema_tests/keywords/invalid/duplicate_keywords.json",
			errors:   []string{"keywords: array items[0,1] must be unique"},
		},
		{
			filePath: "schema_tests/keywords/invalid/empty_keyword.json",
			errors:   []string{"keywords.0: String length must be greater than or equal to 1"},
		},
		{
			filePath: "schema_tests/keywords/invalid/empty_keywords.json",
			errors:   []string{"keywords: Array must have at least 1 items"},
		},
		{
			filePath: "schema_tests/keywords/invalid/missing_keywords.json",
			errors:   []string{"(root): keywords is required"},
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
