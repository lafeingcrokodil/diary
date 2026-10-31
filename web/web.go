// Package web provides the built front end, embedded into the Go binary.
package web

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed all:build
var build embed.FS

// Build returns the files of the built front end, with index.html at the root.
func Build() (fs.FS, error) {
	files, err := fs.Sub(build, "build")
	if err != nil {
		return nil, fmt.Errorf("opening embedded front end: %w", err)
	}

	return files, nil
}
