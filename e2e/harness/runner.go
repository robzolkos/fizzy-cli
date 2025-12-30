package harness

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"syscall"
)

// Execute runs the CLI binary with the given arguments and returns the result.
func Execute(binaryPath string, args []string, env map[string]string) *Result {
	cmd := exec.Command(binaryPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set up environment
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	err := cmd.Run()

	result := &Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: 0,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				result.ExitCode = status.ExitStatus()
			}
		} else {
			// Command failed to start
			result.ExitCode = -1
			result.Stderr = err.Error()
		}
	}

	// Try to parse JSON response
	if result.Stdout != "" {
		var resp Response
		if err := json.Unmarshal([]byte(result.Stdout), &resp); err != nil {
			result.ParseError = err
		} else {
			result.Response = &resp
		}
	}

	return result
}
