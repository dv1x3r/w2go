package w2lib

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed *
var libFS embed.FS

func FS() fs.FS {
	return libFS
}

func FileServerFS() http.Handler {
	return http.FileServerFS(libFS)
}
