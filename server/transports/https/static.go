package https

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

// staticHandler serves the embedded UI: the built assets and, for any path the
// single-page app owns, index.html.
//
// It exists because the previous handler was a bare http.FileServerFS, which
// sent every asset uncompressed and with no Cache-Control. A first paint was
// ~1.25 MB on the wire when the same bytes gzip to ~330 kB, and every reload
// refetched the whole app because the browser was never told it could keep
// anything. The distroless image has no reverse proxy to add either, so this is
// where it has to happen.
//
// The dist is embedded and immutable for the life of the process, so every file
// is read once at construction, gzipped once, and its ETag computed once. Serving
// is then a map lookup and a write — no per-request compression, no per-request
// stat.
type staticHandler struct {
	assets map[string]*staticAsset
	index  *staticAsset
}

type staticAsset struct {
	contentType string
	etag        string
	raw         []byte
	gzipped     []byte // nil when compressing did not pay off
	// immutable is true for content-hashed assets under /assets, which may be
	// cached forever; index.html and the rest must be revalidated so a deploy
	// is seen.
	immutable bool
}

func newStaticHandler(distFS fs.FS) (*staticHandler, error) {
	h := &staticHandler{assets: map[string]*staticAsset{}}
	err := fs.WalkDir(distFS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		raw, err := fs.ReadFile(distFS, p)
		if err != nil {
			return err
		}
		a := buildAsset(p, raw)
		h.assets["/"+p] = a
		if p == "index.html" {
			h.index = a
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return h, nil
}

// contentTypeOverrides are the extensions the platform's own MIME table gets
// wrong or does not know. A web app manifest served as text/plain is ignored by
// the browser, and the app is then simply not installable — with nothing in the
// console to say why.
var contentTypeOverrides = map[string]string{
	".webmanifest": "application/manifest+json",
	".json":        "application/json",
	".js":          "text/javascript; charset=utf-8",
	".mjs":         "text/javascript; charset=utf-8",
	".css":         "text/css; charset=utf-8",
	".svg":         "image/svg+xml",
	".wasm":        "application/wasm",
}

func buildAsset(p string, raw []byte) *staticAsset {
	ext := path.Ext(p)
	ct, ok := contentTypeOverrides[ext]
	if !ok {
		ct = mime.TypeByExtension(ext)
	}
	if ct == "" {
		ct = http.DetectContentType(raw)
	}
	sum := sha256.Sum256(raw)
	a := &staticAsset{
		contentType: ct,
		etag:        `"` + hex.EncodeToString(sum[:16]) + `"`,
		raw:         raw,
		// Vite writes content-hashed names under assets/, so the bytes behind a
		// name never change: cache them forever. Everything else is revalidated.
		immutable: strings.HasPrefix(p, "assets/"),
	}
	if gz := gzipBytes(raw); len(gz) > 0 && len(gz) < len(raw)-64 && compressibleType(ct) {
		a.gzipped = gz
	}
	return a
}

func gzipBytes(raw []byte) []byte {
	var buf bytes.Buffer
	zw, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return nil
	}
	if _, err := zw.Write(raw); err != nil {
		return nil
	}
	if err := zw.Close(); err != nil {
		return nil
	}
	return buf.Bytes()
}

func compressibleType(ct string) bool {
	ct = strings.ToLower(ct)
	switch {
	case strings.HasPrefix(ct, "text/"),
		strings.Contains(ct, "javascript"),
		strings.Contains(ct, "json"),
		strings.Contains(ct, "svg"),
		strings.Contains(ct, "xml"),
		strings.Contains(ct, "wasm"):
		return true
	}
	return false
}

func (h *staticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	asset := h.resolve(r.URL.Path)
	if asset == nil {
		http.NotFound(w, r)
		return
	}

	header := w.Header()
	header.Set("Content-Type", asset.contentType)
	header.Set("ETag", asset.etag)
	header.Set("X-Content-Type-Options", "nosniff")
	if asset.immutable {
		header.Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		// index.html and friends carry the references to the hashed assets, so
		// a stale one pins the browser to an old deploy. Always revalidate.
		header.Set("Cache-Control", "no-cache")
	}

	if match := r.Header.Get("If-None-Match"); match != "" && etagMatches(match, asset.etag) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	body := asset.raw
	if asset.gzipped != nil && acceptsGzip(r) {
		header.Set("Content-Encoding", "gzip")
		header.Add("Vary", "Accept-Encoding")
		body = asset.gzipped
	}

	// Written directly rather than through http.ServeContent: a Range request
	// over the gzipped bytes would otherwise be answered with byte offsets into
	// the compressed stream, which no client can reassemble. The assets are
	// small and fully in memory, so range support buys nothing here.
	header.Set("Content-Length", strconv.Itoa(len(body)))
	if r.Method == http.MethodHead {
		w.WriteHeader(http.StatusOK)
		return
	}
	if _, err := w.Write(body); err != nil {
		// A failed write here is the client having gone away mid-response, which
		// is normal and not actionable — the status line is already sent, so
		// there is nothing left to tell them. Logged at debug rather than
		// discarded so a genuine pattern of truncated assets is still findable.
		log.Debug().Err(err).Str("path", r.URL.Path).Msg("Static asset write did not complete")
	}
}

// resolve maps a request path to an asset, falling back to index.html for the
// paths the single-page router owns. A missing file under /assets is a genuine
// 404 — it is a hashed name that should exist — rather than the app shell, so a
// broken asset reference does not silently return HTML with a JS content type.
func (h *staticHandler) resolve(urlPath string) *staticAsset {
	if urlPath == "/" || urlPath == "" {
		return h.index
	}
	if a, ok := h.assets[urlPath]; ok {
		return a
	}
	if strings.HasPrefix(urlPath, "/assets/") {
		return nil
	}
	return h.index
}

func acceptsGzip(r *http.Request) bool {
	for _, part := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
		if strings.EqualFold(strings.TrimSpace(strings.SplitN(part, ";", 2)[0]), "gzip") {
			return true
		}
	}
	return false
}

func etagMatches(header, etag string) bool {
	for _, candidate := range strings.Split(header, ",") {
		candidate = strings.TrimSpace(candidate)
		candidate = strings.TrimPrefix(candidate, "W/")
		if candidate == etag || candidate == "*" {
			return true
		}
	}
	return false
}

// securityHeaders wraps a handler with the response headers an enterprise buyer
// expects to find, and the CSP that makes the bearer token in localStorage
// harder to exfiltrate if a script ever slips in.
//
// The policy is written for exactly what the UI loads: its own bundle, Mantine's
// inline styles, Google Fonts, same-origin API and SSE, and a service worker.
// Nothing cross-origin is fetched beyond fonts, so the rest is 'self'.
func securityHeaders(next http.Handler) http.Handler {
	const csp = "default-src 'self'; " +
		"script-src 'self'; " +
		"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; " +
		"font-src 'self' https://fonts.gstatic.com data:; " +
		"img-src 'self' data: blob:; " +
		"connect-src 'self'; " +
		"worker-src 'self' blob:; " +
		"manifest-src 'self'; " +
		"object-src 'none'; " +
		"base-uri 'self'; " +
		"form-action 'self'; " +
		"frame-ancestors 'none'"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", csp)
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		// HSTS is only meaningful over TLS, which this server does not terminate
		// itself; the ingress adds it. Set it anyway so a direct HTTPS terminator
		// in front inherits a sane default. Harmless over plain HTTP: browsers
		// ignore it there.
		h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		next.ServeHTTP(w, r)
	})
}
