// Copyright 2025 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package srv

// ErrorDto represents an error response with a code and message.
type ErrorDto struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}
