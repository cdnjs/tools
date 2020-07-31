package kv

import (
	"fmt"
	"log"
	"path"

	"github.com/cdnjs/tools/util"
)

// InsertFromDisk is a helper tool to insert a number of packages from disk.
// Note: Only inserting versions (not updating package metadata).
func InsertFromDisk(logger *log.Logger, pckgs []string) {
	basePath := util.GetCDNJSLibrariesPath()

	for _, pckgname := range pckgs {
		ctx := util.ContextWithEntries(util.GetStandardEntries(pckgname, logger)...)
		pckg, readerr := GetPackage(ctx, pckgname)
		util.Check(readerr)

		for _, version := range pckg.Versions() {
			util.Infof(ctx, "Inserting %s (%s)\n", *pckg.Name, version)
			dir := path.Join(basePath, *pckg.Name, version)
			err, _ := InsertNewVersionToKV(ctx, *pckg.Name, version, dir)
			util.Check(err)
		}
	}
}

// OutputAllFiles outputs all files stored in KV for a particular package.
func OutputAllFiles(logger *log.Logger, pckgName string) {
	ctx := util.ContextWithEntries(util.GetStandardEntries(pckgName, logger)...)

	// output all file names for each version in KV
	if versions, err := GetVersions(pckgName); err != nil {
		util.Infof(ctx, "Failed to get versions: %s\n", err)
	} else {
		for i, v := range versions {
			if files, err := GetFiles(v); err != nil {
				util.Infof(ctx, "(%d/%d) Failed to get version: %s\n", i+1, len(versions), err)
			} else {
				var output string
				if len(files) > 25 {
					output = fmt.Sprintf("(%d files)", len(files))
				} else {
					output = fmt.Sprintf("%v", files)
				}
				util.Infof(ctx, "(%d/%d) Found %s: %s\n", i+1, len(versions), v, output)
			}
		}
	}
}

// OutputAllMeta outputs all metadata associated with a package.
func OutputAllMeta(logger *log.Logger, pckgName string) {
	ctx := util.ContextWithEntries(util.GetStandardEntries(pckgName, logger)...)

	// output package metadata
	if pckg, err := GetPackage(ctx, pckgName); err != nil {
		util.Infof(ctx, "Failed to get package meta: %s\n", err)
	} else {
		util.Infof(ctx, "Parsed package: %s\n", pckg)
	}

	// output versions metadata
	if versions, err := GetVersions(pckgName); err != nil {
		util.Infof(ctx, "Failed to get versions: %s\n", err)
	} else {
		for i, v := range versions {
			if assets, err := GetVersion(ctx, v); err != nil {
				util.Infof(ctx, "(%d/%d) Failed to get version: %s\n", i+1, len(versions), err)
			} else {
				var output string
				if len(assets) > 25 {
					output = fmt.Sprintf("(%d assets)", len(assets))
				} else {
					output = fmt.Sprintf("%v", assets)
				}
				util.Infof(ctx, "(%d/%d) Parsed %s: %s\n", i+1, len(versions), v, output)
			}
		}
	}
}
