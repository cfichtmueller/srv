// Copyright 2025 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package srv

// Validatable represents an object that can be validated.
type Validatable interface {
	// Validate validates the object and returns an error if the object is invalid.
	Validate() error
}
