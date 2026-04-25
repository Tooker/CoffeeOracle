package handlers

import (
	"net/http"
	"path"
	"path/filepath"
	"strings"
)

// NewImageHandler serves stored oracle upload images from /api/image/<uuid>.jpeg.
func NewImageHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		name := path.Base(strings.TrimPrefix(r.URL.Path, "/api/image/"))
		if name == "." || name == "/" || !strings.HasSuffix(name, ".jpeg") {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "image/jpeg")
		http.ServeFile(w, r, filepath.Join(imageStoreDir, name))
	})
}
