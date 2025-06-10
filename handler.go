// Copyright 2025 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package srv

// Handler represents a function that handles an HTTP request and returns a Response.
type Handler func(c *Context) *Response
