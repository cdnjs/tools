package packages

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/xeipuuv/gojsonschema"

	"github.com/cdnjs/tools/util"

	"github.com/pkg/errors"
)

// InvalidSchemaError represents a schema error
// for a human-readable package.
type InvalidSchemaError struct {
	Result *gojsonschema.Result
}

// Error is used to satisfy the error interface.
func (i InvalidSchemaError) Error() string {
	var errors []string
	for _, resErr := range i.Result.Errors() {
		errors = append(errors, resErr.String())
	}
	return strings.Join(errors, ",")
}

// GetHumanPackageJSONFiles gets the paths of the human-readable JSON files from within the `packagesPath`.
func GetHumanPackageJSONFiles(ctx context.Context) []string {
	list, err := util.ListFilesGlob(ctx, util.GetHumanPackagesPath(), "*/*.json")
	util.Check(err)
	return list
}

// ReadHumanPackageJSON parses a JSON file into a Package.
// It will validate the human-readable schema, returning an
// InvalidSchemaError if the schema is invalid.
func ReadHumanPackageJSON(ctx context.Context, file string) (*Package, error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read %s", file)
	}

	// validate against human readable JSON schema
	res, err := HumanReadableSchema.Validate(gojsonschema.NewBytesLoader(bytes))
	if err != nil {
		// invalid JSON
		return nil, errors.Wrapf(err, "failed to parse %s", file)
	}

	if !res.Valid() {
		// invalid schema, so return result and custom error
		return nil, InvalidSchemaError{res}
	}

	// unmarshal JSON into package
	var p Package
	if err := json.Unmarshal(bytes, &p); err != nil {
		return nil, errors.Wrapf(err, "failed to parse %s", file)
	}

	// if `authors` exists, parse `author` field
	if p.Authors != nil {
		author := parseAuthor(p.Authors)
		p.Author = &author
	}

	p.ctx = ctx
	return &p, nil
}

// If `authors` exists, we need to parse `author` field
// for legacy compatibility with API.
func parseAuthor(authors []Author) string {
	var authorStrings []string
	for _, author := range authors {
		authorString := *author.Name
		if author.Email != nil {
			authorString += fmt.Sprintf(" <%s>", *author.Email)
		}
		if author.URL != nil {
			authorString += fmt.Sprintf(" (%s)", *author.URL)
		}
		authorStrings = append(authorStrings, authorString)
	}
	return strings.Join(authorStrings, ",")
}

// ReadNonHumanPackageJSONBytes parses a JSON file as bytes into a Package.
// It will validate the non-human-readable schema, returning an
// InvalidSchemaError if the schema is invalid.
func ReadNonHumanPackageJSONBytes(ctx context.Context, file string, bytes []byte) (*Package, error) {
	// validate the non-human readable JSON schema
	res, err := NonHumanReadableSchema.Validate(gojsonschema.NewBytesLoader(bytes))
	if err != nil {
		// invalid JSON
		return nil, errors.Wrapf(err, "failed to parse %s", file)
	}

	if !res.Valid() {
		// invalid schema, so return result and custom error
		return nil, InvalidSchemaError{res}
	}

	var p Package
	if err := json.Unmarshal(bytes, &p); err != nil {
		return nil, errors.Wrapf(err, "failed to parse %s", file)
	}

	// schema is valid, but we still need to ensure there are either
	// both `author` and `authors` fields or neither
	authorsNil, authorNil := p.Authors == nil, p.Author == nil
	if authorsNil != authorNil {
		return nil, errors.Wrapf(err, "`author` and `authors` must be either both nil or both non-nil - %s", file)
	}

	if !authorsNil {
		// `authors` exists, so need to verify `author` is parsed correctly
		author := *p.Author
		parsedAuthor := parseAuthor(p.Authors)
		if author != parsedAuthor {
			return nil, fmt.Errorf("author parse: actual `%s` != expected `%s`", author, parsedAuthor)
		}
	}

	p.ctx = ctx
	return &p, nil
}
