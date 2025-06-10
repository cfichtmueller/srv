// Copyright 2025 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package srv

import (
	"log/slog"
	"time"
)

// LoggingMiddleware logs the request and response status.
func LoggingMiddleware() Middleware {
	return func(c *Context, next Handler) *Response {
		start := time.Now()
		r := next(c)

		return r.AfterWrite(func() {
			slog.Info("request",
				"ip", c.ClientIP(),
				"method", c.r.Method,
				"path", c.r.URL.Path,
				"status", r.StatusCode,
				"duration", time.Since(start).Milliseconds(),
			)
		})
	}
}
