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

const terminalChildEnv = "ZEBRA_TERMINAL_CHILD"

func main() {
	// ---------------------------------------------------------
	// Initialize Zebra commands
	// ---------------------------------------------------------

	commands.Init()

	// ---------------------------------------------------------
	// Enable ANSI/VT processing on Windows.
	//
	// This makes colors work correctly in:
	//   - CMD
	//   - PowerShell
	//   - VS Code terminal
	//   - Windows Terminal
	// ---------------------------------------------------------

	ui.EnableANSI()

	// ---------------------------------------------------------
	// Determine whether this is a child Zebra terminal.
	// ---------------------------------------------------------

	isChildTerminal := os.Getenv(terminalChildEnv) == "1"

	// ---------------------------------------------------------
	// COMMAND-LINE MODE
	//
	// Examples:
	//
	//     zebra pwd
	//     zebra ls
	//     zebra folder test
	//     zebra file test.txt
	//
	// These execute directly inside the terminal that
	// launched Zebra.
	// ---------------------------------------------------------

	args := os.Args[1:]

	if len(args) > 0 {
		runCommand(args)
		return
	}

	// ---------------------------------------------------------
	// INTERACTIVE MODE
	//
	// We intentionally do NOT automatically create another
	// terminal window here.
	//
	// This is important because:
	//
	//     zebra.exe
	//
	// should behave like:
	//
	//     git
	//     python
	//     node
	//     ssh
	//
	// and directly become an interactive CLI.
	//
	// The "terminal" command can explicitly create another
	// Zebra terminal.
	// ---------------------------------------------------------

	_ = isChildTerminal

	// Start in the directory from which Zebra was launched.
	currentDir := getStartupDirectory()

	runInteractive(&currentDir)

}

// ============================================================
// COMMAND MODE
// ============================================================

func runCommand(args []string) {
	scanner := bufio.NewScanner(os.Stdin)

	currentDir := getStartupDirectory()

	ctx := &commands.Context{
		CurrentDir: &currentDir,
		Scanner:    scanner,
	}

	commands.Execute(args, ctx)

}

// ============================================================
// STARTUP DIRECTORY
// ============================================================
//
// IMPORTANT:
//
// Do NOT use os.UserHomeDir() here.
//
// If the user executes:
//
//     C:\projects\myapp> zebra
//
// Zebra should start at:
//
//     C:\projects\myapp
//
// not:
//
//     C:\Users\Admin
//
// This makes Zebra behave like a real CLI application.
// ============================================================

func getStartupDirectory() string {
	current, err := os.Getwd()

	if err == nil && current != "" {
		absolute, absErr := os.Getwd()

		if absErr == nil && absolute != "" {
			return absolute
		}

		return current
	}

	// Extremely unusual fallback.
	home, err := os.UserHomeDir()

	if err == nil && home != "" {
		return home
	}

	return "."

}

// ============================================================
// INTERACTIVE SHELL
// ============================================================

func runInteractive(currentDir *string) {
	ui.PrintBanner()

	scanner := bufio.NewScanner(os.Stdin)

	ctx := &commands.Context{
		CurrentDir: currentDir,
		Scanner:    scanner,
	}

	editor := shell.NewEditor(
		commands.Names(),
	)

	for {
		prompt := buildPrompt(currentDir)

		line, err := editor.ReadLine(prompt)

		// -----------------------------------------------------
		// Ctrl+C
		// -----------------------------------------------------

		if errors.Is(err, shell.ErrInterrupted) {
			continue
		}

		// -----------------------------------------------------
		// Ctrl+D / EOF
		// -----------------------------------------------------

		if errors.Is(err, shell.ErrEOF) {
			fmt.Println(
				ui.BrightYellow +
					"exit" +
					ui.Reset,
			)

			return
		}

		// -----------------------------------------------------
		// Other shell errors
		// -----------------------------------------------------

		if err != nil {
			fmt.Fprintln(
				os.Stderr,
				ui.BrightRed+"shell input error:"+ui.Reset,
				err,
			)

			return
		}

		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		// -----------------------------------------------------
		// Tokenize command
		// -----------------------------------------------------

		tokens := util.Tokenize(line)

		if len(tokens) == 0 {
			continue
		}

		// -----------------------------------------------------
		// Execute
		// -----------------------------------------------------

		shouldExit := commands.Execute(
			tokens,
			ctx,
		)

		if shouldExit {
			return
		}
	}

}

// ============================================================
// PROMPT
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
