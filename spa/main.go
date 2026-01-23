// The spa package is experimental. The API will most likely change
package spa

import (
	"bytes"
	"html/template"
	"io/fs"
	"log"
	"log/slog"
	"mime"
	"net/http"
	"path"
	"strings"
)

// TODO List:
//	- cache files (optional)
//  - support templating index files
//	- support multiple index files
//	- support index aliases (e.g. /index.html)
//	- custom error response generators
//	- custom 404 pages
//	- test all possible combinations of leading and trailing slashes
//	- 405 when POST, PUT, DELETE index routes
//	- 405 when POST, PUT, DELETE resource routes
//  - handle resource not found

type IndexDataFn func(r *http.Request) (any, error)

type HandlerOpts struct {
	IndexFile           string
	IndexFileIsTemplate bool
	IndexDataFn         IndexDataFn
	IndexRoute          string
	IndexCacheHeader    string
	AssetFS             fs.FS
	AssetFSPrefix       string
	AssetPrefixes       []string
	AssetCacheHeader    string
	NotFoundPrefixes    []string
}

type Handler struct {
	IndexFile        string
	IndexTemplate    *template.Template
	IndexDataFn      IndexDataFn
	IndexRoute       string
	IndexCacheHeader string
	AssetFS          fs.FS
	AssetFSPrefix    string
	AssetPrefixes    []string
	AssetCacheHeader string
	reader           Reader
	NotFoundPrefixes []string
}

func NewHandler(opts *HandlerOpts) *Handler {
	if opts.AssetFS == nil {
		panic("opts.AssetFS is required")
	}
	if opts.IndexFileIsTemplate && opts.IndexDataFn == nil {
		panic("opts.IndexDataFn is required when opts.IndexFileIsTemplate is set")
	}

	indexFile := opts.IndexFile
	if indexFile == "" {
		indexFile = "index.html"
	}
	indexRoute := opts.IndexRoute
	if indexRoute == "" {
		indexRoute = "/"
	}

	var reader Reader = &FsReader{
		fs: opts.AssetFS,
	}

	h := &Handler{
		IndexFile:        indexFile,
		IndexRoute:       indexRoute,
		IndexCacheHeader: opts.IndexCacheHeader,
		AssetFS:          opts.AssetFS,
		AssetFSPrefix:    opts.AssetFSPrefix,
		AssetPrefixes:    opts.AssetPrefixes,
		AssetCacheHeader: opts.AssetCacheHeader,
		reader:           reader,
		NotFoundPrefixes: opts.NotFoundPrefixes,
	}

	if opts.IndexFileIsTemplate {
		b, ok, err := reader.Read(h.AssetFSPrefix + indexFile)
		if err != nil {
			log.Fatalf("unable to read index file '%s': %v", indexFile, err)
		}
		if !ok {
			log.Fatalf("unable to find index file '%s'", indexFile)
		}
		t, err := template.New("index").Parse(string(b))
		if err != nil {
			log.Fatalf("unable to parse index file template '%s': %v", indexFile, err)
		}
		h.IndexTemplate = t
		h.IndexDataFn = opts.IndexDataFn
	}

	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	for _, p := range h.NotFoundPrefixes {
		if strings.HasPrefix(path, p) {
			respond404(w, r)
			return
		}
	}

	if r.Method != "GET" && r.Method != "HEAD" && r.Method != "OPTIONS" {
		respond405(w, r)
		return
	}

	var filename string
	var mimeType string
	var cacheHeader string

	isAssetRoute := false
	for _, prefix := range h.AssetPrefixes {
		if strings.HasPrefix(path, prefix) {
			filename = strings.TrimPrefix(path, "/")
			mimeType = getMimeType(filename)
			cacheHeader = h.AssetCacheHeader
			isAssetRoute = true
			break
		}
	}

	// TODO: don't serve index file for routes with file ending (e.g. .png, .ico etc)

	if !isAssetRoute {
		h.serveIndex(w, r)
		return
	}

	h.serveFile(filename, mimeType, cacheHeader, w, r)
}

func (h *Handler) serveIndex(w http.ResponseWriter, r *http.Request) {
	filename := h.IndexFile
	mimeType := getMimeType(h.IndexFile)
	cacheHeader := h.IndexCacheHeader

	if h.IndexTemplate == nil {
		h.serveFile(filename, mimeType, cacheHeader, w, r)
		return
	}

	data, err := h.IndexDataFn(r)
	if err != nil {
		slog.Error("Unable to get index template data", "error", err)
		h.write500(w)
		return
	}
	out := &bytes.Buffer{}
	if err := h.IndexTemplate.Execute(out, data); err != nil {
		slog.Error("Unable to render index template", "error", err)
		h.write500(w)
		return
	}
	w.Header().Add("Content-Type", mimeType)
	w.WriteHeader(200)
	w.Write(out.Bytes())
}

func (h *Handler) write500(w http.ResponseWriter) {
	w.WriteHeader(500)
	w.Write([]byte("Internal Error"))
}

func (h *Handler) serveFile(name, contentType, cacheHeader string, w http.ResponseWriter, r *http.Request) {
	b, ok, err := h.reader.Read(h.AssetFSPrefix + name)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	if !ok {
		respond404(w, r)
		return
	}
	w.Header().Add("Content-Type", contentType)
	if cacheHeader != "" {
		w.Header().Add("Cache-Control", cacheHeader)
	}
	w.WriteHeader(200)
	w.Write(b)
}

func respond404(w http.ResponseWriter, r *http.Request) {
	contentType := "text/plain"
	body := []byte("Not Found")
	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		contentType = "application/json"
		body = []byte(`{"error":"not found"}`)
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(404)
	w.Write(body)
}

func respond405(w http.ResponseWriter, r *http.Request) {
	contentType := "text/plain"
	body := []byte("Method Not Allowed")
	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		contentType = "application/json"
		body = []byte(`{"error":"method not allowed"}`)
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(405)
	w.Write(body)
}

func getMimeType(filename string) string {
	ext := path.Ext(filename)
	t := mime.TypeByExtension(ext)
	if t == "" {
		t = "application/octet-stream"
	}
	return t
}
