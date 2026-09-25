// Copyright 2026 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package srv

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestContext(target string) *Context {
	r := httptest.NewRequest(http.MethodGet, target, nil)
	w := httptest.NewRecorder()
	return NewContext(w, r, &contextConfig{
		maxMultipartMemory: DefaultMaxMultipartMemory,
		ipResolver:         NewIPResolver(nil, false),
	})
}

func TestStringQueryOrDefault_PreservesPlusFromPercentEncoding(t *testing.T) {
	c := newTestContext("/?x=hello%2Bworld")
	got, res := c.StringQueryOrDefault("x", "")
	if res != nil {
		t.Fatalf("unexpected response: %+v", res)
	}
	if got != "hello+world" {
		t.Errorf("expected %q, got %q", "hello+world", got)
	}
}

func TestStringQueryOrDefault_DecodesPlusAsSpace(t *testing.T) {
	c := newTestContext("/?x=hello+world")
	got, res := c.StringQueryOrDefault("x", "")
	if res != nil {
		t.Fatalf("unexpected response: %+v", res)
	}
	if got != "hello world" {
		t.Errorf("expected %q, got %q", "hello world", got)
	}
}

func TestStringQueryOrDefault_MissingKeyReturnsDefault(t *testing.T) {
	c := newTestContext("/")
	got, res := c.StringQueryOrDefault("x", "fallback")
	if res != nil {
		t.Fatalf("unexpected response: %+v", res)
	}
	if got != "fallback" {
		t.Errorf("expected %q, got %q", "fallback", got)
	}
}

func TestStringQueryOrDefault_EmptyValueReturnsDefault(t *testing.T) {
	c := newTestContext("/?x=")
	got, res := c.StringQueryOrDefault("x", "fallback")
	if res != nil {
		t.Fatalf("unexpected response: %+v", res)
	}
	if got != "fallback" {
		t.Errorf("expected %q, got %q", "fallback", got)
	}
}

func TestResponseController_NotNilAndWired(t *testing.T) {
	c := newTestContext("/")
	rc := c.ResponseController()
	if rc == nil {
		t.Fatal("ResponseController() returned nil")
	}
	// The recorder implements http.Flusher, so Flush must reach it.
	if err := rc.Flush(); err != nil {
		t.Errorf("Flush() error = %v", err)
	}
	// The recorder does not support deadlines; the controller must be wired to
	// it and surface http.ErrNotSupported rather than panicking.
	if err := rc.SetReadDeadline(time.Now()); !errors.Is(err, http.ErrNotSupported) {
		t.Errorf("SetReadDeadline() error = %v, want ErrNotSupported", err)
	}
}

func TestResponseController_SetReadDeadlineOnRealConn(t *testing.T) {
	s := NewServer()
	var deadlineErr error
	s.GET("/", func(c *Context) *Response {
		deadlineErr = c.ResponseController().SetReadDeadline(time.Now().Add(time.Minute))
		return Respond().Text("ok")
	})

	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	res, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	res.Body.Close()

	if deadlineErr != nil {
		t.Errorf("SetReadDeadline() on a real connection error = %v, want nil", deadlineErr)
	}
}
