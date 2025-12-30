package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/robzolkos/fizzy-cli/e2e/harness"
)

// createTestBoard creates a board for card tests and adds it to cleanup
func createTestBoard(t *testing.T, h *harness.Harness) string {
	t.Helper()
	name := fmt.Sprintf("Card Test Board %d", time.Now().UnixNano())
	result := h.Run("board", "create", "--name", name)
	if result.ExitCode != harness.ExitSuccess {
		t.Fatalf("failed to create test board: %s\nstdout: %s", result.Stderr, result.Stdout)
	}
	// Create returns location, not data - extract ID from location
	boardID := result.GetIDFromLocation()
	if boardID == "" {
		// Try data.id as fallback
		boardID = result.GetDataString("id")
	}
	if boardID == "" {
		t.Fatalf("no board ID returned (location: %s)", result.GetLocation())
	}
	h.Cleanup.AddBoard(boardID)
	return boardID
}

func TestCardList(t *testing.T) {
	h := harness.New(t)

	t.Run("returns list of cards", func(t *testing.T) {
		result := h.Run("card", "list")

		if result.ExitCode != harness.ExitSuccess {
			t.Errorf("expected exit code %d, got %d\nstderr: %s", harness.ExitSuccess, result.ExitCode, result.Stderr)
		}

		if result.Response == nil {
			t.Fatalf("expected JSON response, got nil\nstdout: %s", result.Stdout)
		}

		if !result.Response.Success {
			t.Error("expected success=true")
		}

		// Data should be an array
		arr := result.GetDataArray()
		if arr == nil {
			t.Error("expected data to be an array")
		}
	})

	t.Run("supports --page option", func(t *testing.T) {
		result := h.Run("card", "list", "--page", "1")

		if result.ExitCode != harness.ExitSuccess {
			t.Errorf("expected exit code %d, got %d", harness.ExitSuccess, result.ExitCode)
		}

		if result.Response == nil {
			t.Fatal("expected JSON response")
		}

		if !result.Response.Success {
			t.Error("expected success=true")
		}
	})

	t.Run("supports --all flag", func(t *testing.T) {
		result := h.Run("card", "list", "--all")

		if result.ExitCode != harness.ExitSuccess {
			t.Errorf("expected exit code %d, got %d", harness.ExitSuccess, result.ExitCode)
		}
	})
}

func TestCardListWithFilters(t *testing.T) {
	h := harness.New(t)
	defer h.Cleanup.CleanupAll(h)

	boardID := createTestBoard(t, h)

	t.Run("filters by board", func(t *testing.T) {
		result := h.Run("card", "list", "--board", boardID)

		if result.ExitCode != harness.ExitSuccess {
			t.Errorf("expected exit code %d, got %d", harness.ExitSuccess, result.ExitCode)
		}

		if !result.Response.Success {
			t.Error("expected success=true")
		}
	})

	t.Run("filters by status", func(t *testing.T) {
		result := h.Run("card", "list", "--status", "published")

		if result.ExitCode != harness.ExitSuccess {
			t.Errorf("expected exit code %d, got %d", harness.ExitSuccess, result.ExitCode)
		}
	})
}

