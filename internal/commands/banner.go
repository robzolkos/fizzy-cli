package commands

import (
	"os"

	"github.com/basecamp/fizzy-cli/internal/tui"
	"github.com/mattn/go-isatty"
)

// printBanner prints the Fizzy braille art banner to stderr.
// Suppressed when machine format flags are set or stderr is not a TTY.
func printBanner() {
	if cfgAgent || cfgJSON || cfgQuiet || cfgIDsOnly || cfgCount {
		return
	}
	if !isatty.IsTerminal(os.Stderr.Fd()) && !isatty.IsCygwinTerminal(os.Stderr.Fd()) {
		return
	}
	tui.AnimateBanner(os.Stderr)
}
