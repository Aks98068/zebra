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
	currentDir := "."
	cliArgs := os.Args[1:]

	if len(cliArgs) == 0 {
		runInteractive(&currentDir)
		return
	}

	// bash already split cliArgs for us -- no need for util.Tokenize here.
	commands.Dispatch(cliArgs, &currentDir)
}

func runInteractive(currentDir *string) {
	ui.PrintBanner()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("zebra (%s) > ", *currentDir)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			fmt.Printf("zebra (%s) > ", *currentDir)
			continue
		}

		tokens := util.Tokenize(line)

		if commands.Dispatch(tokens, currentDir) {
			return
		}

		fmt.Printf("zebra (%s) > ", *currentDir)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "input error:", err)
	}
}
