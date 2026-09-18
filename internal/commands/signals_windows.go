//go:build windows

package commands

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func handleSignals(args []string, ctx *Context) bool {
	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	fmt.Println("Listening for signals (Ctrl+C to trigger SIGINT)...")
	fmt.Println("Note: Windows supports SIGINT and SIGTERM only")

	sig := <-sigChan

	switch sig {
	case syscall.SIGINT:
		fmt.Println("\nReceived SIGINT (Ctrl+C)")
	case syscall.SIGTERM:
		fmt.Println("\nReceived SIGTERM")
	default:
		fmt.Println("\nReceived signal:", sig)
	}

	return true
}

func handleExitCode(args []string, ctx *Context) bool {
	if len(args) < 1 {
		fmt.Println("Usage: exitcode <code>")
		fmt.Println("Example: exitcode 0")
		return false
	}

	var code int
	fmt.Sscanf(args[0], "%d", &code)
	fmt.Printf("Exiting with code: %d\n", code)
	os.Exit(code)
	return true
}