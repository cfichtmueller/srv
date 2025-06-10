// Copyright 2025 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package srv

import (
	"log/slog"
	"net/http"
)

const (
	DefaultMaxMultipartMemory = 64 << 20
)

// Server represents an HTTP server that can handle requests and responses.
type Server struct {
	MaxMultipartMemory int64
	middleware         []Middleware
	mux                *http.ServeMux
	contextConfig      *contextConfig
}

// NewServer creates a new Server with a new ServeMux.
func NewServer() *Server {
	return &Server{
		middleware: make([]Middleware, 0),
		mux:        http.NewServeMux(),
		contextConfig: &contextConfig{
			maxMultipartMemory: DefaultMaxMultipartMemory,
			ipResolver: NewIPResolver([]string{
				"X-Forwarded-For",
				"Forwarded",
			}, false),
		},
	}
}

func (s *Server) SetMaxMultipartMemory(max int64) *Server {
	s.contextConfig.maxMultipartMemory = max
	return s
}

func (s *Server) SetRemoteIPHeaders(headers ...string) *Server {
	s.contextConfig.ipResolver.RemoteIPHeaders = headers
	return s
}

func (s *Server) SetTrustRemoteIdHeaders(trust bool) *Server {
	s.contextConfig.ipResolver.TrustRemoteIdHeaders = trust
	return s
}

// Group creates a new Group with the given path.
func (s *Server) Group(path string, middleware ...Middleware) *Group {
	return &Group{
		basePath:      path,
		mux:           s.mux,
		middleware:    append(s.middleware[:], middleware...),
		contextConfig: s.contextConfig,
	}
}

// Use adds middleware to the Server.
func (s *Server) Use(middleware ...Middleware) *Server {
	s.middleware = append(s.middleware, middleware...)
	return s
}

// OPTIONS adds a new route for the OPTIONS method with the given path, handler, and middleware.
func (s *Server) OPTIONS(path string, handler Handler, middleware ...Middleware) {
	s.handleMethod("OPTIONS", path, handler, middleware)
}

// HEAD adds a new route for the HEAD method with the given path, handler, and middleware.
func (s *Server) HEAD(path string, handler Handler, middleware ...Middleware) {
	s.handleMethod("HEAD", path, handler, middleware)
}

// GET adds a new route for the GET method with the given path, handler, and middleware.
func (s *Server) GET(path string, handler Handler, middleware ...Middleware) {
	s.handleMethod("GET", path, handler, middleware)
}

// POST adds a new route for the POST method with the given path, handler, and middleware.
func (s *Server) POST(path string, handler Handler, middleware ...Middleware) {
	s.handleMethod("POST", path, handler, middleware)
}

// PUT adds a new route for the PUT method with the given path, handler, and middleware.
func (s *Server) PUT(path string, handler Handler, middleware ...Middleware) {
	s.handleMethod("PUT", path, handler, middleware)
}

// DELETE adds a new route for the DELETE method with the given path, handler, and middleware.
func (s *Server) DELETE(path string, handler Handler, middleware ...Middleware) {
	s.handleMethod("DELETE", path, handler, middleware)
}

// handleMethod adds a new route for the given method, path, handler, and middleware.
func (s *Server) handleMethod(method, path string, handler Handler, middleware []Middleware) {
	if path == "" {
		path = "/"
	}
	pattern := method + " " + path
	s.mux.HandleFunc(pattern, wrap(s.contextConfig, append(s.middleware, middleware...), handler))
}

// ListenAndServe starts the server and listens for incoming requests on the given address.
func (s *Server) ListenAndServe(address string) error {
	return http.ListenAndServe(address, s.mux)
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

type Group struct {
	basePath      string
	middleware    []Middleware
	mux           *http.ServeMux
	contextConfig *contextConfig
}

// Group creates a new Group with the given path.
func (g *Group) Group(path string, middleware ...Middleware) *Group {
	return &Group{
		middleware:    append(g.middleware[:], middleware...),
		basePath:      g.basePath + path,
		mux:           g.mux,
		contextConfig: g.contextConfig,
	}
}

// OPTIONS adds a new route for the OPTIONS method with the given path, handler, and middleware.
func (g *Group) OPTIONS(path string, handler Handler, middleware ...Middleware) {
	g.handleMethod("OPTIONS", path, handler, middleware)
}

// HEAD adds a new route for the HEAD method with the given path, handler, and middleware.
func (g *Group) HEAD(path string, handler Handler, middleware ...Middleware) {
	g.handleMethod("HEAD", path, handler, middleware)
}

// GET adds a new route for the GET method with the given path, handler, and middleware.
func (g *Group) GET(path string, handler Handler, middleware ...Middleware) {
	g.handleMethod("GET", path, handler, middleware)
}

// POST adds a new route for the POST method with the given path, handler, and middleware.
func (g *Group) POST(path string, handler Handler, middleware ...Middleware) {
	g.handleMethod("POST", path, handler, middleware)
}

// PUT adds a new route for the PUT method with the given path, handler, and middleware.
func (g *Group) PUT(path string, handler Handler, middleware ...Middleware) {
	g.handleMethod("PUT", path, handler, middleware)
}

// DELETE adds a new route for the DELETE method with the given path, handler, and middleware.
func (g *Group) DELETE(path string, handler Handler, middleware ...Middleware) {
	g.handleMethod("DELETE", path, handler, middleware)
}

// handleMethod adds a new route for the given method, path, handler, and middleware.
func (g *Group) handleMethod(method, path string, handler Handler, middleware []Middleware) {
	g.mux.HandleFunc(method+" "+g.basePath+path, wrap(g.contextConfig, append(g.middleware, middleware...), handler))
}

func wrap(conf *contextConfig, middleware []Middleware, handler Handler) func(http.ResponseWriter, *http.Request) {
	h := handler
	if len(middleware) > 0 {
		h = wrapMiddleware(middleware, handler)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		res := h(NewContext(w, r, conf))
		if res == nil {
			panic("received nil response from handler")
		}
		if err := res.Write(w); err != nil {
			slog.Error("unable to write response", "error", err.Error())
		}
	}
}

func wrapMiddleware(middleware []Middleware, handler Handler) Handler {
	if len(middleware) == 1 {
		return func(c *Context) *Response {
			return middleware[0](c, handler)
		}
	}
	remaining := wrapMiddleware(middleware[1:], handler)
	return func(c *Context) *Response {
		return middleware[0](c, remaining)
	}
}
