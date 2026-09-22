//go:build windows
// +build windows

package ui

import (
	"os"
	"runtime"
	"syscall"
	"unsafe"
)

// InitTerminal initializes terminal features.
//
// On Windows, ANSI escape sequences are not always enabled,
// especially when Zebra is launched from cmd.exe.
//
// This enables:
//
//   - colors
//   - bold text
//   - ANSI cursor controls
//   - other VT terminal features
//
// On Linux and macOS nothing needs to be done.
func InitTerminal() {
	if runtime.GOOS != "windows" {
		return
	}

	enableWindowsANSI(os.Stdout)
	enableWindowsANSI(os.Stderr)
}

// enableWindowsANSI enables Windows Virtual Terminal
// Processing for the supplied console handle.
func enableWindowsANSI(file *os.File) {
	if file == nil {
		return
	}

	kernel32 := syscall.NewLazyDLL("kernel32.dll")

	getConsoleMode := kernel32.NewProc("GetConsoleMode")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")

	handle := uintptr(file.Fd())

	var mode uint32

	ret, _, _ := getConsoleMode.Call(
		handle,
		uintptr(unsafe.Pointer(&mode)),
	)

	// GetConsoleMode failed.
	//
	// This can happen when stdout/stderr is redirected
	// to a file or pipe instead of a real console.
	if ret == 0 {
		return
	}

	// ENABLE_VIRTUAL_TERMINAL_PROCESSING
	const enableVirtualTerminalProcessing uint32 = 0x0004

	mode |= enableVirtualTerminalProcessing

	setConsoleMode.Call(
		uintptr(handle),
		uintptr(mode),
	)
}