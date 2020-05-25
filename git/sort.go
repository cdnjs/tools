package git

import (
	"github.com/blang/semver"
)

type ByGitVersion []string

func (a ByGitVersion) Len() int      { return len(a) }
func (a ByGitVersion) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByGitVersion) Less(i, j int) bool {
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

// ByGitVersionString implements sort.Interface for []String
type ByGitVersionString []string

func (a ByGitVersionString) Len() int      { return len(a) }
func (a ByGitVersionString) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByGitVersionString) Less(i, j int) bool {
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