func TestCardCRUD(t *testing.T) {
	h := harness.New(t)
	defer h.Cleanup.CleanupAll(h)

	boardID := createTestBoard(t, h)
	var cardNumber int
	cardTitle := fmt.Sprintf("Test Card %d", time.Now().UnixNano())

	t.Run("create card with title", func(t *testing.T) {
		result := h.Run("card", "create", "--board", boardID, "--title", cardTitle)

		if result.ExitCode != harness.ExitSuccess {
			t.Fatalf("expected exit code %d, got %d\nstderr: %s\nstdout: %s",
				harness.ExitSuccess, result.ExitCode, result.Stderr, result.Stdout)
		}

		if result.Response == nil {
			t.Fatalf("expected JSON response, got nil\nstdout: %s", result.Stdout)
		}

		if !result.Response.Success {
			t.Errorf("expected success=true, error: %+v", result.Response.Error)
		}

		// Create returns location - extract number from it
		cardNumber = result.GetNumberFromLocation()
		if cardNumber == 0 {
			// Try data.number as fallback
			cardNumber = result.GetDataInt("number")
		}
		if cardNumber == 0 {
			t.Fatalf("expected card number in response (location: %s)", result.GetLocation())
		}

		h.Cleanup.AddCard(cardNumber)

		// Note: Create returns location, not data with title
		// Verify the card exists via show command
		showResult := h.Run("card", "show", strconv.Itoa(cardNumber))
		if showResult.ExitCode == harness.ExitSuccess {
			title := showResult.GetDataString("title")
			if title != cardTitle {
				t.Errorf("expected title %q, got %q", cardTitle, title)
			}
		}
	})

	t.Run("show card by number", func(t *testing.T) {
		if cardNumber == 0 {
			t.Skip("no card number from create test")
		}

		result := h.Run("card", "show", strconv.Itoa(cardNumber))

		if result.ExitCode != harness.ExitSuccess {
			t.Errorf("expected exit code %d, got %d\nstderr: %s", harness.ExitSuccess, result.ExitCode, result.Stderr)
		}

		if result.Response == nil {
			t.Fatal("expected JSON response")
		}

		if !result.Response.Success {
			t.Error("expected success=true")
		}

		num := result.GetDataInt("number")
		if num != cardNumber {
			t.Errorf("expected number %d, got %d", cardNumber, num)
		}
	})

	t.Run("update card title", func(t *testing.T) {
		if cardNumber == 0 {
			t.Skip("no card number from create test")
		}

		newTitle := cardTitle + " Updated"
		result := h.Run("card", "update", strconv.Itoa(cardNumber), "--title", newTitle)

		if result.ExitCode != harness.ExitSuccess {
			t.Errorf("expected exit code %d, got %d\nstderr: %s", harness.ExitSuccess, result.ExitCode, result.Stderr)
		}

		if result.Response == nil {
			t.Fatal("expected JSON response")
		}

		if !result.Response.Success {
			t.Error("expected success=true")
		}

		title := result.GetDataString("title")
		if title != newTitle {
			t.Errorf("expected title %q, got %q", newTitle, title)
		}
	})

	t.Run("delete card", func(t *testing.T) {
		if cardNumber == 0 {
			t.Skip("no card number from create test")
		}

		result := h.Run("card", "delete", strconv.Itoa(cardNumber))

		if result.ExitCode != harness.ExitSuccess {
			t.Errorf("expected exit code %d, got %d\nstderr: %s", harness.ExitSuccess, result.ExitCode, result.Stderr)
		}

		if result.Response == nil {
			t.Fatal("expected JSON response")
		}

		if !result.Response.Success {
			t.Error("expected success=true")
		}

		deleted := result.GetDataBool("deleted")
		if !deleted {
			t.Error("expected deleted=true")
		}

		// Remove from cleanup since we deleted it
		h.Cleanup.Cards = nil
	})
}

func TestCardCreateWithDescription(t *testing.T) {
	h := harness.New(t)
	defer h.Cleanup.CleanupAll(h)

	boardID := createTestBoard(t, h)

	t.Run("create card with description", func(t *testing.T) {
		title := fmt.Sprintf("Card with Description %d", time.Now().UnixNano())
		description := "<p>This is a <strong>test</strong> description.</p>"

		result := h.Run("card", "create", "--board", boardID, "--title", title, "--description", description)

		if result.ExitCode != harness.ExitSuccess {
			t.Fatalf("expected exit code %d, got %d\nstderr: %s", harness.ExitSuccess, result.ExitCode, result.Stderr)
		}

		// Create returns location - extract number from it
		cardNumber := result.GetNumberFromLocation()
		if cardNumber == 0 {
			cardNumber = result.GetDataInt("number")
		}
		if cardNumber != 0 {
			h.Cleanup.AddCard(cardNumber)
		}

		if !result.Response.Success {
			t.Error("expected success=true")
		}
	})

	t.Run("create card with description_file", func(t *testing.T) {
		// Get the path to the test document fixture
		wd, _ := os.Getwd()
		fixturePath := filepath.Join(wd, "..", "testdata", "fixtures", "test_document.txt")

		// Check if fixture exists
		if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
			t.Skipf("test fixture not found at %s", fixturePath)
		}

		title := fmt.Sprintf("Card from File %d", time.Now().UnixNano())
		result := h.Run("card", "create", "--board", boardID, "--title", title, "--description_file", fixturePath)

		if result.ExitCode != harness.ExitSuccess {
			t.Fatalf("expected exit code %d, got %d\nstderr: %s", harness.ExitSuccess, result.ExitCode, result.Stderr)
		}

		// Create returns location - extract number from it
		cardNumber := result.GetNumberFromLocation()
		if cardNumber == 0 {
			cardNumber = result.GetDataInt("number")
		}
		if cardNumber != 0 {
			h.Cleanup.AddCard(cardNumber)
		}

		if !result.Response.Success {
			t.Error("expected success=true")
		}
	})
}

