package server

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// spaHandler serves the built frontend, falling back to index.html so that
// client-side routes such as /historical resolve on a full page load.
func spaHandler(dist fs.FS) http.Handler {
	files := http.FileServerFS(dist)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAsset := strings.HasPrefix(r.URL.Path, "/assets/")

		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if info, err := fs.Stat(dist, name); err != nil || info.IsDir() {
			// A missing asset must not fall back to index.html: the browser
			// would parse HTML as JavaScript.
			if isAsset {
				http.NotFound(w, r)
				return
			}
			serveIndex(w, dist)
			return
		}
		// Vite fingerprints asset filenames, so they can be cached forever.
		if isAsset {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		files.ServeHTTP(w, r)
	})
}

func serveIndex(w http.ResponseWriter, dist fs.FS) {
	index, err := fs.ReadFile(dist, "index.html")
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Write(index)
}
