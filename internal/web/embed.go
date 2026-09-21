package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var content embed.FS

func Assets() (fs.FS, error) { return fs.Sub(content, "dist") }
