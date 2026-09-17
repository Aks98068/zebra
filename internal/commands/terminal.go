package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ============================================================
// TERMINAL COMMAND
// ============================================================
//
// Usage:
//
//     terminal
//     term
//     newterminal
//
// Opens exactly ONE new Zebra terminal.
//
// Supported:
//
//     Windows
//     Linux
//     macOS
//
// The new terminal starts in Zebra's CURRENT DIRECTORY,
// not in the directory where zebra.exe is installed.
//
// ============================================================

const zebraTerminalChildEnv = "ZEBRA_TERMINAL_CHILD"

// ============================================================
// TERMINAL COMMAND HANDLER
// ============================================================

func handleTerminal(args []string, ctx *Context) bool {
	if len(args) > 0 {
		fmt.Println("Usage: terminal")
		return false
	}

	// --------------------------------------------------------
	// IMPORTANT:
	//
	// Use Zebra's virtual current directory.
	//
	// Example:
	//
	// Zebra executable:
	//     D:\Programs\Zebra\zebra.exe
	//
	// Current Zebra directory:
	//     C:\Users\ACER\Documents
	//
	// The new terminal opens at:
	//     C:\Users\ACER\Documents
	// --------------------------------------------------------

	if err := OpenZebraTerminalAt(*ctx.CurrentDir); err != nil {
		fmt.Println("terminal:", err)
		return false
	}

	fmt.Println("New Zebra terminal opened.")

	return false
}

// ============================================================
// PUBLIC FUNCTION
// ============================================================
//
// Opens a new Zebra terminal using the operating system's
// normal terminal application.
//
// This function is useful when no Zebra Context is available.
//
// ============================================================

func OpenZebraTerminal() error {
	workingDir, err := os.Getwd()

	if err != nil {
		workingDir = "."
	}

	return OpenZebraTerminalAt(workingDir)
}

// ============================================================
// PUBLIC FUNCTION
// ============================================================
//
// Opens a new Zebra terminal at a specific directory.
//
// This is the function that should be used by:
//
//     handleTerminal()
//     main.go
//
// ============================================================

func OpenZebraTerminalAt(workingDir string) error {

	// --------------------------------------------------------
	// Normalize working directory
	// --------------------------------------------------------

	if strings.TrimSpace(workingDir) == "" {
		workingDir = "."
	}

	absoluteDir, err := filepath.Abs(workingDir)

	if err == nil {
		workingDir = filepath.Clean(absoluteDir)
	}

	// --------------------------------------------------------
	// Verify directory exists
	// --------------------------------------------------------

	info, err := os.Stat(workingDir)

	if err != nil {
		return fmt.Errorf(
			"working directory does not exist: %s: %w",
			workingDir,
			err,
		)
	}

	if !info.IsDir() {
		return fmt.Errorf(
			"working path is not a directory: %s",
			workingDir,
		)
	}

	return openNewTerminal(workingDir)
}

// ============================================================
// OPEN NEW TERMINAL
// ============================================================

func openNewTerminal(workingDir string) error {

	// --------------------------------------------------------
	// Find Zebra executable
	// --------------------------------------------------------

	executable, err := os.Executable()

	if err != nil {
		return fmt.Errorf(
			"unable to find Zebra executable: %w",
			err,
		)
	}

	// --------------------------------------------------------
	// Resolve executable to absolute path
	// --------------------------------------------------------

	executable, err = filepath.Abs(executable)

	if err != nil {
		return fmt.Errorf(
			"unable to resolve Zebra executable: %w",
			err,
		)
	}

	executable = filepath.Clean(executable)

	// --------------------------------------------------------
	// Make sure executable exists
	// --------------------------------------------------------

	if _, err := os.Stat(executable); err != nil {
		return fmt.Errorf(
			"Zebra executable does not exist: %s: %w",
			executable,
			err,
		)
	}

	// --------------------------------------------------------
	// Select platform
	// --------------------------------------------------------

	switch runtime.GOOS {

	case "windows":
		return openWindowsTerminal(
			executable,
			workingDir,
		)

	case "linux":
		return openLinuxTerminal(
			executable,
			workingDir,
		)

	case "darwin":
		return openMacTerminal(
			executable,
			workingDir,
		)

	default:
		return fmt.Errorf(
			"unsupported operating system: %s",
			runtime.GOOS,
		)
	}
}

// ============================================================
// CHILD ENVIRONMENT
// ============================================================
//
// The child Zebra receives:
//
//     ZEBRA_TERMINAL_CHILD=1
//
// This is important because main.go automatically opens a new
// terminal when Zebra starts without arguments.
//
// Without this flag:
//
//     Zebra
//       ↓
//     opens Zebra
//       ↓
//     opens Zebra
//       ↓
//     opens Zebra
//
// With this flag, the child starts directly inside the newly
// opened terminal.
//
// ============================================================

func zebraChildEnvironment() []string {
	env := append(
		[]string{},
		os.Environ()...,
	)

	env = append(
		env,
		zebraTerminalChildEnv+"=1",
	)

	return env
}

// ============================================================
// WINDOWS
// ============================================================

