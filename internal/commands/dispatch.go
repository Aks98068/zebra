package commands

import (
	"fmt"

	"zebra/internal/ui"
)

// Dispatch routes tokenized command words to the right handler.
// Note the import above: "zebra/internal/ui" -- "zebra" is the
// module name from go.mod, and the rest is the folder path to
// the ui package. This is how one internal package calls another.
func Dispatch(tokens []string, currentDir *string) (shouldExit bool) {
	if len(tokens) == 0 {
		return false
	}

	cmd := tokens[0]
	args := tokens[1:]

	switch cmd {
	case "folder":
		HandleFolder(currentDir, args)
	case "file":
		HandleFile(*currentDir, args)
	case "cd":
		HandleCd(currentDir, args)
	case "pwd":

		fmt.Println(*currentDir)
	case "delete", "del":
		HandleRemove(*currentDir, args)
	case "help":
		ui.PrintHelp()
	case "exit", "quit":
		fmt.Println("bye")
		return true
	default:
		fmt.Println("unknown command:", cmd)
	}
	return false
}
