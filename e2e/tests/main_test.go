package tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/robzolkos/fizzy-cli/e2e/harness"
)

func TestMain(m *testing.M) {
	cfg := harness.LoadConfig()

	if cfg.BinaryPath == "" || !fileExists(cfg.BinaryPath) {
		repoRoot, err := harness.RepoRoot()
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		tmpDir, err := os.MkdirTemp("", "fizzy-e2e-*")
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		binPath := filepath.Join(tmpDir, "fizzy")
		cmd := exec.Command("go", "build", "-o", binPath, "./cmd/fizzy")
		cmd.Dir = repoRoot
		if out, err := cmd.CombinedOutput(); err != nil {
			_ = os.RemoveAll(tmpDir)
			fmt.Fprintf(os.Stderr, "failed to build e2e binary: %v\n%s\n", err, string(out))
			os.Exit(1)
		}

		_ = os.Setenv("FIZZY_TEST_BINARY", binPath)
		cfg.BinaryPath = binPath

		code := m.Run()
		_ = os.RemoveAll(tmpDir)
		os.Exit(code)
	}

	os.Exit(m.Run())
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !st.IsDir()
}
