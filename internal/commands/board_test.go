package commands

import (
	"testing"

	"github.com/robzolkos/fizzy-cli/internal/client"
	"github.com/robzolkos/fizzy-cli/internal/errors"
)

func TestBoardList(t *testing.T) {
	t.Run("returns list of boards", func(t *testing.T) {
		mock := NewMockClient()
		mock.GetWithPaginationResponse = &client.APIResponse{
			StatusCode: 200,
			Data: []interface{}{
				map[string]interface{}{"id": "1", "name": "Board 1"},
				map[string]interface{}{"id": "2", "name": "Board 2"},
			},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		RunTestCommand(func() {
			boardListCmd.Run(boardListCmd, []string{})
		})

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		if !result.Response.Success {
			t.Error("expected success response")
		}
		if len(mock.GetWithPaginationCalls) != 1 {
			t.Errorf("expected 1 GetWithPagination call, got %d", len(mock.GetWithPaginationCalls))
		}
		if mock.GetWithPaginationCalls[0].Path != "/boards.json" {
			t.Errorf("expected path '/boards.json', got '%s'", mock.GetWithPaginationCalls[0].Path)
		}
	})

	t.Run("handles pagination", func(t *testing.T) {
		mock := NewMockClient()
		mock.GetWithPaginationResponse = &client.APIResponse{
			StatusCode: 200,
			Data:       []interface{}{},
			LinkNext:   "https://api.example.com/boards.json?page=2",
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		boardListPage = 2
		boardListAll = false
		RunTestCommand(func() {
			boardListCmd.Run(boardListCmd, []string{})
		})
		boardListPage = 0 // reset

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
	})

	t.Run("requires authentication", func(t *testing.T) {
		mock := NewMockClient()
		result := SetTestMode(mock)
		SetTestConfig("", "account", "https://api.example.com") // No token
		defer ResetTestMode()

		RunTestCommand(func() {
			boardListCmd.Run(boardListCmd, []string{})
		})

		if result.ExitCode != errors.ExitAuthFailure {
			t.Errorf("expected exit code %d, got %d", errors.ExitAuthFailure, result.ExitCode)
		}
	})

	t.Run("requires account", func(t *testing.T) {
		mock := NewMockClient()
		result := SetTestMode(mock)
		SetTestConfig("token", "", "https://api.example.com") // No account
		defer ResetTestMode()

		RunTestCommand(func() {
			boardListCmd.Run(boardListCmd, []string{})
		})

		if result.ExitCode != errors.ExitInvalidArgs {
			t.Errorf("expected exit code %d, got %d", errors.ExitInvalidArgs, result.ExitCode)
		}
	})
}

func TestBoardShow(t *testing.T) {
	t.Run("shows board by ID", func(t *testing.T) {
		mock := NewMockClient()
		mock.GetResponse = &client.APIResponse{
			StatusCode: 200,
			Data: map[string]interface{}{
				"id":   "123",
				"name": "Test Board",
			},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		RunTestCommand(func() {
			boardShowCmd.Run(boardShowCmd, []string{"123"})
		})

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		if !result.Response.Success {
			t.Error("expected success response")
		}
		if len(mock.GetCalls) != 1 {
			t.Errorf("expected 1 Get call, got %d", len(mock.GetCalls))
		}
		if mock.GetCalls[0].Path != "/boards/123.json" {
			t.Errorf("expected path '/boards/123.json', got '%s'", mock.GetCalls[0].Path)
		}
	})

	t.Run("handles not found", func(t *testing.T) {
		mock := NewMockClient()
		mock.GetError = errors.NewNotFoundError("Board not found")

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		RunTestCommand(func() {
			boardShowCmd.Run(boardShowCmd, []string{"999"})
		})

		if result.ExitCode != errors.ExitNotFound {
			t.Errorf("expected exit code %d, got %d", errors.ExitNotFound, result.ExitCode)
		}
	})
}

func TestBoardCreate(t *testing.T) {
	t.Run("creates board with name", func(t *testing.T) {
		mock := NewMockClient()
		mock.PostResponse = &client.APIResponse{
			StatusCode: 201,
			Location:   "https://api.example.com/boards/456",
			Data:       map[string]interface{}{"id": "456"},
		}
		mock.FollowLocationResponse = &client.APIResponse{
			StatusCode: 200,
			Data: map[string]interface{}{
				"id":   "456",
				"name": "New Board",
			},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		boardCreateName = "New Board"
		RunTestCommand(func() {
			boardCreateCmd.Run(boardCreateCmd, []string{})
		})
		boardCreateName = "" // reset

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		if !result.Response.Success {
			t.Error("expected success response")
		}
		if len(mock.PostCalls) != 1 {
			t.Errorf("expected 1 Post call, got %d", len(mock.PostCalls))
		}
		if mock.PostCalls[0].Path != "/boards.json" {
			t.Errorf("expected path '/boards.json', got '%s'", mock.PostCalls[0].Path)
		}

		// Verify body contains name
		body, ok := mock.PostCalls[0].Body.(map[string]interface{})
		if !ok {
			t.Fatal("expected map body")
		}
		if body["name"] != "New Board" {
			t.Errorf("expected name 'New Board', got '%v'", body["name"])
		}
	})

	t.Run("requires name flag", func(t *testing.T) {
		mock := NewMockClient()
		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		boardCreateName = ""
		RunTestCommand(func() {
			boardCreateCmd.Run(boardCreateCmd, []string{})
		})

		if result.ExitCode != errors.ExitInvalidArgs {
			t.Errorf("expected exit code %d, got %d", errors.ExitInvalidArgs, result.ExitCode)
		}
	})

	t.Run("creates board with options", func(t *testing.T) {
		mock := NewMockClient()
		mock.PostResponse = &client.APIResponse{
			StatusCode: 201,
			Location:   "https://api.example.com/boards/789",
		}
		mock.FollowLocationResponse = &client.APIResponse{
			StatusCode: 200,
			Data:       map[string]interface{}{"id": "789"},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		boardCreateName = "Private Board"
		boardCreateAllAccess = "false"
		boardCreateAutoPostponePeriod = 7
		RunTestCommand(func() {
			boardCreateCmd.Run(boardCreateCmd, []string{})
		})
		boardCreateName = ""
		boardCreateAllAccess = ""
		boardCreateAutoPostponePeriod = 0

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}

		body := mock.PostCalls[0].Body.(map[string]interface{})
		if body["all_access"] != false {
			t.Errorf("expected all_access false, got %v", body["all_access"])
		}
		if body["auto_postpone_period"] != 7 {
			t.Errorf("expected auto_postpone_period 7, got %v", body["auto_postpone_period"])
		}
	})
}

func TestBoardUpdate(t *testing.T) {
	t.Run("updates board name", func(t *testing.T) {
		mock := NewMockClient()
		mock.PatchResponse = &client.APIResponse{
			StatusCode: 200,
			Data: map[string]interface{}{
				"id":   "123",
				"name": "Updated Name",
			},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		boardUpdateName = "Updated Name"
		RunTestCommand(func() {
			boardUpdateCmd.Run(boardUpdateCmd, []string{"123"})
		})
		boardUpdateName = ""

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		if len(mock.PatchCalls) != 1 {
			t.Errorf("expected 1 Patch call, got %d", len(mock.PatchCalls))
		}
		if mock.PatchCalls[0].Path != "/boards/123.json" {
			t.Errorf("expected path '/boards/123.json', got '%s'", mock.PatchCalls[0].Path)
		}
	})

	t.Run("handles API error", func(t *testing.T) {
		mock := NewMockClient()
		mock.PatchError = errors.NewValidationError("Name is too long")

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		boardUpdateName = "Updated"
		RunTestCommand(func() {
			boardUpdateCmd.Run(boardUpdateCmd, []string{"123"})
		})
		boardUpdateName = ""

		if result.ExitCode != errors.ExitValidation {
			t.Errorf("expected exit code %d, got %d", errors.ExitValidation, result.ExitCode)
		}
	})
}

func TestBoardDelete(t *testing.T) {
	t.Run("deletes board", func(t *testing.T) {
		mock := NewMockClient()
		mock.DeleteResponse = &client.APIResponse{
			StatusCode: 204,
			Data:       map[string]interface{}{},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		RunTestCommand(func() {
			boardDeleteCmd.Run(boardDeleteCmd, []string{"123"})
		})

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		if len(mock.DeleteCalls) != 1 {
			t.Errorf("expected 1 Delete call, got %d", len(mock.DeleteCalls))
		}
		if mock.DeleteCalls[0].Path != "/boards/123.json" {
			t.Errorf("expected path '/boards/123.json', got '%s'", mock.DeleteCalls[0].Path)
		}
	})

	t.Run("handles not found", func(t *testing.T) {
		mock := NewMockClient()
		mock.DeleteError = errors.NewNotFoundError("Board not found")

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		RunTestCommand(func() {
			boardDeleteCmd.Run(boardDeleteCmd, []string{"999"})
		})

		if result.ExitCode != errors.ExitNotFound {
			t.Errorf("expected exit code %d, got %d", errors.ExitNotFound, result.ExitCode)
		}
	})
}
