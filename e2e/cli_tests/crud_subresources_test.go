package clitests

import (
	"strconv"
	"testing"
	"time"
)

func TestColumnCRUDAndMove(t *testing.T) {
	h := newHarness(t)
	boardID := createBoard(t, h)
	leftID := createColumn(t, h, boardID, "Left")
	rightID := createColumn(t, h, boardID, "Right")

	list := h.Run("column", "list", "--board", boardID)
	assertOK(t, list)
	if list.GetDataArray() == nil {
		t.Fatal("expected array response")
	}

	assertOK(t, h.Run("column", "show", leftID, "--board", boardID))
	assertOK(t, h.Run("column", "update", leftID, "--board", boardID, "--name", "Renamed Left"))
	assertOK(t, h.Run("column", "move-right", leftID))
	assertOK(t, h.Run("column", "move-left", rightID))
	assertOK(t, h.Run("column", "delete", rightID, "--board", boardID))
}

func TestCommentCRUD(t *testing.T) {
	h := newHarness(t)
	cardNum := fixture.CardNumber
	cardStr := strconv.Itoa(cardNum)

	list := h.Run("comment", "list", "--card", cardStr)
	assertOK(t, list)
	assertOK(t, h.Run("comment", "show", fixture.CommentID, "--card", cardStr))

	commentID := createComment(t, h, cardNum, "CLI comment")
	assertOK(t, h.Run("comment", "update", commentID, "--card", cardStr, "--body", "Updated CLI comment"))
	assertOK(t, h.Run("comment", "delete", commentID, "--card", cardStr))
}

func TestStepCRUD(t *testing.T) {
	h := newHarness(t)
	cardStr := strconv.Itoa(fixture.CardNumber)
	list := h.Run("step", "list", "--card", cardStr)
	assertOK(t, list)
	assertOK(t, h.Run("step", "show", fixture.StepID, "--card", cardStr))

	stepID := createStep(t, h, fixture.CardNumber, "CLI step")
	assertOK(t, h.Run("step", "update", stepID, "--card", cardStr, "--content", "Updated CLI step"))
	assertOK(t, h.Run("step", "delete", stepID, "--card", cardStr))
}

func TestReactionCRUD(t *testing.T) {
	h := newHarness(t)
	cardStr := strconv.Itoa(fixture.CardNumber)
	assertOK(t, h.Run("reaction", "list", "--card", cardStr))
	assertOK(t, h.Run("reaction", "list", "--card", cardStr, "--comment", fixture.CommentID))

	cardReaction := h.Run("reaction", "create", "--card", cardStr, "--content", "+1")
	assertOK(t, cardReaction)
	cardReactionID := cardReaction.GetIDFromLocation()
	if cardReactionID == "" {
		cardReactionID = cardReaction.GetDataString("id")
	}
	if cardReactionID == "" {
		t.Fatal("no reaction ID in create response")
	}
	t.Cleanup(func() { newHarness(t).Run("reaction", "delete", cardReactionID, "--card", cardStr) })
	assertOK(t, h.Run("reaction", "delete", cardReactionID, "--card", cardStr))

	commentReaction := h.Run("reaction", "create", "--card", cardStr, "--comment", fixture.CommentID, "--content", "heart")
	assertOK(t, commentReaction)
	commentReactionID := commentReaction.GetIDFromLocation()
	if commentReactionID == "" {
		commentReactionID = commentReaction.GetDataString("id")
	}
	if commentReactionID == "" {
		t.Fatal("no comment reaction ID in create response")
	}
	t.Cleanup(func() {
		newHarness(t).Run("reaction", "delete", commentReactionID, "--card", cardStr, "--comment", fixture.CommentID)
	})
	assertOK(t, h.Run("reaction", "delete", commentReactionID, "--card", cardStr, "--comment", fixture.CommentID))
}

func TestNotificationCommands(t *testing.T) {
	h := newHarness(t)
	assertOK(t, h.Run("notification", "list"))
	assertOK(t, h.Run("notification", "tray"))
	assertOK(t, h.Run("notification", "settings-show"))
	assertOK(t, h.Run("notification", "read-all"))

	show := h.Run("notification", "settings-show")
	assertOK(t, show)
	currentFreq := show.GetDataString("bundle_email_frequency")
	if currentFreq == "" {
		currentFreq = "never"
	}
	assertOK(t, h.Run("notification", "settings-update", "--bundle-email-frequency", currentFreq))

	id := notificationID(t, h)
	assertOK(t, h.Run("notification", "read", id))
	assertOK(t, h.Run("notification", "unread", id))
}

func TestTagAndPinLists(t *testing.T) {
	h := newHarness(t)
	tagResult := h.Run("tag", "list")
	assertOK(t, tagResult)
	if tagResult.GetDataArray() == nil {
		t.Fatal("expected tag list array response")
	}

	pinResult := h.Run("pin", "list")
	assertOK(t, pinResult)
	if pinResult.GetDataArray() == nil {
		t.Fatal("expected pin list array response")
	}
}

func TestCommentAndStepCreationOnThrowawayCard(t *testing.T) {
	h := newHarness(t)
	cardNum := createCard(t, h, fixture.BoardID)
	commentID := createComment(t, h, cardNum, "Throwaway card comment "+strconv.FormatInt(time.Now().UnixNano(), 10))
	stepID := createStep(t, h, cardNum, "Throwaway card step "+strconv.FormatInt(time.Now().UnixNano(), 10))
	if commentID == "" || stepID == "" {
		t.Fatal("expected comment and step IDs")
	}
}
