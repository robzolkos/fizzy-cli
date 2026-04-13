package clitests

import (
	"fmt"
	"testing"
	"time"

	"github.com/basecamp/fizzy-cli/e2e/harness"
)

func TestBoardList(t *testing.T) {
	h := newHarness(t)
	result := h.Run("board", "list")
	assertOK(t, result)
	if result.GetDataArray() == nil {
		t.Fatal("expected array response")
	}
}

func TestBoardListAll(t *testing.T) {
	assertOK(t, newHarness(t).Run("board", "list", "--all"))
}

func TestBoardListPaginated(t *testing.T) {
	assertOK(t, newHarness(t).Run("board", "list", "--page", "1"))
}

func TestBoardShow(t *testing.T) {
	result := newHarness(t).Run("board", "show", fixture.BoardID)
	assertOK(t, result)
	if got := result.GetDataString("id"); got != fixture.BoardID {
		t.Fatalf("expected board id %q, got %q", fixture.BoardID, got)
	}
}

func TestBoardShowNotFound(t *testing.T) {
	assertResult(t, newHarness(t).Run("board", "show", "nonexistent-board-id-99999"), harness.ExitNotFound)
}

func TestBoardCreateUpdateDelete(t *testing.T) {
	h := newHarness(t)
	boardID := createBoard(t, h)
	result := h.Run("board", "update", boardID, "--name", fmt.Sprintf("Updated Board %d", time.Now().UnixNano()))
	assertOK(t, result)

	deleteResult := h.Run("board", "delete", boardID)
	assertOK(t, deleteResult)
	if !deleteResult.GetDataBool("deleted") {
		t.Fatal("expected deleted=true")
	}
}

func TestBoardPublishUnpublish(t *testing.T) {
	h := newHarness(t)
	boardID := createBoard(t, h)
	publish := h.Run("board", "publish", boardID)
	assertOK(t, publish)
	if publish.GetDataString("public_url") == "" {
		t.Fatal("expected public_url in publish response")
	}
	assertOK(t, h.Run("board", "unpublish", boardID))
}

func TestBoardEntropy(t *testing.T) {
	h := newHarness(t)
	boardID := createBoard(t, h)
	assertOK(t, h.Run("board", "entropy", boardID, "--auto_postpone_period_in_days", "7"))
}

func TestBoardViews(t *testing.T) {
	h := newHarness(t)
	assertOK(t, h.Run("board", "closed", "--board", fixture.BoardID))
	assertOK(t, h.Run("board", "postponed", "--board", fixture.BoardID))
	assertOK(t, h.Run("board", "stream", "--board", fixture.BoardID))
}

func TestBoardInvolvement(t *testing.T) {
	h := newHarness(t)
	boardID := createBoard(t, h)
	assertOK(t, h.Run("board", "involvement", boardID, "--involvement", "watching"))
	assertOK(t, h.Run("board", "involvement", boardID, "--involvement", "access_only"))
}
