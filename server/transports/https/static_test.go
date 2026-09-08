package https

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func testDist() fstest.MapFS {
	js := strings.Repeat("console.log('metis');", 200) // compresses well
	return fstest.MapFS{
		"index.html":            {Data: []byte("<!doctype html><title>Metis</title>")},
		"assets/app-abc123.js":  {Data: []byte(js)},
		"assets/app-abc123.css": {Data: []byte(strings.Repeat(".a{color:red}", 200))},
		"favicon.svg":           {Data: []byte("<svg/>")},
		"manifest.webmanifest":  {Data: []byte(`{"name":"Metis BPM"}`)},
		"sw.js":                 {Data: []byte("self.addEventListener('install',()=>{})")},
	}
}

func TestHashedAssetsAreImmutableAndCompressed(t *testing.T) {
	h, err := newStaticHandler(testDist())
	if err != nil {
		t.Fatalf("build static handler: %v", err)
	}

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/assets/app-abc123.js", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Fatalf("hashed asset Cache-Control %q, want immutable", got)
	}
	if got := rec.Header().Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("Content-Encoding %q, want gzip", got)
	}
	if rec.Header().Get("ETag") == "" {
		t.Fatal("hashed asset has no ETag")
	}
	zr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("body is not gzip: %v", err)
	}
	body, _ := io.ReadAll(zr)
	if !strings.Contains(string(body), "console.log") {
		t.Fatalf("decompressed body is not the asset: %q", string(body)[:20])
	}
}

func TestIndexIsRevalidatedNotCachedForever(t *testing.T) {
	h, _ := newStaticHandler(testDist())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))
	if got := rec.Header().Get("Cache-Control"); got != "no-cache" {
		t.Fatalf("index Cache-Control %q, want no-cache", got)
	}
}

func TestUnknownRouteFallsBackToIndexButMissingAssetIs404(t *testing.T) {
	h, _ := newStaticHandler(testDist())

	// A client-side route: the SPA shell answers it.
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/tasks/inbox", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Metis") {
		t.Fatalf("SPA route did not return the shell: %d", rec.Code)
	}

	// A missing hashed asset is a real 404, not HTML with a JS content type.
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/assets/does-not-exist.js", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("missing asset status %d, want 404", rec.Code)
	}
}

func TestConditionalRequestGetsNotModified(t *testing.T) {
	h, _ := newStaticHandler(testDist())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/assets/app-abc123.css", nil))
	etag := rec.Header().Get("ETag")

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/assets/app-abc123.css", nil)
	req.Header.Set("If-None-Match", etag)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotModified {
		t.Fatalf("conditional request status %d, want 304", rec.Code)
	}
}

func TestSecurityHeadersArePresent(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	rec := httptest.NewRecorder()
	securityHeaders(next).ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))

	for _, h := range []string{
		"Content-Security-Policy",
		"X-Content-Type-Options",
		"X-Frame-Options",
		"Referrer-Policy",
	} {
		if rec.Header().Get(h) == "" {
			t.Fatalf("missing security header %s", h)
		}
	}
	if !strings.Contains(rec.Header().Get("Content-Security-Policy"), "frame-ancestors 'none'") {
		t.Fatal("CSP does not deny framing")
	}
}

/*
 * The two files that make the app installable.
 *
 * A manifest served as text/plain is ignored by the browser and the app is
 * simply not installable, with nothing in the console saying why. And a service
 * worker must never be cached the way a hashed bundle is: sw.js keeps its name
 * across deploys, so an immutable copy would pin the browser to the old one
 * forever.
 */
func TestManifestAndServiceWorkerAreServedCorrectly(t *testing.T) {
	h, _ := newStaticHandler(testDist())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/manifest.webmanifest", nil))
	if got := rec.Header().Get("Content-Type"); got != "application/manifest+json" {
		t.Fatalf("manifest Content-Type %q", got)
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/sw.js", nil))
	if got := rec.Header().Get("Cache-Control"); got != "no-cache" {
		t.Fatalf("service worker Cache-Control %q, want no-cache", got)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/javascript") {
		t.Fatalf("service worker Content-Type %q", ct)
	}
}
