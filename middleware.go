// Copyright 2025 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package srv

// Middleware represents a function that processes an HTTP request and returns a Response.
type Middleware func(c *Context, next Handler) *Response
