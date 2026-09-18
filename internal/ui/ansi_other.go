//go:build !windows

package ui

func EnableANSI() {
	// ANSI escape sequences work natively on
	// Linux and macOS terminals.
}