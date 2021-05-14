package packages

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path"

	"github.com/cdnjs/tools/util"

	"github.com/blang/semver"
)

// Author represents an author.
type Author struct {
	Name  *string `json:"name,omitempty"`
	Email *string `json:"email,omitempty"`
	URL   *string `json:"url,omitempty"`
}

// Autoupdate is used to update particular files from
// a source type located at a target destination.
type Autoupdate struct {
	Source  *string   `json:"source,omitempty"`
	Target  *string   `json:"target,omitempty"`
	FileMap []FileMap `json:"fileMap,omitempty"`
}

// Optimization is used to enable/disable optimization
// for particular file types. By default, we will optimize all files.
type Optimization struct {
	JS  *bool `json:"js,omitempty"`
	CSS *bool `json:"css,omitempty"`
	PNG *bool `json:"png,omitempty"`
	JPG *bool `json:"jpg,omitempty"`
}

// Js returns if we should optimize JavaScript files.
func (o *Optimization) Js() bool {
	return o.JS == nil || *o.JS
}

// Css returns if we should optimize CSS files.
func (o *Optimization) Css() bool {
	return o.CSS == nil || *o.CSS
}

// Png returns if we should optimize PNG files.
func (o *Optimization) Png() bool {
	return o.PNG == nil || *o.PNG
}

// Jpg returns if we should optimize JPG/JPEG files.
func (o *Optimization) Jpg() bool {
	return o.JPG == nil || *o.JPG
}

// FileMap represents a number of files located
// under a base path.
type FileMap struct {
	BasePath *string  `json:"basePath"` // can be empty
	Files    []string `json:"files,omitempty"`
}

// Repository represents a repository.
type Repository struct {
	Type *string `json:"type,omitempty"`
	URL  *string `json:"url,omitempty"`
}

// Package holds metadata about a package.
// Its human-readable properties come from cdnjs/packages.
// The additional properties are used to manage the package.
// Any legacy properties are used to avoid any breaking changes in the API.
type Package struct {
	ctx context.Context // context

	// human-readable properties
	Authors      []Author      `json:"authors,omitempty"`
	Autoupdate   *Autoupdate   `json:"autoupdate,omitempty"`
	Optimization *Optimization `json:"optimization,omitempty"`
	Description  *string       `json:"description,omitempty"`
	Filename     *string       `json:"filename,omitempty"`
	Homepage     *string       `json:"homepage,omitempty"`
	Keywords     []string      `json:"keywords,omitempty"`
	License      *string       `json:"license,omitempty"`
	Name         *string       `json:"name,omitempty"`
	Repository   *Repository   `json:"repository,omitempty"`

	// additional properties
	Version *string `json:"version,omitempty"`

	// legacy
	Author *string `json:"author,omitempty"`

	// for aggregated metadata entries
	Assets []Asset `json:"assets,omitempty"`
}

// String represents the package as its marshalled JSON form.
func (p *Package) String() string {
	bytes, err := p.Marshal()
	if err != nil {
		return err.Error()
	}
	return string(bytes)
}

// Marshal marshals the package into JSON, not escaping HTML characters.
func (p *Package) Marshal() ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(p); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// // Log logs an event to cdnjs/logs.
// // It returns the updated log file as well as if any error occurred.
// // For `My-Package` on July 31, 2020, log file path will be:
// // `<cdnjs logs path>/m/My-Package/2020/07-31.log`
// func (p *Package) Log(format string, a ...interface{}) string {
// 	t := time.Now().UTC()
// 	logFileDir := path.Join(util.GetLogsPath(), "packages", strings.ToLower(string((*p.Name)[0])), *p.Name, t.Format("2006"))
// 	logFile := t.Format("01-02.log")
// 	logFilePath := path.Join(logFileDir, logFile)

// 	util.Debugf(p.ctx, "Logging to %s: %s\n", logFilePath, fmt.Sprintf(format, a...))

