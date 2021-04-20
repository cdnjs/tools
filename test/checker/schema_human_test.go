package main

import (
	"testing"

	"github.com/cdnjs/tools/packages"
)

func TestHumanReadableSchema(t *testing.T) {
	cases := []SchemaTestCase{
		// (root) valid
		{
			filePath: "schema_tests/human_schema_tests/(root)/valid/all_properties.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/human_schema_tests/(root)/valid/only_required_properties.json",
			valid:    true,
		},
		// (root) invalid
		{
			filePath: "schema_tests/human_schema_tests/(root)/invalid/additional_properties.json",
			valid:    false,
			errors: []string{
				"(root): Additional property licenses is not allowed",
				"(root): Additional property author is not allowed",
				"(root): Additional property my_custom_property is not allowed",
			},
		},
		{
			filePath:    "schema_tests/human_schema_tests/(root)/invalid/invalid_json.txt",
			invalidJSON: true,
		},
		// author valid
		{
			filePath: "schema_tests/human_schema_tests/authors/valid/missing_authors.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/human_schema_tests/authors/valid/missing_email.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/human_schema_tests/authors/valid/missing_url.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/human_schema_tests/authors/valid/multiple_authors.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/human_schema_tests/authors/valid/one_author.json",
			valid:    true,
		},
		// author invalid
		{
			filePath: "schema_tests/human_schema_tests/authors/invalid/additional_property.json",
			errors:   []string{"authors.0: Additional property github is not allowed"},
		},
		{
			filePath: "schema_tests/human_schema_tests/authors/invalid/duplicate_authors.json",
			errors:   []string{"authors: array items[0,1] must be unique"},
		},
		{
			filePath: "schema_tests/human_schema_tests/authors/invalid/empty_array.json",
			errors:   []string{"authors: Array must have at least 1 items"},
		},
		{
			filePath: "schema_tests/human_schema_tests/authors/invalid/empty_author_object.json",
			errors:   []string{"authors.0: name is required"},
		},
		{
			filePath: "schema_tests/human_schema_tests/authors/invalid/empty_email.json",
			errors:   []string{"authors.0.email: String length must be greater than or equal to 1"},
		},
		{
			filePath: "schema_tests/human_schema_tests/authors/invalid/empty_name.json",
			errors:   []string{"authors.0.name: String length must be greater than or equal to 1"},
		},
		{
			filePath: "schema_tests/human_schema_tests/authors/invalid/empty_url.json",
			errors:   []string{"authors.0.url: String length must be greater than or equal to 1"},
		},
		{
			filePath: "schema_tests/human_schema_tests/authors/invalid/one_author_no_name.json",
			errors:   []string{"authors.0: name is required"},
		},
		// autoupdate valid
		{
			filePath: "schema_tests/human_schema_tests/autoupdate/valid/empty_basepath.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/human_schema_tests/autoupdate/valid/multiple_filemaps.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/human_schema_tests/autoupdate/valid/multiple_files.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/human_schema_tests/autoupdate/valid/source_git.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/human_schema_tests/autoupdate/valid/source_npm.json",
			valid:    true,
		},
		// autoupdate invalid
		{
			filePath: "schema_tests/human_schema_tests/autoupdate/invalid/additional_properties.json",
			errors: []string{
				"autoupdate: Additional property repo is not allowed",
				"autoupdate.fileMap.0: Additional property directory is not allowed",
			},
		},
		{
			filePath: "schema_tests/human_schema_tests/autoupdate/invalid/duplicate_filemap.json",
			errors:   []string{"autoupdate.fileMap: array items[0,1] must be unique"},
		},
		{
			filePath: "schema_tests/human_schema_tests/autoupdate/invalid/duplicate_files.json",
			errors:   []string{"autoupdate.fileMap.0.files: array items[0,1] must be unique"},
		},
		{
			filePath: "schema_tests/human_schema_tests/autoupdate/invalid/empty_file.json",
			errors:   []string{"autoupdate.fileMap.0.files.0: String length must be greater than or equal to 1"},
		},
		{
			filePath: "schema_tests/human_schema_tests/autoupdate/invalid/empty_filemap.json",
			errors:   []string{"autoupdate.fileMap: Array must have at least 1 items"},
		},
		{
			filePath: "schema_tests/human_schema_tests/autoupdate/invalid/empty_files.json",
			errors:   []string{"autoupdate.fileMap.0.files: Array must have at least 1 items"},
		},
		{
			filePath: "schema_tests/human_schema_tests/autoupdate/invalid/empty_source.json",
			errors:   []string{"autoupdate.source: Does not match pattern '" + autoupdateSourceRegex + "'"},
		},
		{
			filePath: "schema_tests/human_schema_tests/autoupdate/invalid/empty_target.json",
			errors:   []string{"autoupdate.target: String length must be greater than or equal to 1"},
		},
		{
			filePath: "schema_tests/human_schema_tests/autoupdate/invalid/missing_autoupdate.json",
			errors:   []string{"(root): autoupdate is required"},
		},
		{
			filePath: "schema_tests/human_schema_tests/autoupdate/invalid/missing_basepath.json",
			errors:   []string{"autoupdate.fileMap.0: basePath is required"},
		},
		{
			filePath: "schema_tests/human_schema_tests/autoupdate/invalid/missing_filemap.json",
			errors:   []string{"autoupdate: fileMap is required"},
		},
		{
			filePath: "schema_tests/human_schema_tests/autoupdate/invalid/missing_files.json",
			errors:   []string{"autoupdate.fileMap.0: files is required"},
		},
		{
			filePath: "schema_tests/human_schema_tests/autoupdate/invalid/missing_source.json",
			errors:   []string{"autoupdate: source is required"},
		},
		{
			filePath: "schema_tests/human_schema_tests/autoupdate/invalid/missing_target.json",
			errors:   []string{"autoupdate: target is required"},
		},
		{
			filePath: "schema_tests/human_schema_tests/autoupdate/invalid/source_svn.json",
			errors:   []string{"autoupdate.source: Does not match pattern '" + autoupdateSourceRegex + "'"},
		},
		// description valid
		{
			filePath: "schema_tests/human_schema_tests/description/valid/valid_description.json",
			valid:    true,
		},
		// description invalid
		{
			filePath: "schema_tests/human_schema_tests/description/invalid/empty_description.json",
			errors:   []string{"description: String length must be greater than or equal to 1"},
		},
		{
			filePath: "schema_tests/human_schema_tests/description/invalid/missing_description.json",
			errors:   []string{"(root): description is required"},
		},
		// filename valid
		{
			filePath: "schema_tests/human_schema_tests/filename/valid/missing_filename.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/human_schema_tests/filename/valid/valid_filename.json",
			valid:    true,
		},
		// filename invalid
		{
			filePath: "schema_tests/human_schema_tests/filename/invalid/empty_filename.json",
			errors:   []string{"filename: String length must be greater than or equal to 1"},
		},
		// homepage valid
		{
			filePath: "schema_tests/human_schema_tests/homepage/valid/missing_homepage.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/human_schema_tests/homepage/valid/valid_homepage.json",
			valid:    true,
		},
		// homepage invalid
		{
			filePath: "schema_tests/human_schema_tests/homepage/invalid/empty_homepage.json",
			errors:   []string{"homepage: String length must be greater than or equal to 1"},
		},
		// keywords valid
		{
			filePath: "schema_tests/human_schema_tests/keywords/valid/multiple_keywords.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/human_schema_tests/keywords/valid/one_keyword.json",
			valid:    true,
		},
		// keywords invalid
		{
			filePath: "schema_tests/human_schema_tests/keywords/invalid/duplicate_keywords.json",
			errors:   []string{"keywords: array items[0,1] must be unique"},
		},
		{
			filePath: "schema_tests/human_schema_tests/keywords/invalid/empty_keyword.json",
			errors:   []string{"keywords.0: String length must be greater than or equal to 1"},
		},
		{
			filePath: "schema_tests/human_schema_tests/keywords/invalid/empty_keywords.json",
			errors:   []string{"keywords: Array must have at least 1 items"},
		},
		{
			filePath: "schema_tests/human_schema_tests/keywords/invalid/missing_keywords.json",
			errors:   []string{"(root): keywords is required"},
		},
		// license valid
		{
			filePath: "schema_tests/human_schema_tests/license/valid/many_licenses_and_or.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/human_schema_tests/license/valid/many_licenses_or.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/human_schema_tests/license/valid/missing_license.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/human_schema_tests/license/valid/single_license.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/human_schema_tests/license/valid/two_licenses.json",
			valid:    true,
		},
		// license invalid
		{
			filePath: "schema_tests/human_schema_tests/license/invalid/empty_license.json",
			errors:   []string{"license: Does not match pattern '" + licenseRegex + "'"},
		},
		{
			filePath: "schema_tests/human_schema_tests/license/invalid/invalid_multiple_licenses.json",
			errors:   []string{"license: Does not match pattern '" + licenseRegex + "'"},
		},
		{
			filePath: "schema_tests/human_schema_tests/license/invalid/invalid_single_license.json",
			errors:   []string{"license: Does not match pattern '" + licenseRegex + "'"},
		},
		// name valid
		{
			filePath: "schema_tests/human_schema_tests/name/valid/valid_name.json",
			valid:    true,
		},
		// name invalid
		{
			filePath: "schema_tests/human_schema_tests/name/invalid/empty_name.json",
			errors:   []string{"name: Does not match pattern '" + nameRegex + "'"},
		},
		{
			filePath: "schema_tests/human_schema_tests/name/invalid/invalid_name.json",
			errors:   []string{"name: Does not match pattern '" + nameRegex + "'"},
		},
		{
			filePath: "schema_tests/human_schema_tests/name/invalid/missing_name.json",
			errors:   []string{"(root): name is required"},
		},
		// repository valid
		{
			filePath: "schema_tests/human_schema_tests/repository/valid/type_git.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/human_schema_tests/repository/valid/type_hg.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/human_schema_tests/repository/valid/type_svn.json",
			valid:    true,
		},
		// repository invalid
		{
			filePath: "schema_tests/human_schema_tests/repository/invalid/additional_property.json",
			errors:   []string{"repository: Additional property custom_type is not allowed"},
		},
		{
			filePath: "schema_tests/human_schema_tests/repository/invalid/empty_repository.json",
			errors: []string{
				"repository: type is required",
				"repository: url is required",
			},
		},
		{
			filePath: "schema_tests/human_schema_tests/repository/invalid/empty_type.json",
			errors:   []string{"repository.type: Does not match pattern '" + repositoryTypeRegex + "'"},
		},
		{
			filePath: "schema_tests/human_schema_tests/repository/invalid/empty_url.json",
			errors:   []string{"repository.url: String length must be greater than or equal to 1"},
		},
		{
			filePath: "schema_tests/human_schema_tests/repository/invalid/missing_repository.json",
			errors:   []string{"(root): repository is required"},
		},
		{
			filePath: "schema_tests/human_schema_tests/repository/invalid/missing_type.json",
			errors:   []string{"repository: type is required"},
		},
		{
			filePath: "schema_tests/human_schema_tests/repository/invalid/missing_url.json",
			errors:   []string{"repository: url is required"},
		},
		{
			filePath: "schema_tests/human_schema_tests/repository/invalid/type_npm.json",
			errors:   []string{"repository.type: Does not match pattern '" + repositoryTypeRegex + "'"},
		},
		// optimization valid
		{
			filePath: "schema_tests/human_schema_tests/optimization/valid/all_fields.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/human_schema_tests/optimization/valid/empty_optimization.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/human_schema_tests/optimization/valid/js_only.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/human_schema_tests/optimization/valid/no_optimization.json",
			valid:    true,
		},
		// optimization invalid
		{
			filePath: "schema_tests/human_schema_tests/optimization/invalid/invalid_key.json",
			errors:   []string{"optimization: Additional property happy is not allowed"},
		},
		{
			filePath: "schema_tests/human_schema_tests/optimization/invalid/not_boolean.json",
			errors:   []string{"optimization.js: Invalid type. Expected: boolean, given: string"},
		},
	}

	runSchemaTestCases(t, packages.HumanReadableSchema, cases)
}
