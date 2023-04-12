package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed build/*
var staticFS embed.FS

// ----------------------------------------------------------------------
// staticFileSystem serves files out of the embedded build folder
type staticFileSystem struct {
	http.FileSystem
}

func NewStaticFileSystem() *staticFileSystem {
	sub, err := fs.Sub(staticFS, "build")

	if err != nil {
		panic(err)
	}

	return &staticFileSystem{
		FileSystem: http.FS(sub),
	}
}
