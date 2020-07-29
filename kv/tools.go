package kv

import (
	"fmt"
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
		pckg, readerr := packages.ReadHumanJSON(ctx, pckgname)
		util.Check(readerr)

		for _, version := range pckg.Versions() {
			util.Infof(ctx, "Inserting %s (%s)\n", *pckg.Name, version)
			dir := path.Join(basePath, *pckg.Name, version)
			err := InsertNewVersionToKV(ctx, *pckg.Name, version, dir)
			util.Check(err)
		}
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
