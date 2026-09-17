package commands

import (
	"fmt"
	"os"
	"os/exec"
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

	if err := OpenZebraTerminal(); err != nil {
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
// This function can be called from main.go:
//
//     commands.OpenZebraTerminal()
//
// ============================================================

func OpenZebraTerminal() error {
	return openNewTerminal()
}

// ============================================================
// OPEN NEW TERMINAL
// ============================================================

func openNewTerminal() error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf(
			"unable to find Zebra executable: %w",
			err,
		)
	}

	workingDir, err := os.Getwd()
	if err != nil {
		workingDir = "."
	}

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
// This prevents automatic recursive terminal creation.
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

	if wtPath, err := exec.LookPath("wt.exe"); err == nil {

		cmd := exec.Command(
			wtPath,
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
	// Command Prompt fallback
	// --------------------------------------------------------

	cmd := exec.Command(
		"cmd.exe",
		"/c",
		"start",
		"Zebra",
		executable,
	)

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