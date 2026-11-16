package main

import (
	"testing"

	"github.com/cdnjs/tools/version"
	"github.com/stretchr/testify/assert"
)

func versionNames(versions []version.Version) []string {
	names := make([]string, 0, len(versions))
	for _, v := range versions {
		names = append(names, v.Version)
	}
	return names
}

func TestMoveVersionToFront(t *testing.T) {
	versions := []version.Version{{Version: "14.3.3"}, {Version: "15.0.3"}, {Version: "14.3.2"}}

	assert.Equal(t, []string{"15.0.3", "14.3.3", "14.3.2"}, versionNames(moveVersionToFront(versions, "15.0.3")))
	assert.Equal(t, []string{"14.3.3", "15.0.3", "14.3.2"}, versionNames(moveVersionToFront(versions, "14.3.3")))
	assert.Equal(t, []string{"14.3.3", "15.0.3", "14.3.2"}, versionNames(moveVersionToFront(versions, "9.9.9")))
	assert.Equal(t, []string{"14.3.3", "15.0.3", "14.3.2"}, versionNames(versions), "input is not modified")
}
