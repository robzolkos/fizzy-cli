package commands

import (
	"testing"

	"github.com/robzolkos/fizzy-cli/internal/client"
)

func TestUserList(t *testing.T) {
	t.Run("returns list of users", func(t *testing.T) {
		mock := NewMockClient()
		mock.GetWithPaginationResponse = &client.APIResponse{
			StatusCode: 200,
			Data: []interface{}{
				map[string]interface{}{"id": "1", "name": "User 1"},
				map[string]interface{}{"id": "2", "name": "User 2"},
			},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		RunTestCommand(func() {
			userListCmd.Run(userListCmd, []string{})
		})

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		if mock.GetWithPaginationCalls[0].Path != "/users.json" {
			t.Errorf("expected path '/users.json', got '%s'", mock.GetWithPaginationCalls[0].Path)
		}
	})
}

func TestUserShow(t *testing.T) {
	t.Run("shows user by ID", func(t *testing.T) {
		mock := NewMockClient()
		mock.GetResponse = &client.APIResponse{
			StatusCode: 200,
			Data: map[string]interface{}{
				"id":   "user-1",
				"name": "Test User",
			},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		RunTestCommand(func() {
			userShowCmd.Run(userShowCmd, []string{"user-1"})
		})

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		if mock.GetCalls[0].Path != "/users/user-1.json" {
			t.Errorf("expected path '/users/user-1.json', got '%s'", mock.GetCalls[0].Path)
		}
	})
}