func TestCardActions(t *testing.T) {
	h := harness.New(t)
	defer h.Cleanup.CleanupAll(h)

	boardID := createTestBoard(t, h)

	// Create a card for action tests
	title := fmt.Sprintf("Action Test Card %d", time.Now().UnixNano())
	result := h.Run("card", "create", "--board", boardID, "--title", title)
	if result.ExitCode != harness.ExitSuccess {
		t.Fatalf("failed to create test card: %s\nstdout: %s", result.Stderr, result.Stdout)
	}
	// Create returns location - extract number from it
	cardNumber := result.GetNumberFromLocation()
	if cardNumber == 0 {
		cardNumber = result.GetDataInt("number")
	}
	if cardNumber == 0 {
		t.Fatalf("failed to get card number from create (location: %s)", result.GetLocation())
	}
	h.Cleanup.AddCard(cardNumber)
	cardStr := strconv.Itoa(cardNumber)

	t.Run("close card", func(t *testing.T) {
		result := h.Run("card", "close", cardStr)

		if result.ExitCode != harness.ExitSuccess {
			t.Errorf("expected exit code %d, got %d\nstderr: %s", harness.ExitSuccess, result.ExitCode, result.Stderr)
		}

		if !result.Response.Success {
			t.Errorf("expected success=true, error: %+v", result.Response.Error)
		}
	})

	t.Run("reopen card", func(t *testing.T) {
		result := h.Run("card", "reopen", cardStr)

		if result.ExitCode != harness.ExitSuccess {
			t.Errorf("expected exit code %d, got %d\nstderr: %s", harness.ExitSuccess, result.ExitCode, result.Stderr)
		}

		if !result.Response.Success {
			t.Errorf("expected success=true, error: %+v", result.Response.Error)
		}
	})

	t.Run("postpone card", func(t *testing.T) {
		result := h.Run("card", "postpone", cardStr)

		if result.ExitCode != harness.ExitSuccess {
			t.Errorf("expected exit code %d, got %d\nstderr: %s", harness.ExitSuccess, result.ExitCode, result.Stderr)
		}

		if !result.Response.Success {
			t.Errorf("expected success=true, error: %+v", result.Response.Error)
		}
	})

	t.Run("watch card", func(t *testing.T) {
		result := h.Run("card", "watch", cardStr)

		if result.ExitCode != harness.ExitSuccess {
			t.Errorf("expected exit code %d, got %d\nstderr: %s", harness.ExitSuccess, result.ExitCode, result.Stderr)
		}

		if !result.Response.Success {
			t.Errorf("expected success=true, error: %+v", result.Response.Error)
		}
	})

	t.Run("unwatch card", func(t *testing.T) {
		result := h.Run("card", "unwatch", cardStr)

		if result.ExitCode != harness.ExitSuccess {
			t.Errorf("expected exit code %d, got %d\nstderr: %s", harness.ExitSuccess, result.ExitCode, result.Stderr)
		}

		if !result.Response.Success {
			t.Errorf("expected success=true, error: %+v", result.Response.Error)
		}
	})

	t.Run("tag card", func(t *testing.T) {
		tagName := fmt.Sprintf("test-tag-%d", time.Now().UnixNano())
		result := h.Run("card", "tag", cardStr, "--tag", tagName)

		if result.ExitCode != harness.ExitSuccess {
			t.Errorf("expected exit code %d, got %d\nstderr: %s", harness.ExitSuccess, result.ExitCode, result.Stderr)
		}

		if !result.Response.Success {
			t.Errorf("expected success=true, error: %+v", result.Response.Error)
		}
	})
}

