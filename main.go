package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"zebra/internal/commands"
	"zebra/internal/shell"
	"zebra/internal/ui"
	"zebra/internal/util"
)

// ============================================================
// ENVIRONMENT
// ============================================================

const terminalChildEnv = "ZEBRA_TERMINAL_CHILD"

// ============================================================
// MAIN
// ============================================================

func main() {
	// --------------------------------------------------------
	// Initialize all Zebra commands
	// --------------------------------------------------------

	commands.Init()

	// --------------------------------------------------------
	// Detect child Zebra terminal
	// --------------------------------------------------------

	isChildTerminal := os.Getenv(terminalChildEnv) == "1"

	// --------------------------------------------------------
	// Automatically open Zebra in a new terminal
	// --------------------------------------------------------
	//
	// Only when:
	//
	// 1. Zebra is not already a child terminal.
	// 2. No command-line arguments were supplied.
	//
	// Example:
	//
	//     zebra.exe
	//
	// Zebra will open in a new terminal and start in the
	// user's home directory.
	// --------------------------------------------------------

	if !isChildTerminal && len(os.Args) == 1 {
		initialDir := getInitialDirectory()

		if err := commands.OpenZebraTerminalAt(initialDir); err != nil {
			fmt.Fprintln(
				os.Stderr,
				"Terminal error:",
				err,
			)

			fmt.Fprintln(
				os.Stderr,
				"Starting Zebra in the current terminal...",
			)
		} else {
			return
		}
	}

	// --------------------------------------------------------
	// Determine Zebra's virtual working directory
	// --------------------------------------------------------
	//
	// This is Zebra's own current directory.
	//
	// It does NOT depend on the directory where zebra.exe
	// is installed.
	// --------------------------------------------------------

	currentDir := getInitialDirectory()

	// --------------------------------------------------------
	// Command-line mode
	// --------------------------------------------------------
	//
	// Examples:
	//
	//     zebra.exe pwd
	//     zebra.exe ls
	//     zebra.exe get PATH
	//
	// --------------------------------------------------------

	cliArgs := os.Args[1:]

	if len(cliArgs) > 0 {
		scanner := bufio.NewScanner(os.Stdin)

		ctx := &commands.Context{
			CurrentDir: &currentDir,
			Scanner:    scanner,
		}

		commands.Execute(
			cliArgs,
			ctx,
		)

		return
	}

	// --------------------------------------------------------
	// Interactive Zebra shell
	// --------------------------------------------------------

	runInteractive(&currentDir)
}

// ============================================================
// INITIAL DIRECTORY
// ============================================================
//
// Zebra starts in the user's HOME directory.
//
// Windows:
//
//     C:\Users\ACER
//
// Linux:
//
//     /home/username
//
// macOS:
//
//     /Users/username
//
// This is independent of where zebra.exe is installed.
//
// Example:
//
//     Zebra executable:
//         D:\Programs\Zebra\zebra.exe
//
//     Initial Zebra directory:
//         C:\Users\ACER
//
// ============================================================

func getInitialDirectory() string {
	// --------------------------------------------------------
	// First choice: user's home directory
	// --------------------------------------------------------

	home, err := os.UserHomeDir()

	if err == nil && home != "" {
		return home
	}

	// --------------------------------------------------------
	// Fallback: current operating-system directory
	// --------------------------------------------------------

	current, err := os.Getwd()

	if err == nil && current != "" {
		return current
	}

	// --------------------------------------------------------
	// Final fallback
	// --------------------------------------------------------

	return "."
}

// ============================================================
// INTERACTIVE SHELL
// ============================================================

func runInteractive(currentDir *string) {
	// --------------------------------------------------------
	// Banner
	// --------------------------------------------------------

	ui.PrintBanner()

	// --------------------------------------------------------
	// Scanner
	// --------------------------------------------------------
	//
	// Scanner is still required by commands that need
	// additional interactive input, such as:
	//
	//     write
	//
	// The main command line is handled by shell.Editor.
	// --------------------------------------------------------

	scanner := bufio.NewScanner(os.Stdin)

	ctx := &commands.Context{
		CurrentDir: currentDir,
		Scanner:    scanner,
	}

	// --------------------------------------------------------
	// Create interactive line editor
	// --------------------------------------------------------
	//
	// Provides:
	//
	//     Arrow keys
	//     History
	//     Cursor movement
	//     Home
	//     End
	//     Backspace
	//     Delete
	//     Tab completion
	//     Ctrl+C
	//     Ctrl+D
	//     Ctrl+L
	//
	// --------------------------------------------------------

	editor := shell.NewEditor(
		commands.Names(),
	)

	// --------------------------------------------------------
	// Main shell loop
	// --------------------------------------------------------

	for {
		// ----------------------------------------------------
		// Build colorful prompt
		// ----------------------------------------------------

		prompt := buildPrompt(currentDir)

		// ----------------------------------------------------
		// Read command
		// ----------------------------------------------------

		line, err := editor.ReadLine(prompt)

		// ----------------------------------------------------
		// Ctrl+C
		// ----------------------------------------------------

		if errors.Is(err, shell.ErrInterrupted) {
			continue
		}

		// ----------------------------------------------------
		// Ctrl+D
		// ----------------------------------------------------

		if errors.Is(err, shell.ErrEOF) {
			fmt.Println(
				ui.BrightYellow +
					"exit" +
					ui.Reset,
			)

			return
		}

		// ----------------------------------------------------
		// Other input errors
		// ----------------------------------------------------

		if err != nil {
			fmt.Fprintln(
				os.Stderr,
				ui.BrightRed+"shell input error:"+ui.Reset,
				err,
			)

			return
		}

		// ----------------------------------------------------
		// Remove surrounding whitespace
		// ----------------------------------------------------

		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		// ----------------------------------------------------
		// Tokenize command
		// ----------------------------------------------------

		tokens := util.Tokenize(line)

		if len(tokens) == 0 {
			continue
		}

		// ----------------------------------------------------
		// Execute Zebra command
		// ----------------------------------------------------

		shouldExit := commands.Execute(
			tokens,
			ctx,
		)

		// ----------------------------------------------------
		// Exit command
		// ----------------------------------------------------

		if shouldExit {
			return
		}
	}
}

// ============================================================
// ZEBRA PROMPT
// ============================================================
//
// Example:
//
//     zebra (C:\Users\ACER) >
//
// Colors:
//
//     zebra       → cyan
//     ( )         → gray
//     directory   → green
//     >           → magenta
//
// ============================================================

func buildPrompt(currentDir *string) string {
	return ui.BrightCyan +
		ui.Bold +
		"zebra" +
		ui.Reset +

		" " +

		ui.BrightBlack +
		"(" +
		ui.Reset +

		ui.BrightGreen +
		*currentDir +
		ui.Reset +

		ui.BrightBlack +
		")" +
		ui.Reset +

		" " +

		ui.BrightMagenta +
		ui.Bold +
		">" +
		ui.Reset +

		" "
}