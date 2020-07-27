package main

import (
	"io/ioutil"
	"testing"

	"github.com/cdnjs/tools/packages"
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

func TestHumanReadableSchema(t *testing.T) {
	const (
		autoupdateSourceRegex = "^git|npm$"
		licenseRegex          = "^(\\(.+ OR .+\\)|[a-zA-Z0-9].*)$"
		nameRegex             = "^[a-zA-Z0-9._-]+$"
		repositoryTypeRegex   = "^git$"
		repositoryURLRegex    = "github\\.com[/|:]([\\w\\.-]+)/([\\w\\.-]+)/?"
	)

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
			content:  ``,
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
			content: ``,
		},
		{
			filePath:    "schema_tests/(root)/invalid/invalid_json.txt",
			invalidJSON: true,
			content:     ``,
		},
		// author valid
		{
			filePath: "schema_tests/authors/valid/missing_authors.json",
			valid:    true,
			content:  ``,
		},
		{
			filePath: "schema_tests/authors/valid/missing_email.json",
			valid:    true,
			content:  ``,
		},
		{
			filePath: "schema_tests/authors/valid/missing_url.json",
			valid:    true,
			content:  ``,
		},
		{
			filePath: "schema_tests/authors/valid/multiple_authors.json",
			valid:    true,
			content:  ``,
		},
		{
			filePath: "schema_tests/authors/valid/one_author.json",
			valid:    true,
			content:  ``,
		},
		// author invalid
		{
			filePath: "schema_tests/authors/invalid/additional_property.json",
			valid:    false,
			errors:   []string{"authors.0: Additional property github is not allowed"},
			content:  ``,
		},
		{
			filePath: "schema_tests/authors/invalid/duplicate_authors.json",
			valid:    false,
			errors:   []string{"authors: array items[0,1] must be unique"},
			content:  ``,
		},
		{
			filePath: "schema_tests/authors/invalid/empty_array.json",
			valid:    false,
			errors:   []string{"authors: Array must have at least 1 items"},
			content:  ``,
		},
		{
			filePath: "schema_tests/authors/invalid/empty_author_object.json",
			valid:    false,
			errors:   []string{"authors.0: name is required"},
			content:  ``,
		},
		{
			filePath: "schema_tests/authors/invalid/empty_email.json",
			valid:    false,
			errors:   []string{"authors.0.email: String length must be greater than or equal to 1"},
			content:  ``,
		},
		{
			filePath: "schema_tests/authors/invalid/empty_name.json",
			valid:    false,
			errors:   []string{"authors.0.name: String length must be greater than or equal to 1"},
			content:  ``,
		},
		{
			filePath: "schema_tests/authors/invalid/empty_url.json",
			valid:    false,
			errors:   []string{"authors.0.url: String length must be greater than or equal to 1"},
			content:  ``,
		},
		{
			filePath: "schema_tests/authors/invalid/one_author_no_name.json",
			valid:    false,
			errors:   []string{"authors.0: name is required"},
			content:  ``,
		},
		// autoupdate valid
		{
			filePath: "schema_tests/autoupdate/valid/empty_basepath.json",
			valid:    true,
			content:  ``,
		},
		{
			filePath: "schema_tests/autoupdate/valid/multiple_filemaps.json",
			valid:    true,
			content:  ``,
		},
		{
			filePath: "schema_tests/autoupdate/valid/multiple_files.json",
			valid:    true,
			content:  ``,
		},
		{
			filePath: "schema_tests/autoupdate/valid/source_git.json",
			valid:    true,
			content:  ``,
		},
		{
			filePath: "schema_tests/autoupdate/valid/source_npm.json",
			valid:    true,
			content:  ``,
		},
		// autoupdate invalid
		{
			filePath: "schema_tests/autoupdate/invalid/additional_properties.json",
			valid:    false,
			errors: []string{
				"autoupdate: Additional property repo is not allowed",
				"autoupdate.fileMap.0: Additional property directory is not allowed",
			},
			content: ``,
		},
		{
			filePath: "schema_tests/autoupdate/invalid/duplicate_filemap.json",
			valid:    false,
			errors:   []string{"autoupdate.fileMap: array items[0,1] must be unique"},
			content:  ``,
		},
		{
			filePath: "schema_tests/autoupdate/invalid/duplicate_files.json",
			valid:    false,
			errors:   []string{"autoupdate.fileMap.0.files: array items[0,1] must be unique"},
			content:  ``,
		},
		{
			filePath: "schema_tests/autoupdate/invalid/empty_file.json",
			valid:    false,
			errors:   []string{"autoupdate.fileMap.0.files.0: String length must be greater than or equal to 1"},
			content:  ``,
		},
		{
			filePath: "schema_tests/autoupdate/invalid/empty_filemap.json",
			valid:    false,
			errors:   []string{"autoupdate.fileMap: Array must have at least 1 items"},
			content:  ``,
		},
		{
			filePath: "schema_tests/autoupdate/invalid/empty_files.json",
			valid:    false,
			errors:   []string{"autoupdate.fileMap.0.files: Array must have at least 1 items"},
			content:  ``,
		},
		{
			filePath: "schema_tests/autoupdate/invalid/empty_source.json",
			valid:    false,
			errors:   []string{"autoupdate.source: Does not match pattern '" + autoupdateSourceRegex + "'"},
			content:  ``,
		},
		{
			filePath: "schema_tests/autoupdate/invalid/empty_target.json",
			valid:    false,
			errors:   []string{"autoupdate.target: String length must be greater than or equal to 1"},
			content:  ``,
		},
		{
			filePath: "schema_tests/autoupdate/invalid/missing_autoupdate.json",
			valid:    false,
			errors:   []string{"(root): autoupdate is required"},
			content:  ``,
		},
		{
			filePath: "schema_tests/autoupdate/invalid/missing_basepath.json",
			valid:    false,
			errors:   []string{"autoupdate.fileMap.0: basePath is required"},
			content:  ``,
		},
		{
			filePath: "schema_tests/autoupdate/invalid/missing_filemap.json",
			valid:    false,
			errors:   []string{"autoupdate: fileMap is required"},
			content:  ``,
		},
		{
			filePath: "schema_tests/autoupdate/invalid/missing_files.json",
			valid:    false,
			errors:   []string{"autoupdate.fileMap.0: files is required"},
			content:  ``,
		},
		{
			filePath: "schema_tests/autoupdate/invalid/missing_source.json",
			valid:    false,
			errors:   []string{"autoupdate: source is required"},
			content:  ``,
		},
		{
			filePath: "schema_tests/autoupdate/invalid/missing_target.json",
			valid:    false,
			errors:   []string{"autoupdate: target is required"},
			content:  ``,
		},
		{
			filePath: "schema_tests/autoupdate/invalid/source_svn.json",
			valid:    false,
			errors:   []string{"autoupdate.source: Does not match pattern '" + autoupdateSourceRegex + "'"},
			content:  ``,
		},
		// description valid
		{
			filePath: "schema_tests/description/valid/valid_description.json",
			valid:    true,
			content:  ``,
		},
		// description invalid
		{
			filePath: "schema_tests/description/invalid/empty_description.json",
			errors:   []string{"description: String length must be greater than or equal to 1"},
			content:  ``,
		},
		{
			filePath: "schema_tests/description/invalid/missing_description.json",
			errors:   []string{"(root): description is required"},
			content:  ``,
		},
		// filename valid
		{
			filePath: "schema_tests/filename/valid/valid_filename.json",
			valid:    true,
			content:  ``,
		},
		// filename invalid
		{
			filePath: "schema_tests/filename/invalid/empty_filename.json",
			errors:   []string{"filename: String length must be greater than or equal to 1"},
			content:  ``,
		},
		{
			filePath: "schema_tests/filename/invalid/missing_filename.json",
			errors:   []string{"(root): filename is required"},
			content:  ``,
		},
		// homepage valid
		{
			filePath: "schema_tests/homepage/valid/valid_homepage.json",
			valid:    true,
			content:  ``,
		},
		{
			filePath: "schema_tests/homepage/valid/missing_homepage.json",
			valid:    true,
			content:  ``,
		},
		// homepage invalid
		{
			filePath: "schema_tests/homepage/invalid/empty_homepage.json",
			errors:   []string{"homepage: String length must be greater than or equal to 1"},
			content:  ``,
		},
		// keywords valid
		{
			filePath: "schema_tests/keywords/valid/multiple_keywords.json",
			valid:    true,
			content:  ``,
		},
		{
			filePath: "schema_tests/keywords/valid/one_keyword.json",
			valid:    true,
			content:  ``,
		},
		// keywords invalid
		{
			filePath: "schema_tests/keywords/invalid/duplicate_keywords.json",
			errors:   []string{"keywords: array items[0,1] must be unique"},
			content:  ``,
		},
		{
			filePath: "schema_tests/keywords/invalid/empty_keyword.json",
			errors:   []string{"keywords.0: String length must be greater than or equal to 1"},
			content:  ``,
		},
		{
			filePath: "schema_tests/keywords/invalid/empty_keywords.json",
			errors:   []string{"keywords: Array must have at least 1 items"},
			content:  ``,
		},
		{
			filePath: "schema_tests/keywords/invalid/missing_keywords.json",
			errors:   []string{"(root): keywords is required"},
			content:  ``,
		},
		// license valid
		{
			filePath: "schema_tests/license/valid/many_licenses.json",
			valid:    true,
			content:  ``,
		},
		{
			filePath: "schema_tests/license/valid/missing_license.json",
			valid:    true,
			content:  ``,
		},
		{
			filePath: "schema_tests/license/valid/single_license.json",
			valid:    true,
			content:  ``,
		},
		{
			filePath: "schema_tests/license/valid/two_licenses.json",
			valid:    true,
			content:  ``,
		},
		// license invalid
		{
			filePath: "schema_tests/license/invalid/empty_license.json",
			errors:   []string{"license: Does not match pattern '" + licenseRegex + "'"},
			content:  ``,
		},
		{
			filePath: "schema_tests/license/invalid/invalid_multiple_licenses.json",
			errors:   []string{"license: Does not match pattern '" + licenseRegex + "'"},
			content:  ``,
		},
		{
			filePath: "schema_tests/license/invalid/invalid_single_license.json",
			errors:   []string{"license: Does not match pattern '" + licenseRegex + "'"},
			content:  ``,
		},
		// name valid
		{
			filePath: "schema_tests/name/valid/valid_name.json",
			valid:    true,
			content:  ``,
		},
		// name invalid
		{
			filePath: "schema_tests/name/invalid/empty_name.json",
			errors:   []string{"name: Does not match pattern '" + nameRegex + "'"},
			content:  ``,
		},
		{
			filePath: "schema_tests/name/invalid/invalid_name.json",
			errors:   []string{"name: Does not match pattern '" + nameRegex + "'"},
			content:  ``,
		},
		{
			filePath: "schema_tests/name/invalid/missing_name.json",
			errors:   []string{"(root): name is required"},
			content:  ``,
		},
		// repository valid
		{
			filePath: "schema_tests/repository/valid/type_git.json",
			valid:    true,
			content:  ``,
		},
		// repository invalid
		{
			filePath: "schema_tests/repository/invalid/additional_property.json",
			errors:   []string{"repository: Additional property custom_type is not allowed"},
			content:  ``,
		},
		{
			filePath: "schema_tests/repository/invalid/empty_repository.json",
			errors: []string{
				"repository: type is required",
				"repository: url is required",
			},
			content: ``,
		},
		{
			filePath: "schema_tests/repository/invalid/empty_type.json",
			errors:   []string{"repository.type: Does not match pattern '" + repositoryTypeRegex + "'"},
			content:  ``,
		},
		{
			filePath: "schema_tests/repository/invalid/empty_url.json",
			errors:   []string{"repository.url: Does not match pattern '" + repositoryURLRegex + "'"},
			content:  ``,
		},
		{
			filePath: "schema_tests/repository/invalid/missing_repository.json",
			errors:   []string{"(root): repository is required"},
			content:  ``,
		},
		{
			filePath: "schema_tests/repository/invalid/missing_type.json",
			errors:   []string{"repository: type is required"},
			content:  ``,
		},
		{
			filePath: "schema_tests/repository/invalid/missing_url.json",
			errors:   []string{"repository: url is required"},
			content:  ``,
		},
		{
			filePath: "schema_tests/repository/invalid/type_git_invalid_url.json",
			errors:   []string{"repository.url: Does not match pattern '" + repositoryURLRegex + "'"},
			content:  ``,
		},
		{
			filePath: "schema_tests/repository/invalid/type_npm.json",
			errors: []string{
				"repository.type: Does not match pattern '" + repositoryTypeRegex + "'",
				"repository.url: Does not match pattern '" + repositoryURLRegex + "'",
			},
			content: ``,
		},
		{
			filePath: "schema_tests/repository/invalid/type_svn.json",
			errors: []string{
				"repository.type: Does not match pattern '" + repositoryTypeRegex + "'",
				"repository.url: Does not match pattern '" + repositoryURLRegex + "'",
			},
			content: ``,
		},
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
			res, err := packages.HumanReadableSchema.Validate(gojsonschema.NewBytesLoader(testBytes))
			// validate test file against schema
			// res, err := packages.HumanReadableSchema.Validate(gojsonschema.NewStringLoader(tc.content))
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
