package apperr

import "fmt"

type Kind string

const (
	KindValidation   Kind = "validation"
	KindConflict     Kind = "conflict"
	KindNotFound     Kind = "not_found"
	KindUnauthorized Kind = "unauthorized"
	KindForbidden    Kind = "forbidden"
	KindInternal     Kind = "internal"
)

type Error struct {
	Kind    Kind
	Message string
	Details any
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Kind, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Kind, e.Message)
}

func (e *Error) Unwrap() error { return e.Err }

func Validation(message string, details any) *Error {
	return &Error{Kind: KindValidation, Message: message, Details: details}
}

func Conflict(message string) *Error {
	return &Error{Kind: KindConflict, Message: message}
}

func NotFound(message string) *Error {
	return &Error{Kind: KindNotFound, Message: message}
}

func Unauthorized(message string) *Error {
	return &Error{Kind: KindUnauthorized, Message: message}
}

func Forbidden(message string) *Error {
	return &Error{Kind: KindForbidden, Message: message}
}

func Internal(message string, err error) *Error {
	return &Error{Kind: KindInternal, Message: message, Err: err}
}

func As(err error) (*Error, bool) {
	if err == nil {
		return nil, false
	}
	if e, ok := err.(*Error); ok {
		return e, true
	}
	return nil, false
}
