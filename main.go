package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"zebra/internal/commands"
	"zebra/internal/ui"
	"zebra/internal/util"
)

const terminalChildEnv = "ZEBRA_TERMINAL_CHILD"

func main() {

	// =========================================================
	// Initialize commands
	// =========================================================

	commands.Init()

	// =========================================================
	// Check whether this process is already a child terminal.
	// =========================================================

	isChildTerminal := os.Getenv(terminalChildEnv) == "1"

	// =========================================================
	// AUTOMATIC TERMINAL STARTUP
	// =========================================================
	//
	// When the user runs:
	//
	//     .\zebra.exe
	//
	// Zebra opens ONE separate terminal.
	//
	// The new Zebra process receives:
	//
	//     ZEBRA_TERMINAL_CHILD=1
	//
	// Therefore the child does NOT create another terminal.
	//
	// =========================================================

	if !isChildTerminal && len(os.Args) == 1 {

		if err := commands.OpenZebraTerminal(); err != nil {

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

			// Parent process exits.
			return
		}
	}

	// =========================================================
	// Current directory
	// =========================================================

	currentDir := "."

	// =========================================================
	// One-shot command mode
	//
	// Examples:
	//
	//     zebra.exe pwd
	//     zebra.exe get PATH
	//
	// =========================================================

	cliArgs := os.Args[1:]

	if len(cliArgs) > 0 {

		scanner := bufio.NewScanner(os.Stdin)

		ctx := &commands.Context{
			CurrentDir: &currentDir,
			Scanner:    scanner,
		}

		commands.Execute(cliArgs, ctx)

		return
	}

	// =========================================================
	// Interactive mode
	// =========================================================

	runInteractive(&currentDir)
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

	printPrompt(currentDir)

	for scanner.Scan() {

		line := strings.TrimSpace(
			scanner.Text(),
		)

		// -----------------------------------------------------
		// Empty input
		// -----------------------------------------------------

		if line == "" {
			printPrompt(currentDir)
			continue
		}

		// -----------------------------------------------------
		// Tokenize command
		// -----------------------------------------------------

		tokens := util.Tokenize(line)

		if len(tokens) == 0 {
			printPrompt(currentDir)
			continue
		}

		// -----------------------------------------------------
		// Execute command
		// -----------------------------------------------------

		shouldExit := commands.Execute(
			tokens,
			ctx,
		)

		if shouldExit {
			return
		}

		printPrompt(currentDir)
	}

	// =========================================================
	// Scanner error
	// =========================================================

	if err := scanner.Err(); err != nil {

		fmt.Fprintln(
			os.Stderr,
			"input error:",
			err,
		)
	}
}

// ============================================================
// PROMPT
// ============================================================

func printPrompt(currentDir *string) {
	fmt.Printf(
		"zebra (%s) > ",
		*currentDir,
	)
}