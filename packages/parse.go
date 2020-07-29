package packages

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
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

// GetHumanPackageJSONFiles gets the paths of the human-readable JSON files from within cdnjs/packages.
//
// TODO: update this to remove legacy ListFilesGlob
func GetHumanPackageJSONFiles(ctx context.Context) []string {
	list, err := util.ListFilesGlob(ctx, util.GetHumanPackagesPath(), "*/*.json")
	util.Check(err)
	return list
}

// ReadHumanJSON reads this package's human-readable JSON from within cdnjs/packages.
// It will validate the human-readable schema, returning an
// InvalidSchemaError if the schema is invalid.
func ReadHumanJSON(ctx context.Context, name string) (*Package, error) {
	return ReadHumanJSONFile(ctx, path.Join(util.GetHumanPackagesPath(), strings.ToLower(string(name[0])), name+".json"))
}

// ReadNonHumanJSON reads this package's non-human readable JSON.
// It will validate the non-human-readable schema, returning an
// InvalidSchemaError if the schema is invalid.
//
// TODO:
//
// UPDATE TO READ FROM KV.
func ReadNonHumanJSON(ctx context.Context, name string) (*Package, error) {
	return ReadNonHumanJSONFile(ctx, path.Join(util.GetCDNJSLibrariesPath(), name, "package.json"))
}

// readHumanJSONFile parses a JSON file into a Package.
// It will validate the human-readable schema, returning an
// InvalidSchemaError if the schema is invalid.
func ReadHumanJSONFile(ctx context.Context, file string) (*Package, error) {
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

	return readHumanJSONBytes(ctx, file, bytes)
}

// Unmarshals the human-readable JSON into a *Package,
// setting the legacy `author` field if needed.
func readHumanJSONBytes(ctx context.Context, file string, bytes []byte) (*Package, error) {
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

// readNonHumanJSONFile parses a JSON file into a Package.
// It will validate the non-human-readable schema, returning an
// InvalidSchemaError if the schema is invalid.
func ReadNonHumanJSONFile(ctx context.Context, file string) (*Package, error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read %s", file)
	}

	return ReadNonHumanJSONBytes(ctx, file, bytes)
}

// ReadNonHumanJSONBytes unmarshals bytes into a *Package,
// validating against the non-human-readable schema, returning an
// InvalidSchemaError if the schema is invalid.
func ReadNonHumanJSONBytes(ctx context.Context, name string, bytes []byte) (*Package, error) {
	// validate the non-human readable JSON schema
	res, err := NonHumanReadableSchema.Validate(gojsonschema.NewBytesLoader(bytes))
	if err != nil {
		// invalid JSON
		return nil, errors.Wrapf(err, "failed to parse %s", name)
	}

	if !res.Valid() {
		// invalid schema, so return result and custom error
		return nil, InvalidSchemaError{res}
	}

	var p Package
	if err := json.Unmarshal(bytes, &p); err != nil {
		return nil, errors.Wrapf(err, "failed to parse %s", name)
	}

	// schema is valid, but we still need to ensure there are either
	// both `author` and `authors` fields or neither
	authorsNil, authorNil := p.Authors == nil, p.Author == nil
	if authorsNil != authorNil {
		return nil, errors.Wrapf(err, "`author` and `authors` must be either both nil or both non-nil - %s", name)
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
