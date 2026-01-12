// The spa package is experimental. The API will most likely change
package spa

import (
	"io/fs"
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

type HandlerOpts struct {
	IndexFile        string
	IndexRoute       string
	IndexCacheHeader string
	AssetFS          fs.FS
	AssetFSPrefix    string
	AssetPrefixes    []string
	AssetCacheHeader string
	NotFoundPrefixes []string
}

type Handler struct {
	IndexFile        string
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

	return &Handler{
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
		filename = h.IndexFile
		mimeType = getMimeType(h.IndexFile)
		cacheHeader = h.IndexCacheHeader
	}

	h.serveFile(filename, mimeType, cacheHeader, w, r)
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
