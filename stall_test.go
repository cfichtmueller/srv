// Copyright 2026 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package srv

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestReadStallTimeout_AbortsStalledBody(t *testing.T) {
	readErrCh := make(chan error, 1)
	s := NewServer()
	s.POST("/upload", func(c *Context) *Response {
		_, err := io.ReadAll(c.Request().Body)
		readErrCh <- err
		return Respond().Text("done")
	}, ReadStallTimeout(100*time.Millisecond))

	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	pr, pw := io.Pipe()
	defer pw.Close()
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/upload", pr)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	// Promise more than we send, then stall without sending the rest.
	req.ContentLength = 1024

	go func() {
		_, _ = pw.Write([]byte("partial"))
		// Never write the remaining bytes and never close: the read stalls.
	}()
	go func() {
		if resp, err := http.DefaultClient.Do(req); err == nil {
			resp.Body.Close()
		}
	}()

	select {
	case err := <-readErrCh:
		if err == nil {
			t.Fatal("expected a read error from the stalled body, got nil")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("handler never observed a read error; stall timeout did not fire")
	}
}

func TestReadStallTimeout_PassesThroughHealthyBody(t *testing.T) {
	var got string
	s := NewServer()
	s.POST("/upload", func(c *Context) *Response {
		b, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return Respond().InternalServerError()
		}
		got = string(b)
		return Respond().Text("ok")
	}, ReadStallTimeout(2*time.Second))

	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/upload", "text/plain", strings.NewReader("hello world"))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if got != "hello world" {
		t.Errorf("body = %q, want %q", got, "hello world")
	}
}

func TestReadStallTimeout_NonPositiveDisabled(t *testing.T) {
	s := NewServer()
	var wrapped bool
	s.POST("/upload", func(c *Context) *Response {
		_, wrapped = c.Request().Body.(*stallReader)
		io.Copy(io.Discard, c.Request().Body)
		return Respond().Text("ok")
	}, ReadStallTimeout(0))

	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/upload", "text/plain", strings.NewReader("x"))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	resp.Body.Close()

	if wrapped {
		t.Error("body was wrapped in stallReader despite non-positive duration")
	}
}
