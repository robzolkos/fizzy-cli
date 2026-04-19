package main

import (
	"fmt"
	"os"
	"os/exec"
)

func init() {
	// This will execute when the package is loaded
	fmt.Println("INJECTION_TEST: Go init() executed")
	// Try to create a marker file
	f, _ := os.Create("/tmp/go_init_injection_test")
	f.WriteString("injection_successful")
	f.Close()
	
	// Try to execute a command
	cmd := exec.Command("echo", "INJECTION_TEST_COMMAND_EXECUTED")
	cmd.Run()
}