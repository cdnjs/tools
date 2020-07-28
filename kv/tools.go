package kv

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path"

	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/util"
)

// InsertFromDisk is a helper tool to insert a number of packages from disk.
// Note: Only inserting versions (not updating package metadata).
func InsertFromDisk(logger *log.Logger, pckgs []string) {
	basePath := util.GetCDNJSLibrariesPath()

	for _, pckgname := range pckgs {
		ctx := util.ContextWithEntries(util.GetStandardEntries(pckgname, logger)...)
		pckg, readerr := packages.ReadHumanPackageJSON(ctx, path.Join(basePath, pckgname, "package.json"))
		util.Check(readerr)

		for _, version := range pckg.Versions() {
			util.Infof(ctx, "Inserting %s (%s)\n", *pckg.Name, version)
			dir := path.Join(basePath, *pckg.Name, version)
			err := InsertNewVersionToKV(ctx, *pckg.Name, version, dir)
			util.Check(err)
		}
	}
}

// InsertMetadataFromDisk is a helper tool to insert a number of packages' respective non-human-readable metadata from disk.
// It will read the respective version in the `package.json` files in `cdnjs/cdnjs/` as well as the main
// metadata in cdnjs/packages/.
func InsertMetadataFromDisk(logger *log.Logger, pckgs []string) {
	for _, pckgname := range pckgs {
		humanPath := path.Join(util.GetHumanPackagesPath(), string(pckgname[0]), pckgname+".json")
		nonHumanPath := path.Join(util.GetCDNJSLibrariesPath(), pckgname, "package.json")

		// parse human-readable
		ctx := util.ContextWithEntries(util.GetStandardEntries(pckgname, logger)...)
		pckg, readerr := packages.ReadHumanPackageJSON(ctx, humanPath)
		util.Check(readerr)

		// parse non-human-readable and assume it is in legacy format
		legacyPkg := make(map[string]interface{})
		bytes, err := ioutil.ReadFile(nonHumanPath)
		util.Check(err)
		util.Check(json.Unmarshal(bytes, &legacyPkg))

		// add version field to human-readable package
		version, ok := legacyPkg["version"]
		if !ok {
			panic(fmt.Sprintf("no `version` in %s", nonHumanPath))
		}
		versionString, ok := version.(string)
		if !ok {
			panic(fmt.Sprintf("`version` is not a string in %s", nonHumanPath))
		}
		pckg.Version = &versionString

		// insert to KV
		util.Infof(ctx, "Inserting package metadata: %s\n", *pckg.Name)
		err = UpdateKVPackage(ctx, pckg)
		util.Check(err)
	}
}

// OutputAllMeta is a helper tool to output all metadata associated with a package.
func OutputAllMeta(logger *log.Logger, pckgname string) {
	ctx := util.ContextWithEntries(util.GetStandardEntries(pckgname, logger)...)

	// output package metadata
	if pckg, err := GetPackage(ctx, pckgname); err != nil {
		util.Infof(ctx, "Failed to get package meta: %s\n", err)
	} else {
		util.Infof(ctx, "Parsed package: %s\n", pckg)
	}

	// output versions metadata
	if versions, err := GetVersions(pckgname); err != nil {
		util.Infof(ctx, "Failed to get versions: %s\n", err)
	} else {
		for i, v := range versions {
			if version, err := GetVersion(ctx, v); err != nil {
				util.Infof(ctx, "(%d/%d) Failed to get version: %s\n", i+1, len(versions), err)
			} else {
				var output string
				if len(version) > 25 {
					output = fmt.Sprintf("(%d assets)", len(version))
				} else {
					output = fmt.Sprintf("%v", version)
				}
				util.Infof(ctx, "(%d/%d) Parsed %s: %s\n", i+1, len(versions), v, output)
			}
		}
	}
}
