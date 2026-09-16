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

func main() {
	commands.Init()

	currentDir := "."

	cliArgs := os.Args[1:]

	if len(cliArgs) == 0 {
		runInteractive(&currentDir)
		return
	}

	scanner := bufio.NewScanner(os.Stdin)

	ctx := &commands.Context{
		CurrentDir: &currentDir,
		Scanner:    scanner,
	}

	commands.Execute(cliArgs, ctx)
}

func runInteractive(currentDir *string) {
	ui.PrintBanner()

	scanner := bufio.NewScanner(os.Stdin)

	ctx := &commands.Context{
		CurrentDir: currentDir,
		Scanner:    scanner,
	}

	fmt.Printf("zebra (%s) > ", *currentDir)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			fmt.Printf("zebra (%s) > ", *currentDir)
			continue
		}

		tokens := util.Tokenize(line)

		if commands.Execute(tokens, ctx) {
			return
		}

		fmt.Printf("zebra (%s) > ", *currentDir)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "input error:", err)
	}
}
