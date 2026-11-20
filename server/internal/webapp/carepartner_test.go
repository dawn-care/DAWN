package webapp

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func testBuild() fstest.MapFS {
	return fstest.MapFS{
		"index.html":                       {Data: []byte("index")},
		"pair.html":                        {Data: []byte("pair")},
		"+not-found.html":                  {Data: []byte("not found")},
		"_expo/static/js/web/entry-abc.js": {Data: []byte("bundle")},
		"assets/assets/images/icon.png":    {Data: []byte("png")},
	}
}

func get(t *testing.T, h http.Handler, method, target string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, target, nil))
	return rec
}

func TestCarepartnerServesPagesAndFiles(t *testing.T) {
	h := Carepartner(testBuild())
	cases := []struct {
		target string
		status int
		body   string
	}{
		{"/carepartner/?preview=1&carepartner_id=1", http.StatusOK, "index"},
		{"/carepartner/index.html", http.StatusOK, "index"},
		{"/carepartner/pair", http.StatusOK, "pair"},
		{"/carepartner/pair?x=1", http.StatusOK, "pair"},
		{"/carepartner/_expo/static/js/web/entry-abc.js", http.StatusOK, "bundle"},
		{"/carepartner/assets/assets/images/icon.png", http.StatusOK, "png"},
		{"/carepartner/nope", http.StatusNotFound, "not found"},
		{"/carepartner/assets", http.StatusNotFound, "not found"},
		{"/carepartner/../index.html", http.StatusOK, "index"},
	}
	for _, c := range cases {
		rec := get(t, h, http.MethodGet, c.target)
		if rec.Code != c.status || rec.Body.String() != c.body {
			t.Errorf("GET %s = %d %q, want %d %q", c.target, rec.Code, rec.Body.String(), c.status, c.body)
		}
	}
}

func TestCarepartnerRedirectsBarePathKeepingQuery(t *testing.T) {
	rec := get(t, Carepartner(testBuild()), http.MethodGet, "/carepartner?preview=1&carepartner_id=2")
	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("status = %d, want 301", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/carepartner/?preview=1&carepartner_id=2" {
		t.Errorf("Location = %q", loc)
	}
}

func TestCarepartnerAllowsOnlySameOriginFraming(t *testing.T) {
	for _, target := range []string{"/carepartner/", "/carepartner/nope", "/carepartner"} {
		rec := get(t, Carepartner(testBuild()), http.MethodGet, target)
		if got := rec.Header().Get("Content-Security-Policy"); got != "frame-ancestors 'self'" {
			t.Errorf("%s: Content-Security-Policy = %q", target, got)
		}
		if got := rec.Header().Get("X-Frame-Options"); got != "SAMEORIGIN" {
			t.Errorf("%s: X-Frame-Options = %q", target, got)
		}
	}
}

func TestCarepartnerCaching(t *testing.T) {
	h := Carepartner(testBuild())
	if got := get(t, h, http.MethodGet, "/carepartner/").Header().Get("Cache-Control"); got != "no-cache" {
		t.Errorf("page Cache-Control = %q, want no-cache", got)
	}
	if got := get(t, h, http.MethodGet, "/carepartner/_expo/static/js/web/entry-abc.js").Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Errorf("bundle Cache-Control = %q", got)
	}
}

func TestCarepartnerRejectsWrites(t *testing.T) {
	if rec := get(t, Carepartner(testBuild()), http.MethodPost, "/carepartner/"); rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST status = %d, want 405", rec.Code)
	}
}

func TestCarepartnerHeadHasNoBody(t *testing.T) {
	rec := get(t, Carepartner(testBuild()), http.MethodHead, "/carepartner/nope")
	if rec.Code != http.StatusNotFound || rec.Body.Len() != 0 {
		t.Errorf("HEAD = %d with %d bytes, want 404 and none", rec.Code, rec.Body.Len())
	}
}
