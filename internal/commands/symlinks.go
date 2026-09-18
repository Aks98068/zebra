package commands

import (
	"fmt"
	"os"
)

// os.Symlink("target", "link")
func handleSymlink(args []string, ctx *Context) bool {
	if len(args) < 2 {
		fmt.Println("Usage: symlink <target> <link>")
		return false
	}

	err := os.Symlink(args[0], args[1])
	if err != nil {
		fmt.Println("Error creating symlink:", err)
		return false
	}

	fmt.Printf("Symlink created: %s -> %s\n", args[1], args[0])
	return true
}

// os.Readlink("link")
func handleReadlink(args []string, ctx *Context) bool {
	if len(args) < 1 {
		fmt.Println("Usage: readlink <link>")
		return false
	}

	target, err := os.Readlink(args[0])
	if err != nil {
		fmt.Println("Error reading symlink:", err)
		return false
	}

	fmt.Printf("%s -> %s\n", args[0], target)
	return true
}

// os.Link("src", "dst")
func handleHardlink(args []string, ctx *Context) bool {
	if len(args) < 2 {
		fmt.Println("Usage: hardlink <source> <destination>")
		return false
	}

	err := os.Link(args[0], args[1])
	if err != nil {
		fmt.Println("Error creating hard link:", err)
		return false
	}

	fmt.Printf("Hard link created: %s -> %s\n", args[1], args[0])
	return true
}