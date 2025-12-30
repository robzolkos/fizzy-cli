package commands

import (
	"testing"

	"github.com/robzolkos/fizzy-cli/internal/client"
	"github.com/robzolkos/fizzy-cli/internal/errors"
)

func TestCardList(t *testing.T) {
	t.Run("returns list of cards", func(t *testing.T) {
		mock := NewMockClient()
		mock.GetWithPaginationResponse = &client.APIResponse{
			StatusCode: 200,
			Data: []interface{}{
				map[string]interface{}{"id": "1", "title": "Card 1"},
				map[string]interface{}{"id": "2", "title": "Card 2"},
			},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		RunTestCommand(func() {
			cardListCmd.Run(cardListCmd, []string{})
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
		if mock.GetWithPaginationCalls[0].Path != "/cards.json" {
			t.Errorf("expected path '/cards.json', got '%s'", mock.GetWithPaginationCalls[0].Path)
		}
	})

	t.Run("applies filters", func(t *testing.T) {
		mock := NewMockClient()
		mock.GetWithPaginationResponse = &client.APIResponse{
			StatusCode: 200,
			Data:       []interface{}{},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		cardListBoard = "123"
		cardListStatus = "published"
		RunTestCommand(func() {
			cardListCmd.Run(cardListCmd, []string{})
		})
		cardListBoard = ""
		cardListStatus = ""

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		// Check that path contains filters
		path := mock.GetWithPaginationCalls[0].Path
		if path != "/cards.json?board_ids[]=123&status=published" {
			t.Errorf("expected path with filters, got '%s'", path)
		}
	})

	t.Run("filters by pseudo column (maybe)", func(t *testing.T) {
		mock := NewMockClient()
		mock.GetWithPaginationResponse = &client.APIResponse{
			StatusCode: 200,
			Data:       []interface{}{},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		cfg.Board = "123"
		defer ResetTestMode()

		cardListColumn = "maybe"
		RunTestCommand(func() {
			cardListCmd.Run(cardListCmd, []string{})
		})
		cardListColumn = ""

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}

		if mock.GetWithPaginationCalls[0].Path != "/cards.json?board_ids[]=123&indexed_by=not_now" {
			t.Errorf("expected indexed_by filter, got '%s'", mock.GetWithPaginationCalls[0].Path)
		}
	})

	t.Run("requires --all for client-side triage filter", func(t *testing.T) {
		mock := NewMockClient()
		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		cardListColumn = "not-yet"
		cardListAll = false
		cardListPage = 0
		RunTestCommand(func() {
			cardListCmd.Run(cardListCmd, []string{})
		})
		cardListColumn = ""

		if result.ExitCode != errors.ExitInvalidArgs {
			t.Errorf("expected exit code %d, got %d", errors.ExitInvalidArgs, result.ExitCode)
		}
	})

	t.Run("filters triage client-side with --all", func(t *testing.T) {
		mock := NewMockClient()
		mock.GetWithPaginationResponse = &client.APIResponse{
			StatusCode: 200,
			Data: []interface{}{
				map[string]interface{}{"id": "1", "title": "Triage", "column": nil},
				map[string]interface{}{"id": "2", "title": "In Column", "column": map[string]interface{}{"id": "col-1"}},
				map[string]interface{}{"id": "3", "title": "In Column 2", "column_id": "col-2"},
			},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		cardListColumn = "not-yet"
		cardListAll = true
		RunTestCommand(func() {
			cardListCmd.Run(cardListCmd, []string{})
		})
		cardListColumn = ""
		cardListAll = false

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}

		arr, ok := result.Response.Data.([]interface{})
		if !ok {
			t.Fatalf("expected array response data, got %T", result.Response.Data)
		}
		if len(arr) != 1 {
			t.Fatalf("expected 1 triage card, got %d", len(arr))
		}
		card := arr[0].(map[string]interface{})
		if card["id"] != "1" {
			t.Errorf("expected triage card id '1', got '%v'", card["id"])
		}
	})

	t.Run("uses configured board as default filter", func(t *testing.T) {
		mock := NewMockClient()
		mock.GetWithPaginationResponse = &client.APIResponse{
			StatusCode: 200,
			Data:       []interface{}{},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		cfg.Board = "999"
		defer ResetTestMode()

		RunTestCommand(func() {
			cardListCmd.Run(cardListCmd, []string{})
		})

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		if mock.GetWithPaginationCalls[0].Path != "/cards.json?board_ids[]=999" {
			t.Errorf("expected path '/cards.json?board_ids[]=999', got '%s'", mock.GetWithPaginationCalls[0].Path)
		}
	})

	t.Run("requires authentication", func(t *testing.T) {
		mock := NewMockClient()
		result := SetTestMode(mock)
		SetTestConfig("", "account", "https://api.example.com")
		defer ResetTestMode()

		RunTestCommand(func() {
			cardListCmd.Run(cardListCmd, []string{})
		})

		if result.ExitCode != errors.ExitAuthFailure {
			t.Errorf("expected exit code %d, got %d", errors.ExitAuthFailure, result.ExitCode)
		}
	})
}

func TestCardShow(t *testing.T) {
	t.Run("shows card by number", func(t *testing.T) {
		mock := NewMockClient()
		mock.GetResponse = &client.APIResponse{
			StatusCode: 200,
			Data: map[string]interface{}{
				"id":     "123",
				"number": 42,
				"title":  "Test Card",
			},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		RunTestCommand(func() {
			cardShowCmd.Run(cardShowCmd, []string{"42"})
		})

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		if !result.Response.Success {
			t.Error("expected success response")
		}
		if mock.GetCalls[0].Path != "/cards/42.json" {
			t.Errorf("expected path '/cards/42.json', got '%s'", mock.GetCalls[0].Path)
		}
	})

	t.Run("handles not found", func(t *testing.T) {
		mock := NewMockClient()
		mock.GetError = errors.NewNotFoundError("Card not found")

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		RunTestCommand(func() {
			cardShowCmd.Run(cardShowCmd, []string{"999"})
		})

		if result.ExitCode != errors.ExitNotFound {
			t.Errorf("expected exit code %d, got %d", errors.ExitNotFound, result.ExitCode)
		}
	})
}

func TestCardCreate(t *testing.T) {
	t.Run("creates card with required fields", func(t *testing.T) {
		mock := NewMockClient()
		mock.PostResponse = &client.APIResponse{
			StatusCode: 201,
			Location:   "https://api.example.com/cards/42",
		}
		mock.FollowLocationResponse = &client.APIResponse{
			StatusCode: 200,
			Data: map[string]interface{}{
				"id":     "abc",
				"number": 42,
				"title":  "New Card",
			},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		cardCreateBoard = "123"
		cardCreateTitle = "New Card"
		RunTestCommand(func() {
			cardCreateCmd.Run(cardCreateCmd, []string{})
		})
		cardCreateBoard = ""
		cardCreateTitle = ""

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		if mock.PostCalls[0].Path != "/cards.json" {
			t.Errorf("expected path '/cards.json', got '%s'", mock.PostCalls[0].Path)
		}

		body := mock.PostCalls[0].Body.(map[string]interface{})
		if body["board_id"] != "123" {
			t.Errorf("expected board_id '123', got '%v'", body["board_id"])
		}
		if body["title"] != "New Card" {
			t.Errorf("expected title 'New Card', got '%v'", body["title"])
		}
	})

	t.Run("requires board flag", func(t *testing.T) {
		mock := NewMockClient()
		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		cardCreateBoard = ""
		cardCreateTitle = "Test"
		RunTestCommand(func() {
			cardCreateCmd.Run(cardCreateCmd, []string{})
		})
		cardCreateTitle = ""

		if result.ExitCode != errors.ExitInvalidArgs {
			t.Errorf("expected exit code %d, got %d", errors.ExitInvalidArgs, result.ExitCode)
		}
	})

	t.Run("uses configured board when flag omitted", func(t *testing.T) {
		mock := NewMockClient()
		mock.PostResponse = &client.APIResponse{
			StatusCode: 201,
			Location:   "https://api.example.com/cards/42",
		}
		mock.FollowLocationResponse = &client.APIResponse{
			StatusCode: 200,
			Data:       map[string]interface{}{},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		cfg.Board = "555"
		defer ResetTestMode()

		cardCreateBoard = ""
		cardCreateTitle = "New Card"
		RunTestCommand(func() {
			cardCreateCmd.Run(cardCreateCmd, []string{})
		})
		cardCreateTitle = ""

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		body := mock.PostCalls[0].Body.(map[string]interface{})
		if body["board_id"] != "555" {
			t.Errorf("expected board_id '555', got '%v'", body["board_id"])
		}
	})

	t.Run("requires title flag", func(t *testing.T) {
		mock := NewMockClient()
		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		cardCreateBoard = "123"
		cardCreateTitle = ""
		RunTestCommand(func() {
			cardCreateCmd.Run(cardCreateCmd, []string{})
		})
		cardCreateBoard = ""

		if result.ExitCode != errors.ExitInvalidArgs {
			t.Errorf("expected exit code %d, got %d", errors.ExitInvalidArgs, result.ExitCode)
		}
	})

	t.Run("includes optional fields", func(t *testing.T) {
		mock := NewMockClient()
		mock.PostResponse = &client.APIResponse{
			StatusCode: 201,
			Location:   "https://api.example.com/cards/42",
		}
		mock.FollowLocationResponse = &client.APIResponse{
			StatusCode: 200,
			Data:       map[string]interface{}{},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		cardCreateBoard = "123"
		cardCreateTitle = "Test"
		cardCreateDescription = "<p>Description</p>"
		cardCreateTagIDs = "tag1,tag2"
		RunTestCommand(func() {
			cardCreateCmd.Run(cardCreateCmd, []string{})
		})
		cardCreateBoard = ""
		cardCreateTitle = ""
		cardCreateDescription = ""
		cardCreateTagIDs = ""

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}

		body := mock.PostCalls[0].Body.(map[string]interface{})
		if body["description"] != "<p>Description</p>" {
			t.Errorf("expected description '<p>Description</p>', got '%v'", body["description"])
		}
		if body["tag_ids"] != "tag1,tag2" {
			t.Errorf("expected tag_ids 'tag1,tag2', got '%v'", body["tag_ids"])
		}
	})
}

func TestCardUpdate(t *testing.T) {
	t.Run("updates card title", func(t *testing.T) {
		mock := NewMockClient()
		mock.PatchResponse = &client.APIResponse{
			StatusCode: 200,
			Data: map[string]interface{}{
				"id":    "abc",
				"title": "Updated Title",
			},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		cardUpdateTitle = "Updated Title"
		RunTestCommand(func() {
			cardUpdateCmd.Run(cardUpdateCmd, []string{"42"})
		})
		cardUpdateTitle = ""

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		if mock.PatchCalls[0].Path != "/cards/42.json" {
			t.Errorf("expected path '/cards/42.json', got '%s'", mock.PatchCalls[0].Path)
		}
	})
}

func TestCardDelete(t *testing.T) {
	t.Run("deletes card", func(t *testing.T) {
		mock := NewMockClient()
		mock.DeleteResponse = &client.APIResponse{
			StatusCode: 204,
			Data:       map[string]interface{}{},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		RunTestCommand(func() {
			cardDeleteCmd.Run(cardDeleteCmd, []string{"42"})
		})

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		if mock.DeleteCalls[0].Path != "/cards/42.json" {
			t.Errorf("expected path '/cards/42.json', got '%s'", mock.DeleteCalls[0].Path)
		}
	})
}

func TestCardClose(t *testing.T) {
	t.Run("closes card", func(t *testing.T) {
		mock := NewMockClient()
		mock.PostResponse = &client.APIResponse{
			StatusCode: 200,
			Data:       map[string]interface{}{},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		RunTestCommand(func() {
			cardCloseCmd.Run(cardCloseCmd, []string{"42"})
		})

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		if mock.PostCalls[0].Path != "/cards/42/closure.json" {
			t.Errorf("expected path '/cards/42/closure.json', got '%s'", mock.PostCalls[0].Path)
		}
	})
}

func TestCardReopen(t *testing.T) {
	t.Run("reopens card", func(t *testing.T) {
		mock := NewMockClient()
		mock.DeleteResponse = &client.APIResponse{
			StatusCode: 200,
			Data:       map[string]interface{}{},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		RunTestCommand(func() {
			cardReopenCmd.Run(cardReopenCmd, []string{"42"})
		})

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		if mock.DeleteCalls[0].Path != "/cards/42/closure.json" {
			t.Errorf("expected path '/cards/42/closure.json', got '%s'", mock.DeleteCalls[0].Path)
		}
	})
}

func TestCardPostpone(t *testing.T) {
	t.Run("postpones card", func(t *testing.T) {
		mock := NewMockClient()
		mock.PostResponse = &client.APIResponse{
			StatusCode: 200,
			Data:       map[string]interface{}{},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		RunTestCommand(func() {
			cardPostponeCmd.Run(cardPostponeCmd, []string{"42"})
		})

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		if mock.PostCalls[0].Path != "/cards/42/not_now.json" {
			t.Errorf("expected path '/cards/42/not_now.json', got '%s'", mock.PostCalls[0].Path)
		}
	})
}

func TestCardColumn(t *testing.T) {
	t.Run("moves card to column", func(t *testing.T) {
		mock := NewMockClient()
		mock.PostResponse = &client.APIResponse{
			StatusCode: 200,
			Data:       map[string]interface{}{},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		cardColumnColumn = "col-123"
		RunTestCommand(func() {
			cardColumnCmd.Run(cardColumnCmd, []string{"42"})
		})
		cardColumnColumn = ""

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		if mock.PostCalls[0].Path != "/cards/42/triage.json" {
			t.Errorf("expected path '/cards/42/triage.json', got '%s'", mock.PostCalls[0].Path)
		}

		body := mock.PostCalls[0].Body.(map[string]interface{})
		if body["column_id"] != "col-123" {
			t.Errorf("expected column_id 'col-123', got '%v'", body["column_id"])
		}
	})

	t.Run("moves card to pseudo columns", func(t *testing.T) {
		t.Run("not-yet", func(t *testing.T) {
			mock := NewMockClient()
			mock.DeleteResponse = &client.APIResponse{StatusCode: 200, Data: map[string]interface{}{}}

			result := SetTestMode(mock)
			SetTestConfig("token", "account", "https://api.example.com")
			defer ResetTestMode()

			cardColumnColumn = "not-yet"
			RunTestCommand(func() {
				cardColumnCmd.Run(cardColumnCmd, []string{"42"})
			})
			cardColumnColumn = ""

			if result.ExitCode != 0 {
				t.Errorf("expected exit code 0, got %d", result.ExitCode)
			}
			if len(mock.DeleteCalls) != 1 || mock.DeleteCalls[0].Path != "/cards/42/triage.json" {
				t.Errorf("expected delete '/cards/42/triage.json', got %+v", mock.DeleteCalls)
			}
		})

		t.Run("maybe", func(t *testing.T) {
			mock := NewMockClient()
			mock.PostResponse = &client.APIResponse{StatusCode: 200, Data: map[string]interface{}{}}

			result := SetTestMode(mock)
			SetTestConfig("token", "account", "https://api.example.com")
			defer ResetTestMode()

			cardColumnColumn = "maybe"
			RunTestCommand(func() {
				cardColumnCmd.Run(cardColumnCmd, []string{"42"})
			})
			cardColumnColumn = ""

			if result.ExitCode != 0 {
				t.Errorf("expected exit code 0, got %d", result.ExitCode)
			}
			if len(mock.PostCalls) != 1 || mock.PostCalls[0].Path != "/cards/42/not_now.json" {
				t.Errorf("expected post '/cards/42/not_now.json', got %+v", mock.PostCalls)
			}
		})

		t.Run("done", func(t *testing.T) {
			mock := NewMockClient()
			mock.PostResponse = &client.APIResponse{StatusCode: 200, Data: map[string]interface{}{}}

			result := SetTestMode(mock)
			SetTestConfig("token", "account", "https://api.example.com")
			defer ResetTestMode()

			cardColumnColumn = "done"
			RunTestCommand(func() {
				cardColumnCmd.Run(cardColumnCmd, []string{"42"})
			})
			cardColumnColumn = ""

			if result.ExitCode != 0 {
				t.Errorf("expected exit code 0, got %d", result.ExitCode)
			}
			if len(mock.PostCalls) != 1 || mock.PostCalls[0].Path != "/cards/42/closure.json" {
				t.Errorf("expected post '/cards/42/closure.json', got %+v", mock.PostCalls)
			}
		})
	})

	t.Run("requires column flag", func(t *testing.T) {
		mock := NewMockClient()
		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		cardColumnColumn = ""
		RunTestCommand(func() {
			cardColumnCmd.Run(cardColumnCmd, []string{"42"})
		})

		if result.ExitCode != errors.ExitInvalidArgs {
			t.Errorf("expected exit code %d, got %d", errors.ExitInvalidArgs, result.ExitCode)
		}
	})
}

func TestCardUntriage(t *testing.T) {
	t.Run("untriages card", func(t *testing.T) {
		mock := NewMockClient()
		mock.DeleteResponse = &client.APIResponse{
			StatusCode: 200,
			Data:       map[string]interface{}{},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		RunTestCommand(func() {
			cardUntriageCmd.Run(cardUntriageCmd, []string{"42"})
		})

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		if mock.DeleteCalls[0].Path != "/cards/42/triage.json" {
			t.Errorf("expected path '/cards/42/triage.json', got '%s'", mock.DeleteCalls[0].Path)
		}
	})
}

func TestCardAssign(t *testing.T) {
	t.Run("assigns user to card", func(t *testing.T) {
		mock := NewMockClient()
		mock.PostResponse = &client.APIResponse{
			StatusCode: 200,
			Data:       map[string]interface{}{},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		cardAssignUser = "user-123"
		RunTestCommand(func() {
			cardAssignCmd.Run(cardAssignCmd, []string{"42"})
		})
		cardAssignUser = ""

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		if mock.PostCalls[0].Path != "/cards/42/assignments.json" {
			t.Errorf("expected path '/cards/42/assignments.json', got '%s'", mock.PostCalls[0].Path)
		}

		body := mock.PostCalls[0].Body.(map[string]interface{})
		if body["assignee_id"] != "user-123" {
			t.Errorf("expected assignee_id 'user-123', got '%v'", body["assignee_id"])
		}
	})

	t.Run("requires user flag", func(t *testing.T) {
		mock := NewMockClient()
		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		cardAssignUser = ""
		RunTestCommand(func() {
			cardAssignCmd.Run(cardAssignCmd, []string{"42"})
		})

		if result.ExitCode != errors.ExitInvalidArgs {
			t.Errorf("expected exit code %d, got %d", errors.ExitInvalidArgs, result.ExitCode)
		}
	})
}

func TestCardTag(t *testing.T) {
	t.Run("tags card", func(t *testing.T) {
		mock := NewMockClient()
		mock.PostResponse = &client.APIResponse{
			StatusCode: 200,
			Data:       map[string]interface{}{},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		cardTagTag = "bug"
		RunTestCommand(func() {
			cardTagCmd.Run(cardTagCmd, []string{"42"})
		})
		cardTagTag = ""

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		if mock.PostCalls[0].Path != "/cards/42/taggings.json" {
			t.Errorf("expected path '/cards/42/taggings.json', got '%s'", mock.PostCalls[0].Path)
		}

		body := mock.PostCalls[0].Body.(map[string]interface{})
		if body["tag_title"] != "bug" {
			t.Errorf("expected tag_title 'bug', got '%v'", body["tag_title"])
		}
	})

	t.Run("requires tag flag", func(t *testing.T) {
		mock := NewMockClient()
		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		cardTagTag = ""
		RunTestCommand(func() {
			cardTagCmd.Run(cardTagCmd, []string{"42"})
		})

		if result.ExitCode != errors.ExitInvalidArgs {
			t.Errorf("expected exit code %d, got %d", errors.ExitInvalidArgs, result.ExitCode)
		}
	})
}

func TestCardWatch(t *testing.T) {
	t.Run("watches card", func(t *testing.T) {
		mock := NewMockClient()
		mock.PostResponse = &client.APIResponse{
			StatusCode: 200,
			Data:       map[string]interface{}{},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		RunTestCommand(func() {
			cardWatchCmd.Run(cardWatchCmd, []string{"42"})
		})

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		if mock.PostCalls[0].Path != "/cards/42/watch.json" {
			t.Errorf("expected path '/cards/42/watch.json', got '%s'", mock.PostCalls[0].Path)
		}
	})
}

func TestCardUnwatch(t *testing.T) {
	t.Run("unwatches card", func(t *testing.T) {
		mock := NewMockClient()
		mock.DeleteResponse = &client.APIResponse{
			StatusCode: 200,
			Data:       map[string]interface{}{},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		RunTestCommand(func() {
			cardUnwatchCmd.Run(cardUnwatchCmd, []string{"42"})
		})

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		if mock.DeleteCalls[0].Path != "/cards/42/watch.json" {
			t.Errorf("expected path '/cards/42/watch.json', got '%s'", mock.DeleteCalls[0].Path)
		}
	})
}
