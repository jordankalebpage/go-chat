package gochat

import (
	"embed"
	"io/fs"
)

//go:embed web/dist web/dist/*
var webAssets embed.FS

func StaticFS() (fs.FS, error) {
	return fs.Sub(webAssets, "web/dist")
}
