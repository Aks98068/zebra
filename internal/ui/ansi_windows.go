//go:build windows

package ui

import (
	"syscall"
	"unsafe"
)

func EnableANSI() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")

	getConsoleMode := kernel32.NewProc("GetConsoleMode")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")

	stdout := syscall.Handle(syscall.Stdout)

	var mode uint32

	ret, _, _ := getConsoleMode.Call(
		uintptr(stdout),
		uintptr(unsafe.Pointer(&mode)),
	)

	if ret == 0 {
		return
	}

	// ENABLE_VIRTUAL_TERMINAL_PROCESSING
	const enableVirtualTerminalProcessing uint32 = 0x0004

	mode |= enableVirtualTerminalProcessing

	setConsoleMode.Call(
		uintptr(stdout),
		uintptr(mode),
	)
}