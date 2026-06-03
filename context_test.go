// Copyright 2026 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package srv

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
