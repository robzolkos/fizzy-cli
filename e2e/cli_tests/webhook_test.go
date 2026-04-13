package clitests

import (
	"strconv"
	"testing"
	"time"
)

func TestWebhookCRUD(t *testing.T) {
	h := newHarness(t)
	boardID := createBoard(t, h)
	name := "CLI Test Hook " + strconv.FormatInt(time.Now().UnixNano(), 10)

	create := h.Run("webhook", "create",
		"--board", boardID,
		"--name", name,
		"--url", "https://example.com/fizzy-cli-webhook",
	)
	assertOK(t, create)
	webhookID := create.GetIDFromLocation()
	if webhookID == "" {
		webhookID = create.GetDataString("id")
	}
	if webhookID == "" {
		t.Fatal("no webhook ID in create response")
	}
	deleted := false
	t.Cleanup(func() {
		if !deleted {
			newHarness(t).Run("webhook", "delete", "--board", boardID, webhookID)
		}
	})

	show := h.Run("webhook", "show", "--board", boardID, webhookID)
	assertOK(t, show)
	if got := show.GetDataString("name"); got != name {
		t.Fatalf("expected webhook name %q, got %q", name, got)
	}
	if got := show.GetDataString("payload_url"); got == "" {
		t.Fatal("expected payload_url in webhook show response")
	}

	list := h.Run("webhook", "list", "--board", boardID)
	assertOK(t, list)
	found := false
	for _, item := range list.GetDataArray() {
		if mapValueString(asMap(item), "id") == webhookID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected webhook list to include %q", webhookID)
	}

	updatedName := name + " Updated"
	update := h.Run("webhook", "update", "--board", boardID, webhookID, "--name", updatedName, "--actions", "card_closed")
	assertOK(t, update)

	showUpdated := h.Run("webhook", "show", "--board", boardID, webhookID)
	assertOK(t, showUpdated)
	if got := showUpdated.GetDataString("name"); got != updatedName {
		t.Fatalf("expected updated webhook name %q, got %q", updatedName, got)
	}
	actions := asSlice(showUpdated.GetDataMap()["subscribed_actions"])
	if len(actions) != 1 || stringifyID(actions[0]) != "card_closed" {
		t.Fatalf("expected subscribed_actions [card_closed], got %v", actions)
	}

	deleteResult := h.Run("webhook", "delete", "--board", boardID, webhookID)
	assertOK(t, deleteResult)
	deleted = true
	if !deleteResult.GetDataBool("deleted") {
		t.Fatal("expected deleted=true")
	}
}
