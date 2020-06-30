package packages

import (
	"github.com/blang/semver"
)

// ByVersionAsset implements sort.Interface for []Asset based on
// the Version field.
type ByVersionAsset []Asset

func (a ByVersionAsset) Len() int      { return len(a) }
func (a ByVersionAsset) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByVersionAsset) Less(i, j int) bool {
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

// ByVersionString implements sort.Interface for []string based on
// a number of semver strings.
type ByVersionString []string

func (a ByVersionString) Len() int      { return len(a) }
func (a ByVersionString) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByVersionString) Less(i, j int) bool {
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
