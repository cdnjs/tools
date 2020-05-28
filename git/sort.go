package git

import (
	"github.com/blang/semver"
)

type ByGitVersion []GitVersion

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
