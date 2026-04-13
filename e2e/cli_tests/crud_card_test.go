package clitests

import (
	"strconv"
	"testing"
	"time"

	"github.com/basecamp/fizzy-cli/e2e/harness"
)

func TestCardList(t *testing.T) {
	assertOK(t, newHarness(t).Run("card", "list"))
}

func TestCardListOnBoard(t *testing.T) {
	result := newHarness(t).Run("card", "list", "--board", fixture.BoardID)
	assertOK(t, result)
	if result.GetDataArray() == nil {
		t.Fatal("expected array response")
	}
}

func TestCardListAll(t *testing.T) {
	assertOK(t, newHarness(t).Run("card", "list", "--board", fixture.BoardID, "--all"))
}

func TestCardShow(t *testing.T) {
	assertOK(t, newHarness(t).Run("card", "show", strconv.Itoa(fixture.CardNumber)))
}

func TestCardShowNotFound(t *testing.T) {
	assertResult(t, newHarness(t).Run("card", "show", "999999999"), harness.ExitNotFound)
}

func TestCardLifecycle(t *testing.T) {
	h := newHarness(t)
	num := createCard(t, h, fixture.BoardID)
	numStr := strconv.Itoa(num)

	assertOK(t, h.Run("card", "update", numStr, "--title", "Updated Card"))
	assertOK(t, h.Run("card", "column", numStr, "--column", fixture.ColumnID))
	assertOK(t, h.Run("card", "watch", numStr))
	assertOK(t, h.Run("card", "unwatch", numStr))
	assertOK(t, h.Run("card", "mark-read", numStr))
	assertOK(t, h.Run("card", "mark-unread", numStr))
	assertOK(t, h.Run("card", "pin", numStr))
	assertOK(t, h.Run("card", "unpin", numStr))
	assertOK(t, h.Run("card", "golden", numStr))
	assertOK(t, h.Run("card", "ungolden", numStr))
	assertOK(t, h.Run("card", "tag", numStr, "--tag", "cli-test"))
	assertOK(t, h.Run("card", "self-assign", numStr))
	assertOK(t, h.Run("card", "close", numStr))
	assertOK(t, h.Run("card", "reopen", numStr))
	assertOK(t, h.Run("card", "postpone", numStr))
	assertOK(t, h.Run("card", "untriage", numStr))
}

func TestCardAssignToCurrentUser(t *testing.T) {
	h := newHarness(t)
	num := createCard(t, h, fixture.BoardID)
	userID := currentUserID(t, h)

	assertOK(t, h.Run("card", "assign", strconv.Itoa(num), "--user", userID))

	show := h.Run("card", "show", strconv.Itoa(num))
	assertOK(t, show)
	assignees := asSlice(show.GetDataMap()["assignees"])
	if len(assignees) == 0 {
		t.Fatal("expected assigned card to include assignees")
	}
	for _, item := range assignees {
		if mapValueString(asMap(item), "id") == userID {
			return
		}
	}
	t.Fatalf("expected assignees to include user %q", userID)
}

func TestCardImageRemove(t *testing.T) {
	h := newHarness(t)
	boardID := createBoard(t, h)
	ref := uploadFixture(t, h, "test_image.png")
	result := h.Run("card", "create",
		"--board", boardID,
		"--title", "Image Remove Card "+strconv.FormatInt(time.Now().UnixNano(), 10),
		"--image", ref.SignedID,
	)
	assertOK(t, result)
	num := result.GetNumberFromLocation()
	if num == 0 {
		num = result.GetDataInt("number")
	}
	if num == 0 {
		t.Fatal("no card number in create response")
	}
	t.Cleanup(func() { newHarness(t).Run("card", "delete", strconv.Itoa(num)) })

	showWithImage := h.Run("card", "show", strconv.Itoa(num))
	assertOK(t, showWithImage)
	if got := showWithImage.GetDataString("image_url"); got == "" {
		t.Fatal("expected image_url before image removal")
	}

	assertOK(t, h.Run("card", "image-remove", strconv.Itoa(num)))

	showWithoutImage := h.Run("card", "show", strconv.Itoa(num))
	assertOK(t, showWithoutImage)
	if got := showWithoutImage.GetDataString("image_url"); got != "" {
		t.Fatalf("expected image_url to be cleared, got %q", got)
	}
}

func TestCardMoveBetweenBoards(t *testing.T) {
	h := newHarness(t)
	destinationBoardID := createBoard(t, h)
	num := createCard(t, h, fixture.BoardID)
	assertOK(t, h.Run("card", "move", strconv.Itoa(num), "--to", destinationBoardID))
}

func TestCardAttachmentsShow(t *testing.T) {
	assertOK(t, newHarness(t).Run("card", "attachments", "show", strconv.Itoa(fixture.CardNumber)))
}

func TestCardDelete(t *testing.T) {
	h := newHarness(t)
	num := createCard(t, h, fixture.BoardID)
	assertOK(t, h.Run("card", "delete", strconv.Itoa(num)))
}

func TestCardCreateRoundTrip(t *testing.T) {
	h := newHarness(t)
	result := h.Run("card", "create", "--board", fixture.BoardID, "--title", "Round Trip Card", "--description", "created by cli tests")
	assertOK(t, result)
	num := result.GetNumberFromLocation()
	if num == 0 {
		num = result.GetDataInt("number")
	}
	if num == 0 {
		t.Fatal("no card number in create response")
	}
	t.Cleanup(func() { newHarness(t).Run("card", "delete", strconv.Itoa(num)) })
	show := h.Run("card", "show", strconv.Itoa(num))
	assertOK(t, show)
	if title := show.GetDataString("title"); title == "" {
		t.Fatal("expected title in card show response")
	}
}

func TestCardCreateWithUniqueTitle(t *testing.T) {
	h := newHarness(t)
	title := "Unique Card " + strconv.FormatInt(time.Now().UnixNano(), 10)
	result := h.Run("card", "create", "--board", fixture.BoardID, "--title", title)
	assertOK(t, result)
	num := result.GetNumberFromLocation()
	if num == 0 {
		num = result.GetDataInt("number")
	}
	if num == 0 {
		t.Fatal("no card number in create response")
	}
	t.Cleanup(func() { newHarness(t).Run("card", "delete", strconv.Itoa(num)) })
}
