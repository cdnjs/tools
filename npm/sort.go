package npm

import (
	"github.com/blang/semver"
)

// ByNpmVersion implements sort.Interface for []Version
type ByNpmVersion []Version

func (a ByNpmVersion) Len() int      { return len(a) }
func (a ByNpmVersion) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByNpmVersion) Less(i, j int) bool {
	left, leftErr := semver.Make(a[i].Version)
	if leftErr != nil {
		return false
	}
	right, rightErr := semver.Make(a[j].Version)
	if rightErr != nil {
		return true
	}
	return left.Compare(right) == 1
}

// ByNpmVersionString implements sort.Interface for []String
type ByNpmVersionString []string

func (a ByNpmVersionString) Len() int      { return len(a) }
func (a ByNpmVersionString) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByNpmVersionString) Less(i, j int) bool {
	left, leftErr := semver.Make(a[i])
	if leftErr != nil {
		return false
	}
	right, rightErr := semver.Make(a[j])
	if rightErr != nil {
		return true
	}
	return left.Compare(right) == 1
}