func openWindowsTerminal(
	executable string,
	workingDir string,
) error {

	env := zebraChildEnvironment()

	// --------------------------------------------------------
	// Windows Terminal
	// --------------------------------------------------------
	//
	// Prefer Windows Terminal when installed.
	//
	// -w new
	//     Create a NEW Windows Terminal window.
	//
	// new-tab
	//     Create a new tab in that window.
	//
	// --startingDirectory
	//     Start Zebra in the requested directory.
	//
	// --------------------------------------------------------

	if wtPath, err := exec.LookPath("wt.exe"); err == nil {

		cmd := exec.Command(
			wtPath,
			"-w",
			"new",
			"new-tab",
			"--startingDirectory",
			workingDir,
			executable,
		)

		cmd.Env = env

		if err := cmd.Start(); err == nil {
			return nil
		}
	}

	// --------------------------------------------------------
	// Windows CMD fallback
	// --------------------------------------------------------
	//
	// Do NOT rely on:
	//
	//     cmd.exe
	//
	// being available in PATH.
	//
	// Use:
	//
	//     C:\Windows\System32\cmd.exe
	//
	// through %SystemRoot%.
	// --------------------------------------------------------

	systemRoot := os.Getenv("SystemRoot")

	if systemRoot == "" {
		systemRoot = `C:\Windows`
	}

	cmdPath := filepath.Join(
		systemRoot,
		"System32",
		"cmd.exe",
	)

	// --------------------------------------------------------
	// Verify CMD exists
	// --------------------------------------------------------

	if _, err := os.Stat(cmdPath); err != nil {
		return fmt.Errorf(
			"Windows CMD was not found at %s: %w",
			cmdPath,
			err,
		)
	}

	// --------------------------------------------------------
	// Start CMD in the requested directory
	// --------------------------------------------------------

	cmd := exec.Command(
		cmdPath,
		"/c",
		"start",
		"Zebra",
		executable,
	)

	cmd.Dir = workingDir
	cmd.Env = env

	if err := cmd.Start(); err != nil {
		return fmt.Errorf(
			"unable to open Windows terminal: %w",
			err,
		)
	}

	return nil
}

// ============================================================
// LINUX
// ============================================================

func openLinuxTerminal(
	executable string,
	workingDir string,
) error {

	env := zebraChildEnvironment()

	// --------------------------------------------------------
	// x-terminal-emulator
	// --------------------------------------------------------

	if terminal, err := exec.LookPath(
		"x-terminal-emulator",
	); err == nil {

		cmd := exec.Command(
			terminal,
			"--working-directory",
			workingDir,
			"-e",
			executable,
		)

		cmd.Env = env

		if err := cmd.Start(); err == nil {
			return nil
		}
	}

	// --------------------------------------------------------
	// GNOME Terminal
	// --------------------------------------------------------

	if terminal, err := exec.LookPath(
		"gnome-terminal",
	); err == nil {

		cmd := exec.Command(
			terminal,
			"--working-directory",
			workingDir,
			"--",
			executable,
		)

		cmd.Env = env

		if err := cmd.Start(); err == nil {
			return nil
		}
	}

	// --------------------------------------------------------
	// KDE Konsole
	// --------------------------------------------------------

	if terminal, err := exec.LookPath(
		"konsole",
	); err == nil {

		cmd := exec.Command(
			terminal,
			"--workdir",
			workingDir,
			"-e",
			executable,
		)

		cmd.Env = env

		if err := cmd.Start(); err == nil {
			return nil
		}
	}

	// --------------------------------------------------------
	// XFCE Terminal
	// --------------------------------------------------------

	if terminal, err := exec.LookPath(
		"xfce4-terminal",
	); err == nil {

		cmd := exec.Command(
			terminal,
			"--working-directory",
			workingDir,
			"--command",
			executable,
		)

		cmd.Env = env

		if err := cmd.Start(); err == nil {
			return nil
		}
	}

	// --------------------------------------------------------
	// xterm
	// --------------------------------------------------------

	if terminal, err := exec.LookPath(
		"xterm",
	); err == nil {

		cmd := exec.Command(
			terminal,
			"-e",
			executable,
		)

		cmd.Dir = workingDir
		cmd.Env = env

		if err := cmd.Start(); err == nil {
			return nil
		}
	}

	return fmt.Errorf(
		"no supported Linux terminal emulator was found",
	)
}

// ============================================================
// macOS
// ============================================================

func openMacTerminal(
	executable string,
	workingDir string,
) error {

	env := zebraChildEnvironment()

	// --------------------------------------------------------
	// Build shell command
	// --------------------------------------------------------

	command := "cd " +
		shellQuote(workingDir) +
		" && " +
		shellQuote(executable)

	// --------------------------------------------------------
	// Escape for AppleScript
	// --------------------------------------------------------

	appleScriptCommand := escapeAppleScript(command)

	script := fmt.Sprintf(
		`tell application "Terminal"
			activate
			do script "%s"
		end tell`,
		appleScriptCommand,
	)

	cmd := exec.Command(
		"osascript",
		"-e",
		script,
	)

	cmd.Env = env

	if err := cmd.Start(); err != nil {
		return fmt.Errorf(
			"unable to open macOS Terminal.app: %w",
			err,
		)
	}

	return nil
}

// ============================================================
// SHELL QUOTING
// ============================================================

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}

	return "'" +
		strings.ReplaceAll(
			value,
			"'",
			"'\\''",
		) +
		"'"
}

// ============================================================
// APPLESCRIPT ESCAPING
// ============================================================

func escapeAppleScript(value string) string {

	value = strings.ReplaceAll(
		value,
		"\\",
		"\\\\",
	)

	value = strings.ReplaceAll(
		value,
		`"`,
		`\"`,
	)

	return value
}