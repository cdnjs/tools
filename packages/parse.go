package packages

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/cdnjs/tools/util"

	"github.com/pkg/errors"
)

// GetHumanPackageJSONFiles gets the paths of the human-readable JSON files from within the `packagesPath`.
func GetHumanPackageJSONFiles(ctx context.Context) []string {
	list, err := util.ListFilesGlob(ctx, util.GetHumanPackagesPath(), "*/*.json")
	util.Check(err)
	return list
}

// ReadPackageJSON parses a JSON file into a Package.
func ReadHumanPackageJSON(ctx context.Context, file string) (*Package, error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read %s", file)
	}

	return ReadPackageJSONBytes(ctx, file, bytes)
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
