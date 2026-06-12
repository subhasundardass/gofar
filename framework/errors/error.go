// Package errors defines typed HTTP-aware errors used throughout GoFar.
//
// Using typed errors lets the response layer translate them into consistent
// JSON or HTML error responses without scattering status-code logic across
// handlers.
//
// Example:
//
//	// In a handler:
//	user, err := repo.FindByID(id)
//	if err != nil {
//	    return errors.NewNotFound("user", id)
//	}
//
//	// In response middleware:
//	var appErr *errors.AppError
//	if errors.As(err, &appErr) {
//	    return c.Status(appErr.Status).JSON(appErr)
//	}
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError is the base error type for all GoFar application errors.
// It carries an HTTP status code, a short machine-readable code, a
// human-readable message, and an optional wrapped cause.
type AppError struct {
	// Status is the HTTP status code that should be returned to the client.
	Status int `json:"status"`
	// Code is a short camelCase identifier e.g. "notFound", "validation".
	Code string `json:"code"`
	// Message is a human-readable explanation safe to expose to the client.
	Message string `json:"message"`
	// cause is the underlying Go error (not serialised).
	cause error
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the wrapped cause so errors.Is / errors.As work as expected.
func (e *AppError) Unwrap() error { return e.cause }

// New creates an AppError with the given HTTP status, code, and message.
func New(status int, code, message string) *AppError {
	return &AppError{Status: status, Code: code, Message: message}
}

// Wrap creates an AppError that wraps an existing error as its cause.
func Wrap(status int, code, message string, cause error) *AppError {
	return &AppError{Status: status, Code: code, Message: message, cause: cause}
}

// ── Constructors for common HTTP errors ──────────────────────────────────────

// NewNotFound returns a 404 error describing the missing resource.
//
//	errors.NewNotFound("user", id)  →  "user 42 not found"
func NewNotFound(resource string, id any) *AppError {
	return New(http.StatusNotFound, "notFound",
		fmt.Sprintf("%s %v not found", resource, id))
}

// NewValidation returns a 422 error for failed input validation.
// fields should be a map of field name → error message(s).
func NewValidation(fields map[string]string) *ValidationError {
	return &ValidationError{
		AppError: AppError{
			Status:  http.StatusUnprocessableEntity,
			Code:    "validation",
			Message: "validation failed",
		},
		Fields: fields,
	}
}

// NewUnauthorized returns a 401 error.
func NewUnauthorized(message string) *AppError {
	return New(http.StatusUnauthorized, "unauthorized", message)
}

// NewForbidden returns a 403 error.
func NewForbidden(message string) *AppError {
	return New(http.StatusForbidden, "forbidden", message)
}

// NewBadRequest returns a 400 error.
func NewBadRequest(message string) *AppError {
	return New(http.StatusBadRequest, "badRequest", message)
}

// NewInternal wraps an internal error and returns a 500. The original error
// is available via errors.Unwrap but is NOT exposed in Message to avoid
// leaking internals to clients.
func NewInternal(cause error) *AppError {
	return Wrap(http.StatusInternalServerError, "internal",
		"an internal server error occurred", cause)
}

// NewConflict returns a 409 error, useful for duplicate-resource scenarios.
func NewConflict(resource string) *AppError {
	return New(http.StatusConflict, "conflict",
		fmt.Sprintf("%s already exists", resource))
}

// ── ValidationError extends AppError with field-level detail ─────────────────

// ValidationError is a specialised AppError that carries per-field validation
// messages. It is returned by NewValidation and can be decoded by the client
// to highlight specific form fields.
type ValidationError struct {
	AppError
	// Fields maps form/JSON field names to their validation messages.
	Fields map[string]string `json:"fields,omitempty"`
}

// ── Helpers that delegate to the standard library ────────────────────────────

// As reports whether any error in err's chain matches target and if so sets
// target and returns true. Delegates to errors.As.
func As(err error, target any) bool { return errors.As(err, target) }

// Is reports whether any error in err's chain matches target.
// Delegates to errors.Is.
func Is(err, target error) bool { return errors.Is(err, target) }
