package packages

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

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
	return fmt.Sprintf("%v", i.Result)
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

	p.ctx = ctx
	return &p, nil
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
	if (p.Author != nil && p.Author == nil) || (p.Author == nil && p.Authors != nil) {
		return nil, errors.Wrapf(err, "`author` and `authors` must be either both nil or both non-nil - %s", file)
	}

	p.ctx = ctx
	return &p, nil
}
