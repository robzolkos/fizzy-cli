// Package errors defines error types and exit codes for the Fizzy CLI.
package errors

import "fmt"

// Exit codes used by the CLI.
const (
	ExitSuccess     = 0
	ExitError       = 1
	ExitInvalidArgs = 2
	ExitAuthFailure = 3
	ExitForbidden   = 4
	ExitNotFound    = 5
	ExitValidation  = 6
	ExitNetwork     = 7
)

// CLIError represents an error with an associated exit code.
type CLIError struct {
	Code     string
	Message  string
	Status   int
	ExitCode int
}

func (e *CLIError) Error() string {
	return e.Message
}

// NewError creates a general error.
func NewError(message string) *CLIError {
	return &CLIError{
		Code:     "ERROR",
		Message:  message,
		ExitCode: ExitError,
	}
}

// NewAuthError creates an authentication error.
func NewAuthError(message string) *CLIError {
	return &CLIError{
		Code:     "AUTH_ERROR",
		Message:  message,
		Status:   401,
		ExitCode: ExitAuthFailure,
	}
}

// NewForbiddenError creates a permission denied error.
func NewForbiddenError(message string) *CLIError {
	return &CLIError{
		Code:     "FORBIDDEN",
		Message:  message,
		Status:   403,
		ExitCode: ExitForbidden,
	}
}

// NewNotFoundError creates a not found error.
func NewNotFoundError(message string) *CLIError {
	return &CLIError{
		Code:     "NOT_FOUND",
		Message:  message,
		Status:   404,
		ExitCode: ExitNotFound,
	}
}

// NewValidationError creates a validation error.
func NewValidationError(message string) *CLIError {
	return &CLIError{
		Code:     "VALIDATION_ERROR",
		Message:  message,
		Status:   422,
		ExitCode: ExitValidation,
	}
}

// NewNetworkError creates a network error.
func NewNetworkError(message string) *CLIError {
	return &CLIError{
		Code:     "NETWORK_ERROR",
		Message:  message,
		ExitCode: ExitNetwork,
	}
}

// NewInvalidArgsError creates an invalid arguments error.
func NewInvalidArgsError(message string) *CLIError {
	return &CLIError{
		Code:     "INVALID_ARGS",
		Message:  message,
		ExitCode: ExitInvalidArgs,
	}
}

// FromHTTPStatus creates an appropriate error from an HTTP status code.
func FromHTTPStatus(status int, message string) *CLIError {
	switch status {
	case 401:
		return NewAuthError(message)
	case 403:
		return NewForbiddenError(message)
	case 404:
		return NewNotFoundError(message)
	case 422:
		return NewValidationError(message)
	default:
		return &CLIError{
			Code:     "ERROR",
			Message:  fmt.Sprintf("Request failed: %d %s", status, message),
			Status:   status,
			ExitCode: ExitError,
		}
	}
}
