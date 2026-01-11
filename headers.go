// Copyright 2026 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package srv

// ContentDispositionAttachment returns the value of the Content-Disposition header for an attachment.
func ContentDispositionAttachment(filename string) string {
	return "attachment; filename=\"" + filename + "\""
}

// ContentDispositionInline returns the value of the Content-Disposition header for an inline file.
func ContentDispositionInline(filename string) string {
	return "inline; filename=\"" + filename + "\""
}
