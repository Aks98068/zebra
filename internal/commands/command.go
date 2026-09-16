package commands

import "bufio"

// Command represents a Zebra CLI command.
type Command struct {
	Name        string
	Aliases     []string
	Description string
	Usage       string
	Run         func(args []string, ctx *Context) bool
}

// Context contains state shared between commands.
type Context struct {
	CurrentDir *string
	Scanner    *bufio.Scanner
}
