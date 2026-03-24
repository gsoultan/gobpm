package ui

import (
	"embed"
	"io/fs"
)

// Build timestamp: 2026-03-09 12:05:00-v2-refined
//
//go:embed all:dist
var distFS embed.FS

// Dist returns the embedded dist filesystem.
func Dist() fs.FS {
	f, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}
	return f
}
