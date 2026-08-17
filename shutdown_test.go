// Copyright 2026 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package srv

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestShutdown_ReturnsErrServerClosed(t *testing.T) {
	s := NewServer()
	s.GET("/healthz", okHandler)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- s.httpServer.Serve(ln)
	}()

	waitForServer(t, addr)

	resp, err := http.Get("http://" + addr + "/healthz")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	if err := <-serveErr; !errors.Is(err, http.ErrServerClosed) {
		t.Fatalf("expected ErrServerClosed, got %v", err)
	}
}

func TestShutdown_DrainsInFlightRequest(t *testing.T) {
	s := NewServer()
	started := make(chan struct{})
	s.GET("/slow", func(c *Context) *Response {
		close(started)
		time.Sleep(300 * time.Millisecond)
		return Respond().Json(map[string]any{"status": "done"})
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()

	go func() { _ = s.httpServer.Serve(ln) }()
	waitForServer(t, addr)

	respErr := make(chan error, 1)
	body := make(chan string, 1)
	go func() {
		resp, err := http.Get("http://" + addr + "/slow")
		if err != nil {
			respErr <- err
			return
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			respErr <- errors.New("unexpected status: " + resp.Status)
			return
		}
		body <- string(b)
	}()

	<-started // request is being handled; now shut down

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	select {
	case err := <-respErr:
		t.Fatalf("in-flight request failed: %v", err)
	case b := <-body:
		if b == "" {
			t.Fatal("expected non-empty body from drained request")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("in-flight request did not complete")
	}
}

func TestShutdown_BeforeServeIsSafe(t *testing.T) {
	s := NewServer()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestTimeoutSetters(t *testing.T) {
	s := NewServer()
	got := s.
		SetReadHeaderTimeout(1 * time.Second).
		SetReadTimeout(2 * time.Second).
		SetWriteTimeout(3 * time.Second).
		SetIdleTimeout(4 * time.Second)

	if got != s {
		t.Fatal("setters must return the same *Server for chaining")
	}
	hs := s.HTTPServer()
	if hs.ReadHeaderTimeout != 1*time.Second {
		t.Errorf("ReadHeaderTimeout = %v", hs.ReadHeaderTimeout)
	}
	if hs.ReadTimeout != 2*time.Second {
		t.Errorf("ReadTimeout = %v", hs.ReadTimeout)
	}
	if hs.WriteTimeout != 3*time.Second {
		t.Errorf("WriteTimeout = %v", hs.WriteTimeout)
	}
	if hs.IdleTimeout != 4*time.Second {
		t.Errorf("IdleTimeout = %v", hs.IdleTimeout)
	}
}

func waitForServer(t *testing.T, addr string) {
	t.Helper()
	for i := 0; i < 50; i++ {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("server at %s did not become ready", addr)
}
