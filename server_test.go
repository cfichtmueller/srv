// Copyright 2026 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package srv

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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

func TestNewServer_DefaultTimeouts(t *testing.T) {
	hs := NewServer().HTTPServer()

	if hs.ReadHeaderTimeout != DefaultReadHeaderTimeout {
		t.Errorf("ReadHeaderTimeout = %v, want %v", hs.ReadHeaderTimeout, DefaultReadHeaderTimeout)
	}
	if hs.IdleTimeout != DefaultIdleTimeout {
		t.Errorf("IdleTimeout = %v, want %v", hs.IdleTimeout, DefaultIdleTimeout)
	}
	// ReadTimeout / WriteTimeout must stay unset: they would break streaming.
	if hs.ReadTimeout != 0 {
		t.Errorf("ReadTimeout = %v, want 0", hs.ReadTimeout)
	}
	if hs.WriteTimeout != 0 {
		t.Errorf("WriteTimeout = %v, want 0", hs.WriteTimeout)
	}
}

func TestNewServer_TimeoutSettersOverrideDefaults(t *testing.T) {
	s := NewServer().
		SetReadHeaderTimeout(3 * time.Second).
		SetIdleTimeout(0)
	hs := s.HTTPServer()

	if hs.ReadHeaderTimeout != 3*time.Second {
		t.Errorf("ReadHeaderTimeout = %v, want 3s", hs.ReadHeaderTimeout)
	}
	if hs.IdleTimeout != 0 {
		t.Errorf("IdleTimeout = %v, want 0 (disabled)", hs.IdleTimeout)
	}
}
