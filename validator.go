// Copyright 2025 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package srv

import (
	"fmt"
	"regexp"
	"slices"
)

const (
	ValidationCodeTooFewItems  = "too_few_items"
	ValidationCodeTooManyItems = "too_many_items"
	ValidationCodeRequired     = "required"
	ValidationCodeTooShort     = "too_short"
	ValidationCodeTooLong      = "too_long"
	ValidationCodeInvalid      = "invalid"
)

// Validatable represents an object that can be validated.
type Validatable interface {
	// Validate validates the object and returns an error if the object is invalid.
	Validate() error
}

type ValidationError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Errors  []Violation `json:"errors"`
}

func (e *ValidationError) Error() string {
	return "validation error"
}

type Violation struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Require validates a condition and returns a ValidationError if the condition is false.
// If the condition is true, it returns the previous ValidationError unchanged.
// This allows for chaining multiple validation checks together.
func Require(field, code, message string, cond bool, prev *ValidationError) *ValidationError {
	if cond {
		return prev
	}
	return merge(prev, Violation{
		Field:   field,
		Code:    code,
		Message: message,
	})
}

// RequireIndexed validates a condition and returns a ValidationError if the condition is false.
// If the condition is true, it returns the previous ValidationError unchanged.
// This allows for chaining multiple validation checks together.
// The field name is formatted using the fieldFormat string and the index.
// The message is formatted using the messageFormat string and the index.
func RequireIndexed(fieldFormat string, index int, code string, messageFormat string, cond bool, prev *ValidationError) *ValidationError {
	if cond {
		return prev
	}
	return merge(prev, Violation{
		Field:   fmt.Sprintf(fieldFormat, index),
		Code:    code,
		Message: fmt.Sprintf(messageFormat, index),
	})
}

// RequireNotEmpty validates that a string value is not empty.
// It returns a ValidationError with ValidationCodeRequired if the value is empty.
// If the value is not empty, it returns the previous ValidationError unchanged.
func RequireNotEmpty(field string, value string, prev *ValidationError) *ValidationError {
	if value != "" {
		return prev
	}
	return merge(prev, Violation{
		Field:   field,
		Code:    ValidationCodeRequired,
		Message: field + " is required",
	})
}

// RequireNotEmptyIndexed validates that a string value is not empty.
// It returns a ValidationError with ValidationCodeRequired if the value is empty.
// If the value is not empty, it returns the previous ValidationError unchanged.
// The field name is formatted using the fieldFormat string and the index.
// The message is formatted using the fieldFormat string and the index.
func RequireNotEmptyIndexed(fieldFormat string, index int, value string, prev *ValidationError) *ValidationError {
	if value != "" {
		return prev
	}
	f := fmt.Sprintf(fieldFormat, index)
	return merge(prev, Violation{
		Field:   f,
		Code:    ValidationCodeRequired,
		Message: f + " is required",
	})
}

// RequireMinLength validates that a string value has at least the specified minimum length.
// It returns a ValidationError with ValidationCodeTooShort if the value is shorter than min.
// If the value meets the minimum length, it returns the previous ValidationError unchanged.
func RequireMinLength(field string, min int, value string, prev *ValidationError) *ValidationError {
	if min < 0 {
		panic("min must be greater than or equal to 0")
	}
	if len(value) >= min {
		return prev
	}
	return merge(prev, Violation{
		Field:   field,
		Code:    ValidationCodeTooShort,
		Message: "Value for " + field + " is too short",
	})
}

// RequireMinLengthIndexed validates that a string value has at least the specified minimum length.
// It returns a ValidationError with ValidationCodeTooShort if the value is shorter than min.
// If the value meets the minimum length, it returns the previous ValidationError unchanged.
// The field name is formatted using the fieldFormat string and the index.
// The message is formatted using the fieldFormat string and the index.
func RequireMinLengthIndexed(fieldFormat string, index int, min int, value string, prev *ValidationError) *ValidationError {
	if min < 0 {
		panic("min must be greater than or equal to 0")
	}
	if len(value) >= min {
		return prev
	}
	f := fmt.Sprintf(fieldFormat, index)
	return merge(prev, Violation{
		Field:   f,
		Code:    ValidationCodeTooShort,
		Message: "Value for " + f + " is too short",
	})
}

