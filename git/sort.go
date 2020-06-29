package git

import (
	"context"
	"fmt"
)

// ByTimeStamp implements the sort.Interface for []Version,
// ordering from most recent to least recent time stamps.
type ByTimeStamp []Version

func (a ByTimeStamp) Len() int      { return len(a) }
func (a ByTimeStamp) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByTimeStamp) Less(i, j int) bool {
	return a[i].TimeStamp.After(a[j].TimeStamp)
}

// GetMostRecentExistingVersion gets the most recent git.Version based on time stamp
// that is currently downloaded.
func GetMostRecentExistingVersion(ctx context.Context, existingVersions []string, gitVersions []Version) *Version {
	// create map for fast lookups
	gitMap := make(map[string]Version)
	for _, v := range gitVersions {
		gitMap[v.Version] = v
	}

	// find most recent version
	var mostRecent *Version
	for _, existingVersion := range existingVersions {
		if version, ok := gitMap[existingVersion]; ok {
			if mostRecent == nil || version.TimeStamp.After(mostRecent.TimeStamp) {
				mostRecent = &version // new most recent found
			}
			continue
		}
		// util.Debugf(ctx, "existing version not found on git: %s", existingVersion)
		panic(fmt.Sprintf("existing version not found on git: %s", existingVersion))
	}

	return mostRecent
}
