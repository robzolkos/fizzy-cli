package commands

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/basecamp/cli/surface"
	"github.com/spf13/cobra"
)

func TestSurfaceSnapshot(t *testing.T) {
	// Ensure Cobra has initialized all default help flags and commands
	initAllHelpFlags(rootCmd)

	// Generate fresh snapshot from the command tree
	fresh := surface.SnapshotString(rootCmd)

	// Read golden file
	_, thisFile, _, _ := runtime.Caller(0)
	goldenPath := filepath.Join(filepath.Dir(thisFile), "..", "..", "SURFACE.txt")
	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("Failed to read SURFACE.txt (run 'make surface-snapshot' to generate): %v", err)
	}

	goldenStr := strings.TrimSpace(string(golden))
	freshStr := strings.TrimSpace(fresh)

	if goldenStr != freshStr {
		// Show the diff for debugging
		oldEntries := surface.Snapshot(rootCmd)
		goldenEntries := parseSurfaceEntries(goldenStr)
		diff := surface.Diff(goldenEntries, oldEntries)

		var sb strings.Builder
		sb.WriteString("SURFACE.txt is out of date. Run 'make surface-snapshot' to regenerate.\n\n")
		if len(diff.Added) > 0 {
			sb.WriteString("Added:\n")
			for _, e := range diff.Added {
				sb.WriteString("  + " + e.String() + "\n")
			}
		}
		if len(diff.Removed) > 0 {
			sb.WriteString("Removed (BREAKING):\n")
			for _, e := range diff.Removed {
				sb.WriteString("  - " + e.String() + "\n")
			}
		}
		t.Fatal(sb.String())
	}
}

// initAllHelpFlags recursively initializes help flags on all commands.
// Cobra adds --help lazily; this ensures deterministic snapshots.
func initAllHelpFlags(cmd *cobra.Command) {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()
	for _, sub := range cmd.Commands() {
		initAllHelpFlags(sub)
	}
}

// parseSurfaceEntries parses a surface snapshot string back into entries for diffing.
func parseSurfaceEntries(s string) []surface.Entry {
	var entries []surface.Entry
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}
		kind := surface.EntryKind(parts[0])
		rest := parts[1]

		switch kind {
		case surface.KindCmd:
			entries = append(entries, surface.Entry{Kind: kind, Path: rest})
		case surface.KindFlag:
			// FLAG path --name type=flagtype
			flagParts := strings.Fields(rest)
			if len(flagParts) >= 3 {
				path := flagParts[0]
				name := strings.TrimPrefix(flagParts[1], "--")
				flagType := strings.TrimPrefix(flagParts[2], "type=")
				// Path may be multi-word — find the -- prefix
				for i, p := range flagParts {
					if strings.HasPrefix(p, "--") {
						path = strings.Join(flagParts[:i], " ")
						name = strings.TrimPrefix(flagParts[i], "--")
						if i+1 < len(flagParts) {
							flagType = strings.TrimPrefix(flagParts[i+1], "type=")
						}
						break
					}
				}
				entries = append(entries, surface.Entry{Kind: kind, Path: path, Name: name, FlagType: flagType})
			}
		case surface.KindSub:
			subParts := strings.Fields(rest)
			if len(subParts) >= 2 {
				name := subParts[len(subParts)-1]
				path := strings.Join(subParts[:len(subParts)-1], " ")
				entries = append(entries, surface.Entry{Kind: kind, Path: path, Name: name})
			}
		}
	}
	return entries
}
