package commands

import (
	"fmt"

	"zebra/internal/ui"
)

func handlePwd(args []string, ctx *Context) bool {
	fmt.Println(*ctx.CurrentDir)

	// false = keep Zebra running
	return false
}

func handleHelp(args []string, ctx *Context) bool {
	ui.PrintHelp()

	// false = keep Zebra running
	return false
}

func handleExit(args []string, ctx *Context) bool {
	fmt.Println("bye")

	// true = exit the interactive Zebra shell
	return true
}