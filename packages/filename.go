package packages

import (
	"context"
	"log"
	"math"

	"github.com/agnivade/levenshtein"
)

// Update the package's filename if the latest
// version does not contain the filename
// Note that if the filename is nil it will stay nil.
func UpdateFilenameIfMissing(ctx context.Context, pkg *Package, files []string) error {
	if len(files) == 0 {
		log.Printf("%s: KV version contains no files\n", *pkg.Name)
		return nil
	}

	if pkg.Filename != nil {
		// check if assets contains filename
		filename := *pkg.Filename
		for _, asset := range files {
			if asset == filename {
				return nil // filename included in latest version, so return
			}
		}

		// set filename to be the most similar string in []assets
		mostSimilar := getMostSimilarFilename(filename, files)
		pkg.Filename = &mostSimilar
		log.Printf("%s: Updated filename `%s` -> `%s`\n", *pkg.Name, filename, mostSimilar)
		return nil
	}
	log.Printf("%s: filename missing, so will stay missing.\n", *pkg.Name)
	return nil
}

// Gets the most similar filename to a target filename.
// The []string of alternatives must have at least one element.
func getMostSimilarFilename(target string, filenames []string) string {
	var mostSimilar string
	var minDist int = math.MaxInt32
	for _, f := range filenames {
		if dist := levenshtein.ComputeDistance(target, f); dist < minDist {
			mostSimilar = f
			minDist = dist
		}
	}
	return mostSimilar
}
