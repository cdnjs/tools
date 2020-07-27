package packages

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/xeipuuv/gojsonschema"

	"github.com/cdnjs/tools/util"

	"github.com/pkg/errors"
)

// InvalidHumanReadableSchemaError represents a schema error
// for a human-readable package.
type InvalidHumanReadableSchemaError struct {
	Result *gojsonschema.Result
	err    error
}

// Error is used to satisfy the error interface.
func (i InvalidHumanReadableSchemaError) Error() string {
	return i.err.Error()
}

// GetHumanPackageJSONFiles gets the paths of the human-readable JSON files from within the `packagesPath`.
func GetHumanPackageJSONFiles(ctx context.Context) []string {
	list, err := util.ListFilesGlob(ctx, util.GetHumanPackagesPath(), "*/*.json")
	util.Check(err)
	return list
}

// ReadPackageJSON parses a JSON file into a Package.
// If the schema is invalid, *gojsonschema.Result will be non-nil.
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
		return nil, InvalidHumanReadableSchemaError{res, err}
	}

	// unmarshal JSON into package
	var p Package
	if err := json.Unmarshal(bytes, &p); err != nil {
		return nil, errors.Wrapf(err, "failed to parse %s", file)
	}

	p.ctx = ctx
	return &p, nil
}

// ReadPackageJSONBytes parses a JSON file as bytes into a Package.
func ReadPackageJSONBytes(ctx context.Context, file string, bytes []byte) (*Package, error) {
	var p Package
	if err := json.Unmarshal(bytes, &p); err != nil {
		return nil, errors.Wrapf(err, "failed to parse %s", file)
	}

	p.ctx = ctx
	return &p, nil
}
