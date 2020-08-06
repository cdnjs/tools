package kv

import (
	"fmt"
	"log"
	"path"

	"github.com/cdnjs/tools/sentry"
	"github.com/cdnjs/tools/util"
)

// InsertFromDisk is a helper tool to insert a number of packages from disk.
// Note: Only inserting versions (not updating package metadata).
func InsertFromDisk(logger *log.Logger, pckgs []string, metaOnly bool) {
	basePath := util.GetCDNJSLibrariesPath()

	for i, pckgname := range pckgs {
		ctx := util.ContextWithEntries(util.GetStandardEntries(pckgname, logger)...)
		pckg, readerr := GetPackage(ctx, pckgname)
		if readerr != nil {
			util.Infof(ctx, "p(%d/%d) FAILED TO GET PACKAGE %s: %s\n", i+1, len(pckgs), pckgname, readerr)
			sentry.NotifyError(fmt.Errorf("failed to get package from KV: %s: %s", pckgname, readerr))
			continue
		}

		versions := pckg.Versions()
		for j, version := range versions {
			util.Infof(ctx, "p(%d/%d) v(%d/%d) Inserting %s (%s)\n", i+1, len(pckgs), j+1, len(versions), *pckg.Name, version)
			dir := path.Join(basePath, *pckg.Name, version)
			_, _, err := InsertNewVersionToKV(ctx, *pckg.Name, version, dir, metaOnly)
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