// RequireMaxLength validates that a string value has at most the specified maximum length.
// It returns a ValidationError with ValidationCodeTooLong if the value is longer than max.
// If the value meets the maximum length, it returns the previous ValidationError unchanged.
func RequireMaxLength(field string, max int, value string, prev *ValidationError) *ValidationError {
	if max < 0 {
		panic("max must be greater than or equal to 0")
	}
	if len(value) <= max {
		return prev
	}
	return merge(prev, Violation{
		Field:   field,
		Code:    ValidationCodeTooLong,
		Message: "Value for " + field + " is too long",
	})
}

// RequireMaxLengthIndexed validates that a string value has at most the specified maximum length.
// It returns a ValidationError with ValidationCodeTooLong if the value is longer than max.
// If the value meets the maximum length, it returns the previous ValidationError unchanged.
// The field name is formatted using the fieldFormat string and the index.
// The message is formatted using the fieldFormat string and the index.
func RequireMaxLengthIndexed(fieldFormat string, index int, max int, value string, prev *ValidationError) *ValidationError {
	if max < 0 {
		panic("max must be greater than or equal to 0")
	}
	if len(value) <= max {
		return prev
	}
	f := fmt.Sprintf(fieldFormat, index)
	return merge(prev, Violation{
		Field:   f,
		Code:    ValidationCodeTooLong,
		Message: "Value for " + f + " is too long",
	})
}

// RequireEnumValue validates that a value is in the allowed list.
// It returns a ValidationError with ValidationCodeInvalid if the value is not in the allowed list.
// If the value is in the allowed list, it returns the previous ValidationError unchanged.
func RequireEnumValue[T comparable](field string, value T, allowed []T, prev *ValidationError) *ValidationError {
	if slices.Contains(allowed, value) {
		return prev
	}
	return merge(prev, Violation{
		Field:   field,
		Code:    ValidationCodeInvalid,
		Message: "Value for " + field + " is invalid",
	})
}

// RequireEnumValueIndexed validates that a value is in the allowed list.

// It returns a ValidationError with ValidationCodeInvalid if the value is not in the allowed list.
// If the value is in the allowed list, it returns the previous ValidationError unchanged.
func RequireEnumValueIndexed[T comparable](fieldFormat string, index int, value T, allowed []T, prev *ValidationError) *ValidationError {
	if slices.Contains(allowed, value) {
		return prev
	}
	f := fmt.Sprintf(fieldFormat, index)
	return merge(prev, Violation{
		Field:   f,
		Code:    ValidationCodeInvalid,
		Message: "Value for " + f + " is invalid",
	})
}

// RequireRegex validates that a string value matches the specified regular expression.
// It returns a ValidationError with ValidationCodeInvalid if the value does not match the pattern.
// If the value matches the pattern, it returns the previous ValidationError unchanged.
func RequireRegex(field string, value string, pattern *regexp.Regexp, prev *ValidationError) *ValidationError {
	if pattern.MatchString(value) {
		return prev
	}
	return merge(prev, Violation{
		Field:   field,
		Code:    ValidationCodeInvalid,
		Message: "Value for " + field + " is invalid",
	})
}

// RequireRegexIndexed validates that a string value matches the specified regular expression.
// It returns a ValidationError with ValidationCodeInvalid if the value does not match the pattern.
// If the value matches the pattern, it returns the previous ValidationError unchanged.
func RequireRegexIndexed(fieldFormat string, index int, value string, pattern *regexp.Regexp, prev *ValidationError) *ValidationError {
	if pattern.MatchString(value) {
		return prev
	}
	f := fmt.Sprintf(fieldFormat, index)
	return merge(prev, Violation{
		Field:   f,
		Code:    ValidationCodeInvalid,
		Message: "Value for " + f + " is invalid",
	})
}

