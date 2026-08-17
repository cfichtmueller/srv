// Copyright 2026 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package srv

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGroupPath_CollapsesDoubleSlashAtSeam(t *testing.T) {
	tests := []struct {
		name     string
		register func(s *Server)
		path     string
	}{
		{
			name: "trailing slash on base + leading slash on sub",
			register: func(s *Server) {
				s.Group("/api/").GET("/v1", okHandler)
			},
			path: "/api/v1",
		},
		{
			name: "documented convention still works",
			register: func(s *Server) {
				s.Group("/api").GET("/v1", okHandler)
			},
			path: "/api/v1",
		},
		{
			name: "empty sub path does not insert trailing slash",
			register: func(s *Server) {
				s.Group("/api").GET("", okHandler)
			},
			path: "/api",
		},
		{
			name: "nested groups collapse seams",
			register: func(s *Server) {
				s.Group("/api/").Group("/v1/").GET("/users", okHandler)
			},
			path: "/api/v1/users",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := NewServer()
			tc.register(s)
			ts := httptest.NewServer(s.Handler())
			defer ts.Close()

			res, err := http.Get(ts.URL + tc.path)
			if err != nil {
				t.Fatalf("GET %s: %v", tc.path, err)
			}
			res.Body.Close()
			if res.StatusCode != http.StatusOK {
				t.Errorf("GET %s: got %d, want 200", tc.path, res.StatusCode)
			}
		})
	}
}

func okHandler(c *Context) *Response {
	return Respond().Status(http.StatusOK).Text("ok")
}
