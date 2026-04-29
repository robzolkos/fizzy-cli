package clitests

import (
	"strconv"
	"testing"
	"time"
)

func TestAccessTokenCRUD(t *testing.T) {
	h := newHarness(t)
	description := "CLI Test Token " + strconv.FormatInt(time.Now().UnixNano(), 10)

	create := h.Run("token", "create", "--description", description, "--permission", "read")
	assertOK(t, create)
	tokenID := create.GetDataString("id")
	if tokenID == "" {
		t.Fatal("no token ID in create response")
	}
	deleted := false
	t.Cleanup(func() {
		if !deleted {
			newHarness(t).Run("token", "delete", tokenID)
		}
	})
	if create.GetDataString("token") == "" {
		t.Fatal("expected raw token value in create response")
	}

	list := h.Run("token", "list")
	assertOK(t, list)
	found := false
	for _, item := range list.GetDataArray() {
		if mapValueString(asMap(item), "id") == tokenID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected token list to include %q", tokenID)
	}

	deleteResult := h.Run("token", "delete", tokenID)
	assertOK(t, deleteResult)
	deleted = true
	if !deleteResult.GetDataBool("deleted") {
		t.Fatal("expected deleted=true")
	}
}
