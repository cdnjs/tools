package compress

import (
	"encoding/json"
	"os"
	"path"
)

func getNpmVersion(pkg string) string {
	type packageJSON struct {
		Version string `json:"version"`
	}

	data, err := os.ReadFile(path.Join("/node_modules", pkg, "package.json"))
	if err != nil {
		return "<failed to read version>"
	}

	var p packageJSON
	if err := json.Unmarshal(data, &p); err != nil {
		return "<failed to read version>"
	}

	return p.Version
}
