package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"path"

	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/sri"
	"github.com/pkg/errors"
)

func calcSriPackage(ctx context.Context, config *packages.Package) error {
	files, err := ioutil.ReadDir(OUTPUT)
	if err != nil {
		return errors.Wrap(err, "failed to list output files")
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filename := path.Join(OUTPUT, file.Name())
		ext := path.Ext(filename)
		if _, ok := calculateSRI[ext]; ok {
			outSRI := fmt.Sprintf("%s.sri", filename)
			sri.CalculateFileSRI(filename, outSRI)
			log.Printf("sri %s -> %s\n", filename, outSRI)
		}
	}

	return nil
}
