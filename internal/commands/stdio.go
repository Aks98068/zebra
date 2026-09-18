package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// write to stdout directly
func handleStdout(args []string, ctx *Context) bool {
	if len(args) < 1 {
		fmt.Println("Usage: stdout <message>")
		return false
	}

	message := strings.Join(args, " ")
	os.Stdout.WriteString(message + "\n")
	return true
}

// write to stderr directly
func handleStderr(args []string, ctx *Context) bool {
	if len(args) < 1 {
		fmt.Println("Usage: stderr <message>")
		return false
	}

	message := strings.Join(args, " ")
	os.Stderr.WriteString(message + "\n")
	return true
}

// read a line from stdin
func handleStdin(args []string, ctx *Context) bool {
	fmt.Print("Input: ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		fmt.Println("You entered:", scanner.Text())
	}
	return true
}