// RequireNotEmptySlice validates that a slice is not empty.
// It returns a ValidationError with ValidationCodeRequired if the slice is empty.
// If the slice is not empty, it returns the previous ValidationError unchanged.
func RequireNotEmptySlice[T any](field string, value []T, prev *ValidationError) *ValidationError {
	if len(value) > 0 {
		return prev
	}
	return merge(prev, Violation{
		Field:   field,
		Code:    ValidationCodeRequired,
		Message: "Value for " + field + " is required",
	})
}

// RequireNotEmptySliceIndexed validates that a slice is not empty.
// It returns a ValidationError with ValidationCodeRequired if the slice is empty.
// If the slice is not empty, it returns the previous ValidationError unchanged.
func RequireNotEmptySliceIndexed[T any](fieldFormat string, index int, value []T, prev *ValidationError) *ValidationError {
	if len(value) > 0 {
		return prev
	}
	f := fmt.Sprintf(fieldFormat, index)
	return merge(prev, Violation{
		Field:   f,
		Code:    ValidationCodeRequired,
		Message: "Value for " + f + " is required",
	})
}

// RequireMinLengthSlice validates that a slice has at least the specified minimum length.
// It returns a ValidationError with ValidationCodeTooFewItems if the slice is shorter than min.
// If the slice meets the minimum length, it returns the previous ValidationError unchanged.
func RequireMinLengthSlice[T any](field string, min int, value []T, prev *ValidationError) *ValidationError {
	if len(value) >= min {
		return prev
	}
	return merge(prev, Violation{
		Field:   field,
		Code:    ValidationCodeTooFewItems,
		Message: "Too few items in " + field,
	})
}

// RequireMinLengthSliceIndexed validates that a slice has at least the specified minimum length.
// It returns a ValidationError with ValidationCodeTooFewItems if the slice is shorter than min.
// If the slice meets the minimum length, it returns the previous ValidationError unchanged.
func RequireMinLengthSliceIndexed[T any](fieldFormat string, index int, min int, value []T, prev *ValidationError) *ValidationError {
	if min < 0 {
		panic("min must be greater than or equal to 0")
	}
	if len(value) >= min {
		return prev
	}
	f := fmt.Sprintf(fieldFormat, index)
	return merge(prev, Violation{
		Field:   f,
		Code:    ValidationCodeTooFewItems,
		Message: "Too few items in " + f,
	})
}

// RequireMaxLengthSlice validates that a slice has at most the specified maximum length.
// It returns a ValidationError with ValidationCodeTooManyItems if the slice is longer than max.
// If the slice meets the maximum length, it returns the previous ValidationError unchanged.
func RequireMaxLengthSlice[T any](field string, max int, value []T, prev *ValidationError) *ValidationError {
	if len(value) <= max {
		return prev
	}
	return merge(prev, Violation{
		Field:   field,
		Code:    ValidationCodeTooManyItems,
		Message: "Too many items in " + field,
	})
}

// RequireMaxLengthSliceIndexed validates that a slice has at most the specified maximum length.

// It returns a ValidationError with ValidationCodeTooManyItems if the slice is longer than max.
// If the slice meets the maximum length, it returns the previous ValidationError unchanged.
func RequireMaxLengthSliceIndexed[T any](fieldFormat string, index int, max int, value []T, prev *ValidationError) *ValidationError {
	if max < 0 {
		panic("max must be greater than or equal to 0")
	}
	if len(value) <= max {
		return prev
	}
	f := fmt.Sprintf(fieldFormat, index)
	return merge(prev, Violation{
		Field:   f,
		Code:    ValidationCodeTooManyItems,
		Message: "Too many items in " + f,
	})
}

// Validate converts a ValidationError to a standard error.
// If the ValidationError is nil, it returns nil.
func Validate(v *ValidationError) error {
	if v == nil {
		return nil
	}
	return v
}

func merge(prev *ValidationError, v ...Violation) *ValidationError {
	if prev != nil {
		prev.Errors = append(prev.Errors, v...)
		return prev
	}

	return &ValidationError{
		Code:    "invalid_data",
		Message: "Invalid data",
		Errors:  v,
	}
}
