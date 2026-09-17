//go:build windows

package privilege

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"
)

var (
	shell32           = syscall.NewLazyDLL("shell32.dll")
	procShellExecuteW = shell32.NewProc("ShellExecuteW")
)

const swShow = 1

// EnsureElevated makes sure Zebra is running as Administrator.
//
// IMPORTANT:
// Only call this when the operation actually requires elevation.
func EnsureElevated() error {
	if isAdmin() {
		return nil
	}

	return relaunchAsAdmin()
}

// isAdmin checks whether the current process is running
// with an Administrator token.
func isAdmin() bool {
	cmd := exec.Command(
		"powershell.exe",
		"-NoProfile",
		"-NonInteractive",
		"-ExecutionPolicy",
		"Bypass",
		"-Command",
		`$p = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent()); if ($p.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) { "TRUE" } else { "FALSE" }`,
	)

	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(output)) == "TRUE"
}

// relaunchAsAdmin starts a new elevated Zebra process.
//
// After successfully starting the elevated process,
// the current non-elevated process terminates immediately.
// This prevents the original + elevated processes from
// both continuing to run.
func relaunchAsAdmin() error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("unable to find Zebra executable: %w", err)
	}

	// Convert the executable path to UTF-16.
	executablePtr, err := syscall.UTF16PtrFromString(executable)
	if err != nil {
		return fmt.Errorf("unable to convert executable path: %w", err)
	}

	// "runas" tells Windows to launch the application elevated.
	verbPtr, err := syscall.UTF16PtrFromString("runas")
	if err != nil {
		return fmt.Errorf("unable to convert runas verb: %w", err)
	}

	// Preserve the original command-line arguments.
	parameters := buildWindowsArguments(os.Args[1:])

	parametersPtr, err := syscall.UTF16PtrFromString(parameters)
	if err != nil {
		return fmt.Errorf("unable to convert command arguments: %w", err)
	}

	ret, _, _ := procShellExecuteW.Call(
		0,
		uintptr(unsafe.Pointer(verbPtr)),
		uintptr(unsafe.Pointer(executablePtr)),
		uintptr(unsafe.Pointer(parametersPtr)),
		0,
		uintptr(swShow),
	)

	// ShellExecute returns a value <= 32 when it fails.
	if ret <= 32 {
		switch ret {
		case 5:
			return fmt.Errorf("access denied: Windows refused to launch Zebra as Administrator")
		case 2:
			return fmt.Errorf("Zebra executable was not found: %s", executable)
		case 3:
			return fmt.Errorf("Zebra executable path was not found: %s", executable)
		case 1223:
			return fmt.Errorf("Administrator elevation was cancelled")
		default:
			return fmt.Errorf(
				"failed to start elevated Zebra (ShellExecuteW code %d)",
				ret,
			)
		}
	}

	// CRITICAL:
	// The elevated Zebra process has now been launched.
	// Kill this original process so we don't create duplicates.
	os.Exit(0)

	return nil
}

// buildWindowsArguments converts arguments into a Windows-compatible
// command-line argument string.
func buildWindowsArguments(args []string) string {
	if len(args) == 0 {
		return ""
	}

	var result strings.Builder

	for i, arg := range args {
		if i > 0 {
			result.WriteByte(' ')
		}

		result.WriteString(quoteWindowsArgument(arg))
	}

	return result.String()
}

// quoteWindowsArgument safely quotes one Windows command-line argument.
func quoteWindowsArgument(arg string) string {
	if arg == "" {
		return `""`
	}

	needsQuotes := false

	for _, r := range arg {
		if r == ' ' || r == '\t' || r == '"' {
			needsQuotes = true
			break
		}
	}

	if !needsQuotes {
		return arg
	}

	var result strings.Builder

	result.WriteByte('"')

	backslashes := 0

	for _, r := range arg {
		switch r {
		case '\\':
			backslashes++

		case '"':
			for i := 0; i < backslashes*2+1; i++ {
				result.WriteByte('\\')
			}

			result.WriteByte('"')
			backslashes = 0

		default:
			for i := 0; i < backslashes; i++ {
				result.WriteByte('\\')
			}

			backslashes = 0
			result.WriteRune(r)
		}
	}

	// Backslashes immediately before the closing quote
	// must be doubled.
	for i := 0; i < backslashes*2; i++ {
		result.WriteByte('\\')
	}

	result.WriteByte('"')

	return result.String()
}