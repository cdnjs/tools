package version

import (
	"context"
	"log"
)

// ByTimeStamp implements the sort.Interface for []Version,
// ordering from most recent to least recent time stamps.
type ByDate []Version

func (a ByDate) Len() int      { return len(a) }
func (a ByDate) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByDate) Less(i, j int) bool {
	return a[i].Date.After(a[j].Date)
}

// GetMostRecentExistingVersion gets the most recent npm.Version based on time stamp
// that is currently downloaded as well as all existing versions in npm.Version form.
func GetMostRecentExistingVersion(ctx context.Context, existingVersions []string, npmVersions []Version) (*Version, []Version) {
	// create map for fast lookups
	npmMap := make(map[string]Version)
	for _, v := range npmVersions {
		npmMap[v.Version] = v
	}

	var allExisting []Version

	// find most recent version
	var mostRecent *Version
	for _, existingVersion := range existingVersions {
		if version, ok := npmMap[existingVersion]; ok {
			allExisting = append(allExisting, version)
			if mostRecent == nil || version.Date.After(mostRecent.Date) {
				mostRecent = &version // new most recent found
			}
			continue
		}
		log.Printf("existing version not found on npm: %s", existingVersion)
	}

	return mostRecent, allExisting
}
