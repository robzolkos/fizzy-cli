package tests

import (
	"testing"

	"github.com/robzolkos/fizzy-cli/e2e/harness"
)

func TestUserList(t *testing.T) {
	h := harness.New(t)

	t.Run("returns list of users", func(t *testing.T) {
		result := h.Run("user", "list")

		if result.ExitCode != harness.ExitSuccess {
			t.Errorf("expected exit code %d, got %d\nstderr: %s", harness.ExitSuccess, result.ExitCode, result.Stderr)
		}

		if result.Response == nil {
			t.Fatalf("expected JSON response, got nil\nstdout: %s", result.Stdout)
		}

		if !result.Response.Success {
			t.Error("expected success=true")
		}

		arr := result.GetDataArray()
		if arr == nil {
			t.Error("expected data to be an array")
		}

		// Should have at least one user (the authenticated user)
		if len(arr) == 0 {
			t.Error("expected at least one user")
		}
	})

	t.Run("supports --page option", func(t *testing.T) {
		result := h.Run("user", "list", "--page", "1")

		if result.ExitCode != harness.ExitSuccess {
			t.Errorf("expected exit code %d, got %d", harness.ExitSuccess, result.ExitCode)
		}

		if result.Response == nil || !result.Response.Success {
			t.Error("expected successful response")
		}
	})

	t.Run("supports --all flag", func(t *testing.T) {
		result := h.Run("user", "list", "--all")

		if result.ExitCode != harness.ExitSuccess {
			t.Errorf("expected exit code %d, got %d", harness.ExitSuccess, result.ExitCode)
		}
	})
}

func TestUserShow(t *testing.T) {
	h := harness.New(t)

	// First get a valid user ID from the list
	listResult := h.Run("user", "list")
	if listResult.ExitCode != harness.ExitSuccess {
		t.Fatalf("failed to list users: %s", listResult.Stderr)
	}

	users := listResult.GetDataArray()
	if len(users) == 0 {
		t.Skip("no users available")
	}

	firstUser, ok := users[0].(map[string]interface{})
	if !ok {
		t.Fatal("expected user to be a map")
	}

	userID, ok := firstUser["id"].(string)
	if !ok || userID == "" {
		t.Fatal("expected user to have id")
	}

	t.Run("returns user details", func(t *testing.T) {
		result := h.Run("user", "show", userID)

		if result.ExitCode != harness.ExitSuccess {
			t.Errorf("expected exit code %d, got %d\nstderr: %s", harness.ExitSuccess, result.ExitCode, result.Stderr)
		}

		if result.Response == nil {
			t.Fatal("expected JSON response")
		}

		if !result.Response.Success {
			t.Error("expected success=true")
		}

		id := result.GetDataString("id")
		if id != userID {
			t.Errorf("expected id %q, got %q", userID, id)
		}
	})
}

func TestUserShowNotFound(t *testing.T) {
	h := harness.New(t)

	t.Run("returns not found for non-existent user", func(t *testing.T) {
		result := h.Run("user", "show", "non-existent-user-id-12345")

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
	})
}

// Note: We don't test user update/deactivate as they would modify real users
// These should be tested manually or with a dedicated test account
