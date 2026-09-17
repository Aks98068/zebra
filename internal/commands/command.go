package commands

import "bufio"

type Command struct {
	Name        string
	Aliases     []string
	Description string
	Usage       string
	Run         func(args []string, ctx *Context) bool
}

type Context struct {
	CurrentDir *string
	Scanner    *bufio.Scanner
}