package harness

import (
	"fmt"
	"strconv"
)

// CleanupTracker tracks created resources for cleanup after tests.
type CleanupTracker struct {
	// Boards to delete (by ID)
	Boards []string

	// Cards to delete (by number)
	Cards []int

	// Columns to delete (by ID, with board ID)
	Columns []ColumnRef

	// Comments to delete (by ID, with card number)
	Comments []CommentRef

	// Steps to delete (by ID, with card number)
	Steps []StepRef

	// Reactions to delete (by ID, with card number and comment ID)
	Reactions []ReactionRef
}

// ColumnRef references a column for cleanup.
type ColumnRef struct {
	ID      string
	BoardID string
}

// CommentRef references a comment for cleanup.
type CommentRef struct {
	ID         string
	CardNumber int
}

// StepRef references a step for cleanup.
type StepRef struct {
	ID         string
	CardNumber int
}

// ReactionRef references a reaction for cleanup.
type ReactionRef struct {
	ID         string
	CardNumber int
	CommentID  string
}

// NewCleanupTracker creates a new cleanup tracker.
func NewCleanupTracker() *CleanupTracker {
	return &CleanupTracker{
		Boards:    make([]string, 0),
		Cards:     make([]int, 0),
		Columns:   make([]ColumnRef, 0),
		Comments:  make([]CommentRef, 0),
		Steps:     make([]StepRef, 0),
		Reactions: make([]ReactionRef, 0),
	}
}

// AddBoard adds a board to the cleanup list.
func (c *CleanupTracker) AddBoard(id string) {
	c.Boards = append(c.Boards, id)
}

// AddCard adds a card to the cleanup list.
func (c *CleanupTracker) AddCard(number int) {
	c.Cards = append(c.Cards, number)
}

// AddColumn adds a column to the cleanup list.
func (c *CleanupTracker) AddColumn(id, boardID string) {
	c.Columns = append(c.Columns, ColumnRef{ID: id, BoardID: boardID})
}

// AddComment adds a comment to the cleanup list.
func (c *CleanupTracker) AddComment(id string, cardNumber int) {
	c.Comments = append(c.Comments, CommentRef{ID: id, CardNumber: cardNumber})
}

// AddStep adds a step to the cleanup list.
func (c *CleanupTracker) AddStep(id string, cardNumber int) {
	c.Steps = append(c.Steps, StepRef{ID: id, CardNumber: cardNumber})
}

// AddReaction adds a reaction to the cleanup list.
func (c *CleanupTracker) AddReaction(id string, cardNumber int, commentID string) {
	c.Reactions = append(c.Reactions, ReactionRef{ID: id, CardNumber: cardNumber, CommentID: commentID})
}

// CleanupAll deletes all tracked resources in reverse dependency order.
// It uses the provided harness to execute delete commands.
func (c *CleanupTracker) CleanupAll(h *Harness) []error {
	var errors []error

	// Delete in reverse dependency order:
	// 1. Reactions (depend on comments)
	for i := len(c.Reactions) - 1; i >= 0; i-- {
		ref := c.Reactions[i]
		result := h.Run("reaction", "delete", ref.ID,
			"--card", strconv.Itoa(ref.CardNumber),
			"--comment", ref.CommentID)
		if result.ExitCode != 0 && result.ExitCode != ExitNotFound {
			errors = append(errors, fmt.Errorf("failed to delete reaction %s: exit %d", ref.ID, result.ExitCode))
		}
	}

	// 2. Comments (depend on cards)
	for i := len(c.Comments) - 1; i >= 0; i-- {
		ref := c.Comments[i]
		result := h.Run("comment", "delete", ref.ID,
			"--card", strconv.Itoa(ref.CardNumber))
		if result.ExitCode != 0 && result.ExitCode != ExitNotFound {
			errors = append(errors, fmt.Errorf("failed to delete comment %s: exit %d", ref.ID, result.ExitCode))
		}
	}

	// 3. Steps (depend on cards)
	for i := len(c.Steps) - 1; i >= 0; i-- {
		ref := c.Steps[i]
		result := h.Run("step", "delete", ref.ID,
			"--card", strconv.Itoa(ref.CardNumber))
		if result.ExitCode != 0 && result.ExitCode != ExitNotFound {
			errors = append(errors, fmt.Errorf("failed to delete step %s: exit %d", ref.ID, result.ExitCode))
		}
	}

	// 4. Cards (depend on boards)
	for i := len(c.Cards) - 1; i >= 0; i-- {
		number := c.Cards[i]
		result := h.Run("card", "delete", strconv.Itoa(number))
		if result.ExitCode != 0 && result.ExitCode != ExitNotFound {
			errors = append(errors, fmt.Errorf("failed to delete card %d: exit %d", number, result.ExitCode))
		}
	}

	// 5. Columns (depend on boards)
	for i := len(c.Columns) - 1; i >= 0; i-- {
		ref := c.Columns[i]
		result := h.Run("column", "delete", ref.ID,
			"--board", ref.BoardID)
		if result.ExitCode != 0 && result.ExitCode != ExitNotFound {
			errors = append(errors, fmt.Errorf("failed to delete column %s: exit %d", ref.ID, result.ExitCode))
		}
	}

	// 6. Boards (no dependencies)
	for i := len(c.Boards) - 1; i >= 0; i-- {
		id := c.Boards[i]
		result := h.Run("board", "delete", id)
		if result.ExitCode != 0 && result.ExitCode != ExitNotFound {
			errors = append(errors, fmt.Errorf("failed to delete board %s: exit %d", id, result.ExitCode))
		}
	}

	return errors
}
