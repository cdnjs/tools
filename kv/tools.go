package kv

import (
	"encoding/json"
	"fmt"
	"log"
	"path"
	"strings"

	"github.com/cdnjs/tools/compress"

	"github.com/cdnjs/tools/packages"
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
			util.Infof(ctx, "p(%d/%d) failed to get package %s: %s\n", i+1, len(pckgs), pckgname, readerr)
			sentry.NotifyError(fmt.Errorf("failed to get package from KV: %s: %s", pckgname, readerr))
			continue
		}

		versions := pckg.Versions()
		for j, version := range versions {
			util.Infof(ctx, "p(%d/%d) v(%d/%d) Inserting %s (%s)\n", i+1, len(pckgs), j+1, len(versions), *pckg.Name, version)
			dir := path.Join(basePath, *pckg.Name, version)
			_, _, _, err := InsertNewVersionToKV(ctx, *pckg.Name, version, dir, metaOnly)
			util.Check(err)
		}
	}
}

// InsertAggregateMetadataFromScratch is a helper tool to insert a number of packages' aggregated metadata
// into KV from scratch. The tool will scrape all metadata for each package from KV to create the aggregated entry.
func InsertAggregateMetadataFromScratch(logger *log.Logger, pckgs []string) {
	for i, pckgName := range pckgs {
		ctx := util.ContextWithEntries(util.GetStandardEntries(pckgName, logger)...)
		pckg, err := GetPackage(ctx, pckgName)
		if err != nil {
			util.Infof(ctx, "p(%d/%d) failed to get package %s: %s\n", i+1, len(pckgs), pckgName, err)
			sentry.NotifyError(fmt.Errorf("failed to get package from KV: %s: %s", pckgName, err))
			continue
		}

		util.Infof(ctx, "p(%d/%d) Fetching %s versions...\n", i+1, len(pckgs), *pckg.Name)
		versions, err := GetVersions(pckgName)
		util.Check(err)

		var assets []packages.Asset
		for j, version := range versions {
			util.Infof(ctx, "p(%d/%d) v(%d/%d) Fetching %s (%s)\n", i+1, len(pckgs), j+1, len(versions), *pckg.Name, version)
			files, err := GetVersion(ctx, version)
			util.Check(err)
			assets = append(assets, packages.Asset{
				Version: strings.TrimPrefix(version, pckgName+"/"),
				Files:   files,
			})
		}

		pckg.Assets = assets
		successfulWrites, err := writeAggregatedMetadata(ctx, pckg)
		util.Check(err)

		if len(successfulWrites) == 0 {
			panic(fmt.Sprintf("p(%d/%d) %s: failed to write aggregated metadata", i+1, len(pckgs), *pckg.Name))
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

// OutputAggregate outputs the aggregated metadata associated with a package.
func OutputAggregate(logger *log.Logger, pckgName string) {
	bytes, err := Read(pckgName, aggregatedMetadataNamespaceID)
	util.Check(err)

	uncompressed := compress.UnGzip(bytes)

	// check if it can unmarshal into a package successfully
	var p packages.Package
	util.Check(json.Unmarshal(uncompressed, &p))

	fmt.Printf("%s\n", uncompressed)
}
