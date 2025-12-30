package errors

import (
	"testing"
)

func TestExitCodes(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected int
	}{
		{"ExitSuccess", ExitSuccess, 0},
		{"ExitError", ExitError, 1},
		{"ExitInvalidArgs", ExitInvalidArgs, 2},
		{"ExitAuthFailure", ExitAuthFailure, 3},
		{"ExitForbidden", ExitForbidden, 4},
		{"ExitNotFound", ExitNotFound, 5},
		{"ExitValidation", ExitValidation, 6},
		{"ExitNetwork", ExitNetwork, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, tt.code)
			}
		})
	}
}

func TestCLIError_Error(t *testing.T) {
	err := &CLIError{
		Code:    "TEST_ERROR",
		Message: "test message",
	}

	if err.Error() != "test message" {
		t.Errorf("expected 'test message', got '%s'", err.Error())
	}
}

func TestNewError(t *testing.T) {
	err := NewError("something went wrong")

	if err.Code != "ERROR" {
		t.Errorf("expected code 'ERROR', got '%s'", err.Code)
	}
	if err.Message != "something went wrong" {
		t.Errorf("expected message 'something went wrong', got '%s'", err.Message)
	}
	if err.ExitCode != ExitError {
		t.Errorf("expected exit code %d, got %d", ExitError, err.ExitCode)
	}
}

func TestNewAuthError(t *testing.T) {
	err := NewAuthError("invalid token")

	if err.Code != "AUTH_ERROR" {
		t.Errorf("expected code 'AUTH_ERROR', got '%s'", err.Code)
	}
	if err.Message != "invalid token" {
		t.Errorf("expected message 'invalid token', got '%s'", err.Message)
	}
	if err.Status != 401 {
		t.Errorf("expected status 401, got %d", err.Status)
	}
	if err.ExitCode != ExitAuthFailure {
		t.Errorf("expected exit code %d, got %d", ExitAuthFailure, err.ExitCode)
	}
}

func TestNewForbiddenError(t *testing.T) {
	err := NewForbiddenError("access denied")

	if err.Code != "FORBIDDEN" {
		t.Errorf("expected code 'FORBIDDEN', got '%s'", err.Code)
	}
	if err.Message != "access denied" {
		t.Errorf("expected message 'access denied', got '%s'", err.Message)
	}
	if err.Status != 403 {
		t.Errorf("expected status 403, got %d", err.Status)
	}
	if err.ExitCode != ExitForbidden {
		t.Errorf("expected exit code %d, got %d", ExitForbidden, err.ExitCode)
	}
}

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("resource not found")

	if err.Code != "NOT_FOUND" {
		t.Errorf("expected code 'NOT_FOUND', got '%s'", err.Code)
	}
	if err.Message != "resource not found" {
		t.Errorf("expected message 'resource not found', got '%s'", err.Message)
	}
	if err.Status != 404 {
		t.Errorf("expected status 404, got %d", err.Status)
	}
	if err.ExitCode != ExitNotFound {
		t.Errorf("expected exit code %d, got %d", ExitNotFound, err.ExitCode)
	}
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("invalid input")

	if err.Code != "VALIDATION_ERROR" {
		t.Errorf("expected code 'VALIDATION_ERROR', got '%s'", err.Code)
	}
	if err.Message != "invalid input" {
		t.Errorf("expected message 'invalid input', got '%s'", err.Message)
	}
	if err.Status != 422 {
		t.Errorf("expected status 422, got %d", err.Status)
	}
	if err.ExitCode != ExitValidation {
		t.Errorf("expected exit code %d, got %d", ExitValidation, err.ExitCode)
	}
}

func TestNewNetworkError(t *testing.T) {
	err := NewNetworkError("connection failed")

	if err.Code != "NETWORK_ERROR" {
		t.Errorf("expected code 'NETWORK_ERROR', got '%s'", err.Code)
	}
	if err.Message != "connection failed" {
		t.Errorf("expected message 'connection failed', got '%s'", err.Message)
	}
	if err.ExitCode != ExitNetwork {
		t.Errorf("expected exit code %d, got %d", ExitNetwork, err.ExitCode)
	}
}

func TestNewInvalidArgsError(t *testing.T) {
	err := NewInvalidArgsError("missing required flag")

	if err.Code != "INVALID_ARGS" {
		t.Errorf("expected code 'INVALID_ARGS', got '%s'", err.Code)
	}
	if err.Message != "missing required flag" {
		t.Errorf("expected message 'missing required flag', got '%s'", err.Message)
	}
	if err.ExitCode != ExitInvalidArgs {
		t.Errorf("expected exit code %d, got %d", ExitInvalidArgs, err.ExitCode)
	}
}

func TestFromHTTPStatus(t *testing.T) {
	tests := []struct {
		name         string
		status       int
		message      string
		expectedCode string
		expectedExit int
	}{
		{"401 Unauthorized", 401, "Unauthorized", "AUTH_ERROR", ExitAuthFailure},
		{"403 Forbidden", 403, "Forbidden", "FORBIDDEN", ExitForbidden},
		{"404 Not Found", 404, "Not Found", "NOT_FOUND", ExitNotFound},
		{"422 Unprocessable", 422, "Validation failed", "VALIDATION_ERROR", ExitValidation},
		{"500 Server Error", 500, "Internal Server Error", "ERROR", ExitError},
		{"502 Bad Gateway", 502, "Bad Gateway", "ERROR", ExitError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := FromHTTPStatus(tt.status, tt.message)

			if err.Code != tt.expectedCode {
				t.Errorf("expected code '%s', got '%s'", tt.expectedCode, err.Code)
			}
			if err.ExitCode != tt.expectedExit {
				t.Errorf("expected exit code %d, got %d", tt.expectedExit, err.ExitCode)
			}
		})
	}
}
