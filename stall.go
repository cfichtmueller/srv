// Copyright 2026 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package srv

import (
	"errors"
	"io"
	"net/http"
	"time"
)

// ReadStallTimeout returns middleware that aborts a request whose body stops
// making progress for longer than d. The read deadline is reset on every Read,
// so a slow-but-progressing client (for example a large upload over a thin link)
// keeps going, while a stalled or dead connection is cut off after d.
//
// It resets the deadline through Context.ResponseController. On connections that
// do not support read deadlines it degrades to a no-op. A non-positive d disables
// the timeout, leaving the body untouched.
func ReadStallTimeout(d time.Duration) Middleware {
	return func(c *Context, next Handler) *Response {
		if d > 0 && c.r.Body != nil {
			c.r.Body = &stallReader{
				r:  c.r.Body,
				rc: c.ResponseController(),
				d:  d,
			}
		}
		return next(c)
	}
}

// stallReader wraps a request body and refreshes the connection read deadline
// before each Read, so the read fails if no bytes arrive within d.
type stallReader struct {
	r  io.ReadCloser
	rc *http.ResponseController
	d  time.Duration
}

func (s *stallReader) Read(p []byte) (int, error) {
	if err := s.rc.SetReadDeadline(time.Now().Add(s.d)); err != nil && !errors.Is(err, http.ErrNotSupported) {
		return 0, err
	}
	return s.r.Read(p)
}

func (s *stallReader) Close() error {
	return s.r.Close()
}
