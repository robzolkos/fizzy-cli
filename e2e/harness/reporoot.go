package harness

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// RepoRoot returns the repository root (directory containing go.mod).
func RepoRoot() (string, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to determine caller location")
	}

	dir := filepath.Dir(thisFile)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("could not find repository root (go.mod not found)")
}
