package commands

import (
	"fmt"

	"zebra/internal/ui"
)

func handlePwd(args []string, ctx *Context) bool {
	fmt.Println(*ctx.CurrentDir)
	return false
}

func handleHelp(args []string, ctx *Context) bool {
	ui.PrintHelp()
	return false
}

func handleExit(args []string, ctx *Context) bool {
	fmt.Println("bye")
	return true
}
