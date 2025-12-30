package commands

import (
	"testing"

	"github.com/robzolkos/fizzy-cli/internal/client"
	"github.com/robzolkos/fizzy-cli/internal/errors"
)

func TestReactionList(t *testing.T) {
	t.Run("returns list of reactions", func(t *testing.T) {
		mock := NewMockClient()
		mock.GetResponse = &client.APIResponse{
			StatusCode: 200,
			Data: []interface{}{
				map[string]interface{}{"id": "1", "content": "👍"},
			},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		reactionListCard = "42"
		reactionListComment = "comment-1"
		RunTestCommand(func() {
			reactionListCmd.Run(reactionListCmd, []string{})
		})
		reactionListCard = ""
		reactionListComment = ""

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		if mock.GetCalls[0].Path != "/cards/42/comments/comment-1/reactions.json" {
			t.Errorf("expected path '/cards/42/comments/comment-1/reactions.json', got '%s'", mock.GetCalls[0].Path)
		}
	})

	t.Run("requires card flag", func(t *testing.T) {
		mock := NewMockClient()
		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		reactionListCard = ""
		reactionListComment = "comment-1"
		RunTestCommand(func() {
			reactionListCmd.Run(reactionListCmd, []string{})
		})
		reactionListComment = ""

		if result.ExitCode != errors.ExitInvalidArgs {
			t.Errorf("expected exit code %d, got %d", errors.ExitInvalidArgs, result.ExitCode)
		}
	})

	t.Run("requires comment flag", func(t *testing.T) {
		mock := NewMockClient()
		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		reactionListCard = "42"
		reactionListComment = ""
		RunTestCommand(func() {
			reactionListCmd.Run(reactionListCmd, []string{})
		})
		reactionListCard = ""

		if result.ExitCode != errors.ExitInvalidArgs {
			t.Errorf("expected exit code %d, got %d", errors.ExitInvalidArgs, result.ExitCode)
		}
	})
}

func TestReactionCreate(t *testing.T) {
	t.Run("creates reaction", func(t *testing.T) {
		mock := NewMockClient()
		mock.PostResponse = &client.APIResponse{
			StatusCode: 201,
			Data:       map[string]interface{}{},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		reactionCreateCard = "42"
		reactionCreateComment = "comment-1"
		reactionCreateContent = "👍"
		RunTestCommand(func() {
			reactionCreateCmd.Run(reactionCreateCmd, []string{})
		})
		reactionCreateCard = ""
		reactionCreateComment = ""
		reactionCreateContent = ""

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		if mock.PostCalls[0].Path != "/cards/42/comments/comment-1/reactions.json" {
			t.Errorf("expected path '/cards/42/comments/comment-1/reactions.json', got '%s'", mock.PostCalls[0].Path)
		}

		body := mock.PostCalls[0].Body.(map[string]interface{})
		if body["content"] != "👍" {
			t.Errorf("expected content '👍', got '%v'", body["content"])
		}
	})
}

func TestReactionDelete(t *testing.T) {
	t.Run("deletes reaction", func(t *testing.T) {
		mock := NewMockClient()
		mock.DeleteResponse = &client.APIResponse{
			StatusCode: 204,
			Data:       map[string]interface{}{},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		reactionDeleteCard = "42"
		reactionDeleteComment = "comment-1"
		RunTestCommand(func() {
			reactionDeleteCmd.Run(reactionDeleteCmd, []string{"reaction-1"})
		})
		reactionDeleteCard = ""
		reactionDeleteComment = ""

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		if mock.DeleteCalls[0].Path != "/cards/42/comments/comment-1/reactions/reaction-1.json" {
			t.Errorf("expected path '/cards/42/comments/comment-1/reactions/reaction-1.json', got '%s'", mock.DeleteCalls[0].Path)
		}
	})
}