func TestCardColumn(t *testing.T) {
	h := harness.New(t)
	defer h.Cleanup.CleanupAll(h)

	boardID := createTestBoard(t, h)

	// Create a column
	columnName := fmt.Sprintf("Test Column %d", time.Now().UnixNano())
	colResult := h.Run("column", "create", "--board", boardID, "--name", columnName)
	if colResult.ExitCode != harness.ExitSuccess {
		t.Fatalf("failed to create column: %s\nstdout: %s", colResult.Stderr, colResult.Stdout)
	}
	// Create returns location - extract ID from it
	columnID := colResult.GetIDFromLocation()
	if columnID == "" {
		columnID = colResult.GetDataString("id")
	}
	if columnID == "" {
		t.Fatalf("failed to get column ID (location: %s)", colResult.GetLocation())
	}
	h.Cleanup.AddColumn(columnID, boardID)

	// Create a card
	title := fmt.Sprintf("Column Test Card %d", time.Now().UnixNano())
	cardResult := h.Run("card", "create", "--board", boardID, "--title", title)
	if cardResult.ExitCode != harness.ExitSuccess {
		t.Fatalf("failed to create card: %s\nstdout: %s", cardResult.Stderr, cardResult.Stdout)
	}
	// Create returns location - extract number from it
	cardNumber := cardResult.GetNumberFromLocation()
	if cardNumber == 0 {
		cardNumber = cardResult.GetDataInt("number")
	}
	if cardNumber == 0 {
		t.Fatalf("failed to get card number (location: %s)", cardResult.GetLocation())
	}
	h.Cleanup.AddCard(cardNumber)
	cardStr := strconv.Itoa(cardNumber)

	t.Run("move card to column", func(t *testing.T) {
		result := h.Run("card", "column", cardStr, "--column", columnID)

		if result.ExitCode != harness.ExitSuccess {
			t.Errorf("expected exit code %d, got %d\nstderr: %s", harness.ExitSuccess, result.ExitCode, result.Stderr)
		}

		if !result.Response.Success {
			t.Errorf("expected success=true, error: %+v", result.Response.Error)
		}
	})

	t.Run("untriage card (send back to triage)", func(t *testing.T) {
		result := h.Run("card", "untriage", cardStr)

		if result.ExitCode != harness.ExitSuccess {
			t.Errorf("expected exit code %d, got %d\nstderr: %s", harness.ExitSuccess, result.ExitCode, result.Stderr)
		}

		if !result.Response.Success {
			t.Errorf("expected success=true, error: %+v", result.Response.Error)
		}
	})
}

func TestCardShowNotFound(t *testing.T) {
	h := harness.New(t)

	t.Run("returns not found for non-existent card", func(t *testing.T) {
		result := h.Run("card", "show", "999999999")

		if result.ExitCode != harness.ExitNotFound {
			t.Errorf("expected exit code %d, got %d\nstdout: %s",
				harness.ExitNotFound, result.ExitCode, result.Stdout)
		}

		if result.Response == nil {
			t.Fatal("expected JSON response")
		}

		if result.Response.Success {
			t.Error("expected success=false")
		}

		if result.Response.Error == nil {
			t.Error("expected error in response")
		}
	})
}

func TestCardCreateMissingBoard(t *testing.T) {
	h := harness.New(t)

	t.Run("fails without required --board option", func(t *testing.T) {
		result := h.Run("card", "create", "--title", "Test")

		// Should fail with error exit code
		if result.ExitCode == harness.ExitSuccess {
			t.Error("expected non-zero exit code for missing required option")
		}
	})
}

func TestCardCreateMissingTitle(t *testing.T) {
	h := harness.New(t)
	defer h.Cleanup.CleanupAll(h)

	boardID := createTestBoard(t, h)

	t.Run("fails without required --title option", func(t *testing.T) {
		result := h.Run("card", "create", "--board", boardID)

		// Should fail with error exit code
		if result.ExitCode == harness.ExitSuccess {
			t.Error("expected non-zero exit code for missing required option")
		}
	})
}
