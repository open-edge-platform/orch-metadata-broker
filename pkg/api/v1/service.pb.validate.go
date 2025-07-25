/*
* SPDX-FileCopyrightText: (C) 2023 Intel Corporation
* SPDX-License-Identifier: Apache-2.0
*/
// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: v1/service.proto

package v1

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/protobuf/types/known/anypb"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = anypb.Any{}
	_ = sort.Sort
)

// Validate checks the field values on MetadataList with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *MetadataList) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on MetadataList with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in MetadataListMultiError, or
// nil if none found.
func (m *MetadataList) ValidateAll() error {
	return m.validate(true)
}

func (m *MetadataList) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	for idx, item := range m.GetMetadata() {
		_, _ = idx, item

		if all {
			switch v := interface{}(item).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, MetadataListValidationError{
						field:  fmt.Sprintf("Metadata[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, MetadataListValidationError{
						field:  fmt.Sprintf("Metadata[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return MetadataListValidationError{
					field:  fmt.Sprintf("Metadata[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	if len(errors) > 0 {
		return MetadataListMultiError(errors)
	}

	return nil
}

// MetadataListMultiError is an error wrapping multiple validation errors
// returned by MetadataList.ValidateAll() if the designated constraints aren't met.
type MetadataListMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m MetadataListMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m MetadataListMultiError) AllErrors() []error { return m }

// MetadataListValidationError is the validation error returned by
// MetadataList.Validate if the designated constraints aren't met.
type MetadataListValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e MetadataListValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e MetadataListValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e MetadataListValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e MetadataListValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e MetadataListValidationError) ErrorName() string { return "MetadataListValidationError" }

// Error satisfies the builtin error interface
func (e MetadataListValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sMetadataList.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = MetadataListValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = MetadataListValidationError{}

// Validate checks the field values on CreateOrUpdateRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *CreateOrUpdateRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on CreateOrUpdateRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// CreateOrUpdateRequestMultiError, or nil if none found.
func (m *CreateOrUpdateRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *CreateOrUpdateRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if all {
		switch v := interface{}(m.GetBody()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, CreateOrUpdateRequestValidationError{
					field:  "Body",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, CreateOrUpdateRequestValidationError{
					field:  "Body",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetBody()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return CreateOrUpdateRequestValidationError{
				field:  "Body",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if len(errors) > 0 {
		return CreateOrUpdateRequestMultiError(errors)
	}

	return nil
}

// CreateOrUpdateRequestMultiError is an error wrapping multiple validation
// errors returned by CreateOrUpdateRequest.ValidateAll() if the designated
// constraints aren't met.
type CreateOrUpdateRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m CreateOrUpdateRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m CreateOrUpdateRequestMultiError) AllErrors() []error { return m }

// CreateOrUpdateRequestValidationError is the validation error returned by
// CreateOrUpdateRequest.Validate if the designated constraints aren't met.
type CreateOrUpdateRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e CreateOrUpdateRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e CreateOrUpdateRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e CreateOrUpdateRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e CreateOrUpdateRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e CreateOrUpdateRequestValidationError) ErrorName() string {
	return "CreateOrUpdateRequestValidationError"
}

// Error satisfies the builtin error interface
func (e CreateOrUpdateRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sCreateOrUpdateRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = CreateOrUpdateRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = CreateOrUpdateRequestValidationError{}

// Validate checks the field values on MetadataResponse with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *MetadataResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on MetadataResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// MetadataResponseMultiError, or nil if none found.
func (m *MetadataResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *MetadataResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	for idx, item := range m.GetMetadata() {
		_, _ = idx, item

		if all {
			switch v := interface{}(item).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, MetadataResponseValidationError{
						field:  fmt.Sprintf("Metadata[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, MetadataResponseValidationError{
						field:  fmt.Sprintf("Metadata[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return MetadataResponseValidationError{
					field:  fmt.Sprintf("Metadata[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	if len(errors) > 0 {
		return MetadataResponseMultiError(errors)
	}

	return nil
}

// MetadataResponseMultiError is an error wrapping multiple validation errors
// returned by MetadataResponse.ValidateAll() if the designated constraints
// aren't met.
type MetadataResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m MetadataResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m MetadataResponseMultiError) AllErrors() []error { return m }

// MetadataResponseValidationError is the validation error returned by
// MetadataResponse.Validate if the designated constraints aren't met.
type MetadataResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e MetadataResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e MetadataResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e MetadataResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e MetadataResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e MetadataResponseValidationError) ErrorName() string { return "MetadataResponseValidationError" }

// Error satisfies the builtin error interface
func (e MetadataResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sMetadataResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = MetadataResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = MetadataResponseValidationError{}

// Validate checks the field values on DeleteProjectRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *DeleteProjectRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on DeleteProjectRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// DeleteProjectRequestMultiError, or nil if none found.
func (m *DeleteProjectRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *DeleteProjectRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Id

	if len(errors) > 0 {
		return DeleteProjectRequestMultiError(errors)
	}

	return nil
}

// DeleteProjectRequestMultiError is an error wrapping multiple validation
// errors returned by DeleteProjectRequest.ValidateAll() if the designated
// constraints aren't met.
type DeleteProjectRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m DeleteProjectRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m DeleteProjectRequestMultiError) AllErrors() []error { return m }

// DeleteProjectRequestValidationError is the validation error returned by
// DeleteProjectRequest.Validate if the designated constraints aren't met.
type DeleteProjectRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e DeleteProjectRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e DeleteProjectRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e DeleteProjectRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e DeleteProjectRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e DeleteProjectRequestValidationError) ErrorName() string {
	return "DeleteProjectRequestValidationError"
}

// Error satisfies the builtin error interface
func (e DeleteProjectRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDeleteProjectRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = DeleteProjectRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = DeleteProjectRequestValidationError{}
