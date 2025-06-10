// Copyright 2025 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package srv

import "time"

func maxTime(t []time.Time) time.Time {
	mt := time.Time{}
	for _, ct := range t {
		if ct.After(mt) {
			mt = ct
		}
	}
	return mt
}
