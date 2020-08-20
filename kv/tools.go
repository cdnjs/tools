package kv

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/cdnjs/tools/compress"

	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/sentry"
	"github.com/cdnjs/tools/util"
)

// InsertFromDisk is a helper tool to insert a number of packages from disk.
// Note: Only inserting versions (not updating package metadata).
func InsertFromDisk(logger *log.Logger, pckgs []string, metaOnly, srisOnly, filesOnly bool) {
	basePath := util.GetCDNJSLibrariesPath()

	var wg sync.WaitGroup
	done := make(chan string)

	log.Println("Starting...")

	for index, name := range pckgs {
		wg.Add(1)
		go func(i int, pckgName string) {
			defer wg.Done()
			defer func() { done <- pckgName }()

			ctx := util.ContextWithEntries(util.GetStandardEntries(pckgName, logger)...)
			pckg, readerr := GetPackage(ctx, pckgName)
			if readerr != nil {
				util.Infof(ctx, "p(%d/%d) failed to get package %s: %s\n", i+1, len(pckgs), pckgName, readerr)
				sentry.NotifyError(fmt.Errorf("failed to get package from KV: %s: %s", pckgName, readerr))
				return
			}

			versions := pckg.Versions()
			for j, version := range versions {
				util.Debugf(ctx, "p(%d/%d) v(%d/%d) Inserting %s (%s)\n", i+1, len(pckgs), j+1, len(versions), *pckg.Name, version)
				dir := path.Join(basePath, *pckg.Name, version)
				_, _, _, _, err := InsertNewVersionToKV(ctx, *pckg.Name, version, dir, metaOnly, srisOnly, filesOnly)
				if err != nil {
					util.Infof(ctx, "p(%d/%d) v(%d/%d) failed to insert %s (%s): %s\n", i+1, len(pckgs), j+1, len(versions), *pckg.Name, version, err)
					sentry.NotifyError(fmt.Errorf("p(%d/%d) v(%d/%d) failed to insert %s (%s) to KV: %s\n", i+1, len(pckgs), j+1, len(versions), *pckg.Name, version, err))
					return
				}
			}
		}(index, name)
	}

	// show some progress
	go func() {
		i := 0
		for {
			name := <-done
			i++
			log.Printf("Completed (%d/%d): %s\n", i, len(pckgs), name)
		}
	}()

	wg.Wait()
	log.Println("Done.")
}

// InsertAggregateMetadataFromScratch is a helper tool to insert a number of packages' aggregated metadata
// into KV from scratch. The tool will scrape all metadata for each package from KV to create the aggregated entry.
func InsertAggregateMetadataFromScratch(logger *log.Logger, pckgs []string) {
	var wg sync.WaitGroup
	done := make(chan bool)

	log.Println("Starting...")
	for index, name := range pckgs {
		wg.Add(1)
		go func(i int, pckgName string) {
			defer wg.Done()
			defer func() { done <- true }()

			ctx := util.ContextWithEntries(util.GetStandardEntries(pckgName, logger)...)
			pckg, err := GetPackage(ctx, pckgName)
			if err != nil {
				util.Infof(ctx, "p(%d/%d) failed to get package %s: %s\n", i+1, len(pckgs), pckgName, err)
				sentry.NotifyError(fmt.Errorf("failed to get package from KV: %s: %s", pckgName, err))
				return
			}

			util.Debugf(ctx, "p(%d/%d) Fetching %s versions...\n", i+1, len(pckgs), *pckg.Name)
			versions, err := GetVersions(pckgName)
			util.Check(err)

			var assets []packages.Asset
			for j, version := range versions {
				util.Debugf(ctx, "p(%d/%d) v(%d/%d) Fetching %s (%s)\n", i+1, len(pckgs), j+1, len(versions), *pckg.Name, version)
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
				util.Infof(ctx, "p(%d/%d) %s: failed to write aggregated metadata", i+1, len(pckgs), *pckg.Name)
				sentry.NotifyError(fmt.Errorf("p(%d/%d) %s: failed to write aggregated metadata", i+1, len(pckgs), *pckg.Name))
			}
		}(index, name)
	}

	// show some progress
	go func() {
		i := 0
		for {
			<-done
			i++
			log.Printf("Completed (%d/%d)\n", i, len(pckgs))
		}
	}()

	wg.Wait()
	log.Println("Done.")
}

// OutputAllAggregatePackages outputs all the names of all aggregated package metadata entries in KV.
func OutputAllAggregatePackages() {
	res, err := listByPrefixNamesOnly("", aggregatedMetadataNamespaceID)
	util.Check(err)

	bytes, err := json.Marshal(res)
	util.Check(err)

	fmt.Printf("%s\n", bytes)
}

// OutputAllPackages outputs the names of all packages in KV.
func OutputAllPackages() {
	res, err := listByPrefixNamesOnly("", packagesNamespaceID)
	util.Check(err)

	bytes, err := json.Marshal(res)
	util.Check(err)

	fmt.Printf("%s\n", bytes)
}

// OutputFile outputs a file stored in KV.
func OutputFile(logger *log.Logger, fileKey string, ungzip, unbrotli bool) {
	ctx := util.ContextWithEntries(util.GetStandardEntries(fileKey, logger)...)

	util.Infof(ctx, "Fetching file from KV...\n")
	bytes, err := read(fileKey, filesNamespaceID)
	util.Check(err)

	if ungzip {
		util.Infof(ctx, "Decompressing gzip...\n")
		bytes = compress.UnGzip(bytes)
	} else if unbrotli {
		util.Infof(ctx, "Decompressing brotli...\n")
		file, err := ioutil.TempFile("", "")
		util.Check(err)
		defer os.Remove(file.Name())

		_, err = file.Write(bytes)
		util.Check(err)
		bytes = compress.UnBrotliCLI(ctx, file.Name())
	}

	fmt.Printf("%s\n", bytes)
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
func OutputAggregate(pckgName string) {
	bytes, err := read(pckgName, aggregatedMetadataNamespaceID)
	util.Check(err)

	uncompressed := compress.UnGzip(bytes)

	// check if it can unmarshal into a package successfully
	var p packages.Package
	util.Check(json.Unmarshal(uncompressed, &p))

	fmt.Printf("%s\n", uncompressed)
}

// OutputSRIs lists the SRIs namespace by prefix.
func OutputSRIs(prefix string) {
	res, err := listByPrefix(prefix, srisNamespaceID)
	util.Check(err)

	sris := make(map[string]string)
	for _, r := range res {
		sris[r.Name] = r.Metadata.(map[string]interface{})["sri"].(string)
	}

	bytes, err := json.Marshal(sris)
	util.Check(err)

	fmt.Printf("%s\n", bytes)
}
