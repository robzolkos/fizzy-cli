package tests

import (
	"testing"

	"github.com/robzolkos/fizzy-cli/e2e/harness"
)

func TestErrorAuthFailure(t *testing.T) {
	cfg := harness.LoadConfig()
	if cfg.Token == "" || cfg.Account == "" {
		t.Skip("FIZZY_TEST_TOKEN or FIZZY_TEST_ACCOUNT not set")
	}

	// Create harness with invalid token
	h := harness.NewWithConfig(t, &harness.Config{
		BinaryPath: cfg.BinaryPath,
		Token:      "invalid-token-12345",
		Account:    cfg.Account,
		APIURL:     cfg.APIURL,
	})

	t.Run("returns exit code 3 for auth failure", func(t *testing.T) {
		result := h.Run("board", "list")

		if result.ExitCode != harness.ExitAuthFailure {
			t.Errorf("expected exit code %d, got %d\nstdout: %s",
				harness.ExitAuthFailure, result.ExitCode, result.Stdout)
		}

		if result.Response == nil {
			t.Fatal("expected JSON response")
		}

		if result.Response.Success {
			t.Error("expected success=false")
		}

		if result.Response.Error == nil {
			t.Error("expected error in response")
		} else {
			if result.Response.Error.Code != "AUTH_ERROR" {
				t.Errorf("expected error code AUTH_ERROR, got %s", result.Response.Error.Code)
			}
		}
	})
}

func TestErrorNotFound(t *testing.T) {
	h := harness.New(t)

	testCases := []struct {
		name string
		args []string
	}{
		{"board show", []string{"board", "show", "non-existent-id"}},
		{"card show", []string{"card", "show", "999999999"}},
		{"user show", []string{"user", "show", "non-existent-id"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name+" returns exit code 5", func(t *testing.T) {
			result := h.Run(tc.args...)

			if result.ExitCode != harness.ExitNotFound {
				t.Errorf("expected exit code %d, got %d\nstdout: %s",
					harness.ExitNotFound, result.ExitCode, result.Stdout)
			}

			if result.Response == nil {
				t.Fatal("expected JSON response")
			}

			if result.Response.Success {
				t.Error("expected success=false")
			}

			if result.Response.Error == nil {
				t.Error("expected error in response")
			} else {
				if result.Response.Error.Code != "NOT_FOUND" {
					t.Errorf("expected error code NOT_FOUND, got %s", result.Response.Error.Code)
				}
			}
		})
	}
}

func TestErrorNetworkFailure(t *testing.T) {
	cfg := harness.LoadConfig()
	if cfg.Token == "" || cfg.Account == "" {
		t.Skip("FIZZY_TEST_TOKEN or FIZZY_TEST_ACCOUNT not set")
	}

	// Create harness with invalid API URL (unreachable)
	h := harness.NewWithConfig(t, &harness.Config{
		BinaryPath: cfg.BinaryPath,
		Token:      cfg.Token,
		Account:    cfg.Account,
		APIURL:     "http://localhost:59999", // Unlikely to be listening
	})

	t.Run("returns exit code 7 for network error", func(t *testing.T) {
		result := h.Run("board", "list")

		if result.ExitCode != harness.ExitNetwork {
			t.Errorf("expected exit code %d, got %d\nstdout: %s\nstderr: %s",
				harness.ExitNetwork, result.ExitCode, result.Stdout, result.Stderr)
		}

		if result.Response == nil {
			t.Fatal("expected JSON response")
		}

		if result.Response.Success {
			t.Error("expected success=false")
		}

		if result.Response.Error == nil {
			t.Error("expected error in response")
		} else {
			if result.Response.Error.Code != "NETWORK_ERROR" {
				t.Errorf("expected error code NETWORK_ERROR, got %s", result.Response.Error.Code)
			}
		}
	})
}

func TestErrorResponseFormat(t *testing.T) {
	h := harness.New(t)

	t.Run("error response has correct structure", func(t *testing.T) {
		// Use an invalid board ID to trigger a not found error
		result := h.Run("board", "show", "non-existent-board-id")

		if result.Response == nil {
			t.Fatal("expected JSON response")
		}

		// Check response structure
		if result.Response.Success {
			t.Error("expected success=false for error")
		}

		if result.Response.Error == nil {
			t.Fatal("expected error object in response")
		}

		// Error should have code and message
		if result.Response.Error.Code == "" {
			t.Error("expected error code")
		}

		if result.Response.Error.Message == "" {
			t.Error("expected error message")
		}

		// Meta should still be present
		if result.Response.Meta == nil {
			t.Error("expected meta in response even for errors")
		}
	})
}

func TestSuccessResponseFormat(t *testing.T) {
	h := harness.New(t)

	t.Run("success response has correct structure", func(t *testing.T) {
		result := h.Run("board", "list")

		if result.ExitCode != harness.ExitSuccess {
			t.Fatalf("expected exit code %d, got %d", harness.ExitSuccess, result.ExitCode)
		}

		if result.Response == nil {
			t.Fatal("expected JSON response")
		}

		// Check response structure
		if !result.Response.Success {
			t.Error("expected success=true")
		}

		// Data should be present (can be empty array)
		if result.Response.Data == nil {
			t.Error("expected data in response")
		}

		// Error should be nil/absent for success
		if result.Response.Error != nil {
			t.Error("expected no error in success response")
		}

		// Pagination may be absent when all results fit in a single page
		// (the CLI only includes pagination when there is a next page)

		// Meta should be present
		if result.Response.Meta == nil {
			t.Error("expected meta in response")
		}
	})
}
