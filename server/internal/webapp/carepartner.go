// Package webapp serves the carepartner app's web build, which the caregiver
// app shows in an iframe as its preview of a carepartner's screen.
package webapp

import (
	"bytes"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// CarepartnerPath is where the carepartner web build is served. It must match
// CAREPARTNER_WEB_BASE_URL in server/Dockerfile, which the build's asset and
// route addresses start with.
const CarepartnerPath = "/carepartner"

// Carepartner serves the carepartner web build (`expo export --platform web`,
// static output) found in files, for requests under CarepartnerPath:
//
//   - a file in the build is served as it is (`/carepartner/_expo/...`);
//   - a route is served its page (`/carepartner/pair` → `pair.html`, `/carepartner/` → `index.html`);
//   - anything else gets `+not-found.html` with a 404.
//
// `/carepartner` redirects to `/carepartner/`, keeping the query string. Every
// response may be framed by pages from this origin (the caregiver app), and no
// other.
func Carepartner(files fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "frame-ancestors 'self'")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")

		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path == CarepartnerPath {
			target := CarepartnerPath + "/"
			if r.URL.RawQuery != "" {
				target += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, target, http.StatusMovedPermanently)
			return
		}

		rel, ok := strings.CutPrefix(r.URL.Path, CarepartnerPath+"/")
		if !ok {
			http.NotFound(w, r)
			return
		}
		name := strings.TrimPrefix(path.Clean("/"+rel), "/")
		if name == "" {
			name = "index"
		}

		for _, candidate := range []string{name, name + ".html"} {
			if serveFile(w, r, files, candidate, http.StatusOK) {
				return
			}
		}
		if !serveFile(w, r, files, "+not-found.html", http.StatusNotFound) {
			http.NotFound(w, r)
		}
	})
}

// serveFile writes the named regular file with status, or reports false when
// there is no such file.
func serveFile(w http.ResponseWriter, r *http.Request, files fs.FS, name string, status int) bool {
	if !fs.ValidPath(name) {
		return false
	}
	info, err := fs.Stat(files, name)
	if err != nil || !info.Mode().IsRegular() {
		return false
	}
	data, err := fs.ReadFile(files, name)
	if err != nil {
		return false
	}

	if strings.HasSuffix(name, ".html") {
		// Pages name the current bundle, so always check for a new one.
		w.Header().Set("Cache-Control", "no-cache")
	} else if strings.HasPrefix(name, "_expo/static/") {
		// Bundles are named by their content hash.
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	}
	if status == http.StatusOK {
		http.ServeContent(w, r, name, info.ModTime(), bytes.NewReader(data))
		return true
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if r.Method != http.MethodHead {
		_, _ = w.Write(data)
	}
	return true
}
