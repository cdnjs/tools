package kv

import (
	"log"
	"path"

	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/util"
)

// InsertFromDisk is a helper tool to insert a number of packages from disk.
// Note: Only inserting versions (not updating package metadata).
func InsertFromDisk(logger *log.Logger, pckgs []string) {
	basePath := util.GetCDNJSPackages()

	for _, pckgname := range pckgs {
		ctx := util.ContextWithEntries(util.GetStandardEntries(pckgname, logger)...)
		pckg, readerr := packages.ReadPackageJSON(ctx, path.Join(basePath, pckgname, "package.json"))
		util.Check(readerr)

		for _, version := range pckg.Versions() {
			util.Infof(ctx, "Inserting %s (%s)\n", pckg.Name, version)
			dir := path.Join(basePath, pckg.Name, version)
			err := InsertNewVersionToKV(ctx, pckg.Name, version, dir)
			util.Check(err)
		}
	}
}

// InsertMetadataFromDisk is a helper tool to insert a number of packages' respective metadata from disk.
// It will read the respective `package.json` files in `cdnjs/cdnjs/` and insert them directly to KV.
// Note: In the future, the `package.json` files will be removed completely, and when dealing with
// new packages we will read the respective JSON file in `cdnjs/packages` with the `version` attribute appended.
func InsertMetadataFromDisk(logger *log.Logger, pckgs []string) {
	basePath := util.GetCDNJSPackages()

	for _, pckgname := range pckgs {
		ctx := util.ContextWithEntries(util.GetStandardEntries(pckgname, logger)...)
		pckg, readerr := packages.ReadPackageJSON(ctx, path.Join(basePath, pckgname, "package.json"))
		util.Check(readerr)

		err := UpdateKVPackage(ctx, pckg)
		util.Check(err)
	}
}