// 	// make dir path
// 	util.Check(os.MkdirAll(logFileDir, 0755))

// 	// open log file
// 	f, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
// 	util.Check(err)
// 	defer f.Close()

// 	// append to log file
// 	logger := log.New(f, fmt.Sprintf("%s %s: ", t.Format("2006-01-02 15:04:05"), *p.Name), 0)
// 	logger.Printf(format, a...)

// 	return logFilePath
// }

// LatestVersionKVKey gets the key needed to get the latest KV version metadata.
func (p *Package) LatestVersionKVKey() string {
	return path.Join(*p.Name, *p.Version)
}

// // LibraryPath returns the location of the package in the cdnjs repo.
// func (p *Package) LibraryPath() string {
// 	return path.Join(util.GetCDNJSLibrariesPath(), *p.Name)
// }

// TODO: Remove when no longer writing files to disk, and all files can
// be removed after temporarily processed and uploaded to KV.

// NpmFileMoveOp represents an operation to move files
// from a source destination to a target destination.
type NpmFileMoveOp struct {
	From string
	To   string
}

func (p *Package) HasVersion(name string) bool {
	for _, asset := range p.Assets {
		if asset.Version == name {
			return true
		}
	}
	return false
}

func (p *Package) UpdateVersion(name string, newAsset Asset) {
	for i, asset := range p.Assets {
		if asset.Version == name {
			p.Assets[i] = newAsset
			return
		}
	}
}

// NpmFilesFrom lists files that match the npm glob pattern in the `base` directory
// Returns a struct that represent the move semantics
func (p *Package) NpmFilesFrom(base string) []NpmFileMoveOp {
	out := make([]NpmFileMoveOp, 0)

	// map used to determine if a file path has already been processed
	seen := make(map[string]bool)

	for _, fileMap := range p.Autoupdate.FileMap {
		for _, pattern := range fileMap.Files {
			basePath := path.Join(base, *fileMap.BasePath)

			// find files that match glob
			list, err := util.ListFilesGlob(p.ctx, basePath, pattern)
			util.Check(err) // should have already run before in checker so panic if glob invalid

			for _, f := range list {
				fp := path.Join(basePath, f)

				// check if file has been processed before
				if _, ok := seen[fp]; ok {
					continue
				}
				seen[fp] = true

				info, staterr := os.Stat(fp)
				if staterr != nil {
					util.Warnf(p.ctx, "stat: "+staterr.Error())
					continue
				}

				// warn for files with sizes exceeding max file size
				size := info.Size()
				if size > util.MaxFileSize {
					util.Warnf(p.ctx, "file %s ignored due to byte size (%d > %d)", f, size, util.MaxFileSize)
					continue
				}

				// file is ok
				out = append(out, NpmFileMoveOp{
					From: path.Join(*fileMap.BasePath, f),
					To:   f,
				})
			}
		}
	}

	return out
}

// // AllFiles lists all files in the version directory.
// func (p *Package) AllFiles(version string) []string {
// 	out := make([]string, 0)

// 	absPath := path.Join(p.LibraryPath(), version)
// 	list, err := util.ListFilesInVersion(p.ctx, absPath)
// 	util.Check(err)
// 	out = append(out, list...)

// 	return out
// }

// Asset associates a number of files as strings
// with a version.
type Asset struct {
	Version string   `json:"version"`
	Files   []string `json:"files"`
}

// A "stable" version is considered to be a version that contains no pre-releases.
// If no latest stable version is found (ex. all are non-semver), a nil *string
// will be returned.
func GetLatestStableVersion(versions []string) *string {
	var latest *semver.Version
	for _, version := range versions {
		if s, err := semver.Parse(version); err == nil && len(s.Pre) == 0 {
			if latest != nil {
				if latest.LT(s) {
					latest = &s
				}
			} else {
				latest = &s
			}
		}
	}
	if latest != nil {
		s := latest.String()
		return &s
	} else {
		return nil
	}
}
