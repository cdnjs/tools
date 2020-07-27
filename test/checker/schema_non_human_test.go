package main

import (
	"testing"

	"github.com/cdnjs/tools/packages"
)

func TestNonHumanReadableSchema(t *testing.T) {
	cases := []SchemaTestCase{
		// author valid
		{
			filePath: "schema_tests/non_human_schema_tests/author/valid/missing_author.json",
			valid:    true,
		},
		{
			filePath: "schema_tests/non_human_schema_tests/author/valid/valid_author.json",
			valid:    true,
		},
		// author invalid
		{
			filePath: "schema_tests/non_human_schema_tests/author/invalid/empty_author.json",
			errors:   []string{"author: String length must be greater than or equal to 1"},
		},
		// autoupdate valid
		{
			filePath: "schema_tests/non_human_schema_tests/autoupdate/valid/missing_autoupdate.json",
			valid:    true,
		},
		// repository valid
		{
			filePath: "schema_tests/non_human_schema_tests/repository/valid/missing_repository.json",
			valid:    true,
		},
		// version valid
		{
			filePath: "schema_tests/non_human_schema_tests/version/valid/valid_version.json",
			valid:    true,
		},
		// version invalid
		{
			filePath: "schema_tests/non_human_schema_tests/version/invalid/empty_version.json",
			errors:   []string{"version: String length must be greater than or equal to 1"},
		},
		{
			filePath: "schema_tests/non_human_schema_tests/version/invalid/missing_version.json",
			errors:   []string{"(root): version is required"},
		},
	}

	runSchemaTestCases(t, packages.NonHumanReadableSchema, cases)
}
