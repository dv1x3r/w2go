package w2ui

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed w2ui.*
var w2uiFS embed.FS

func FS() fs.FS {
	return w2uiFS
}

func FileServerFS() http.Handler {
	return http.FileServerFS(w2uiFS)
}
