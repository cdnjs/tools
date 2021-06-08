package version

import (
	"time"

	"github.com/cdnjs/tools/packages"

	"github.com/gobwas/glob"
)

// Version represents a version of a git repo or npm.
type Version struct {
	Version string
	Tarball string
	Date    time.Time
	Source  string // npm or git
}

func IsVersionIgnored(config *packages.Autoupdate, version string) bool {
	for _, ignored := range config.IgnoreVersions {
		g := glob.MustCompile(ignored)
		if g.Match(version) {
			return true
		}
	}
	return false
}

func VersionDiff(a []Version, b []string) []Version {
	diff := make([]Version, 0)
	m := make(map[string]bool)

	for _, item := range b {
		m[item] = true
	}

	for _, item := range a {
		if _, ok := m[item.Version]; !ok {
			diff = append(diff, item)
		}
	}

	return diff
}
