package ui

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed web/dist/*
var rawFS embed.FS

// GetStaticFS returns an http.FileSystem rooted at the dist directory.
func GetStaticFS() http.FileSystem {
	dist, err := fs.Sub(rawFS, "web/dist")
	if err != nil {
		panic(err)
	}
	return http.FS(dist)
}
