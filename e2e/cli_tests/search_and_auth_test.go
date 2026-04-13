package clitests

import (
	"testing"

	"github.com/basecamp/fizzy-cli/e2e/harness"
)

func TestSearch(t *testing.T) {
	h := newHarness(t)
	assertOK(t, h.Run("search", "test"))
	assertOK(t, h.Run("search", "test", "--board", fixture.BoardID))
	assertOK(t, h.Run("search", "test", "--all"))
}

func TestAuthInvalidToken(t *testing.T) {
	badCfg := *cfg
	badCfg.Token = "fizzy_invalid_token"
	h := harness.NewWithConfig(t, &badCfg)
	assertResult(t, h.Run("board", "list"), harness.ExitAuth)
}

func TestAuthMissingToken(t *testing.T) {
	missingCfg := *cfg
	missingCfg.Token = ""
	h := harness.NewWithConfig(t, &missingCfg)
	assertResult(t, h.Run("board", "list"), harness.ExitAuth)
}
