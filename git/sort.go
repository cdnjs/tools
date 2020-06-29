package git

import (
	"sort"

	"github.com/blang/semver"
)

// SortByTimeStamp sorts a []Version, ordering
// from most recent to least recent time stamps.
func SortByTimeStamp(vs []Version) {
	sort.Slice(vs, func(i, j int) bool {
		return vs[i].TimeStamp.After(vs[j].TimeStamp)
	})
}

// ByGitVersion implements sort.Interface for []Version
type ByGitVersion []Version

func (a ByGitVersion) Len() int      { return len(a) }
func (a ByGitVersion) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByGitVersion) Less(i, j int) bool {
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
