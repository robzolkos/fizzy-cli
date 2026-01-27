package commands

import (
	"testing"

	"github.com/robzolkos/fizzy-cli/internal/client"
)

func TestNotificationList(t *testing.T) {
	t.Run("returns list of notifications", func(t *testing.T) {
		mock := NewMockClient()
		mock.GetWithPaginationResponse = &client.APIResponse{
			StatusCode: 200,
			Data: []interface{}{
				map[string]interface{}{"id": "1", "message": "You have a notification"},
			},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		RunTestCommand(func() {
			notificationListCmd.Run(notificationListCmd, []string{})
		})

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		if mock.GetWithPaginationCalls[0].Path != "/notifications.json" {
			t.Errorf("expected path '/notifications.json', got '%s'", mock.GetWithPaginationCalls[0].Path)
		}
	})
}

func TestNotificationRead(t *testing.T) {
	t.Run("marks notification as read", func(t *testing.T) {
		mock := NewMockClient()
		mock.PostResponse = &client.APIResponse{
			StatusCode: 200,
			Data:       map[string]interface{}{},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		RunTestCommand(func() {
			notificationReadCmd.Run(notificationReadCmd, []string{"notif-1"})
		})

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		if mock.PostCalls[0].Path != "/notifications/notif-1/reading.json" {
			t.Errorf("expected path '/notifications/notif-1/reading.json', got '%s'", mock.PostCalls[0].Path)
		}
	})
}

func TestNotificationUnread(t *testing.T) {
	t.Run("marks notification as unread", func(t *testing.T) {
		mock := NewMockClient()
		mock.DeleteResponse = &client.APIResponse{
			StatusCode: 200,
			Data:       map[string]interface{}{},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		RunTestCommand(func() {
			notificationUnreadCmd.Run(notificationUnreadCmd, []string{"notif-1"})
		})

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		if mock.DeleteCalls[0].Path != "/notifications/notif-1/reading.json" {
			t.Errorf("expected path '/notifications/notif-1/reading.json', got '%s'", mock.DeleteCalls[0].Path)
		}
	})
}

func TestNotificationReadAll(t *testing.T) {
	t.Run("marks all notifications as read", func(t *testing.T) {
		mock := NewMockClient()
		mock.PostResponse = &client.APIResponse{
			StatusCode: 200,
			Data:       map[string]interface{}{},
		}

		result := SetTestMode(mock)
		SetTestConfig("token", "account", "https://api.example.com")
		defer ResetTestMode()

		RunTestCommand(func() {
			notificationReadAllCmd.Run(notificationReadAllCmd, []string{})
		})

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		if mock.PostCalls[0].Path != "/notifications/bulk_reading.json" {
			t.Errorf("expected path '/notifications/bulk_reading.json', got '%s'", mock.PostCalls[0].Path)
		}
	})
}
