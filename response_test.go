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

func TestWithWriteStallTimeout_SetsFieldAndChains(t *testing.T) {
	r := Respond()
	if got := r.WithWriteStallTimeout(5 * time.Second); got != r {
		t.Error("WithWriteStallTimeout did not return the same *Response")
	}
	if r.writeStallTimeout != 5*time.Second {
		t.Errorf("writeStallTimeout = %v, want 5s", r.writeStallTimeout)
	}
}

func TestWithWriteStallTimeout_HealthyDownloadCompletes(t *testing.T) {
	payload := strings.Repeat("srv", 1000)
	s := NewServer()
	s.GET("/download", func(c *Context) *Response {
		return Respond().
			BodyReader("text/plain", strings.NewReader(payload)).
			WithWriteStallTimeout(2 * time.Second)
	})

	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/download")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if string(b) != payload {
		t.Errorf("body = %d bytes, want %d", len(b), len(payload))
	}
}

func TestWithWriteStallTimeout_AbortsStalledDownload(t *testing.T) {
	writeErrCh := make(chan error, 1)
	chunk := make([]byte, 1<<20) // 1 MiB
	s := NewServer()
	s.GET("/download", func(c *Context) *Response {
		return Respond().
			BodyFn("application/octet-stream", func(w io.Writer) error {
				var err error
				for i := 0; i < 256 && err == nil; i++ { // up to 256 MiB
					_, err = w.Write(chunk)
				}
				writeErrCh <- err
				return err
			}).
			WithWriteStallTimeout(100 * time.Millisecond)
	})

	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/download")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	// Read a little, then stop reading so the server's writes fill the socket
	// buffers and stall.
	buf := make([]byte, 4096)
	_, _ = resp.Body.Read(buf)

	select {
	case err := <-writeErrCh:
		if err == nil {
			t.Fatal("expected a write error from the stalled download, got nil")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("body writer never observed a stall error; write stall timeout did not fire")
	}
}
