//go:build windows

package commands

import (
	"fmt"
	"os"
	"strconv"
)

// Windows only supports read-only flag via chmod
// It does NOT support Unix-style octal permissions
func handleChmod(args []string, ctx *Context) bool {
	if len(args) < 2 {
		fmt.Println("Usage: chmod <permissions> <file>")
		fmt.Println("Note: Windows only supports 444 (read-only) or 666 (read-write)")
		return false
	}

	perm, err := strconv.ParseUint(args[0], 8, 32)
	if err != nil {
		fmt.Println("Invalid permission format. Use octal e.g: 444, 666")
		return false
	}

	err = os.Chmod(args[1], os.FileMode(perm))
	if err != nil {
		fmt.Println("Error changing permissions:", err)
		return false
	}

	fmt.Printf("Permissions of %s changed to %s\n", args[1], args[0])
	fmt.Println("Note: Windows only applies read-only flag, not full Unix permissions")
	return true
}

// Windows uses ACLs not uid/gid — not supported
func handleChown(args []string, ctx *Context) bool {
	fmt.Println("chown is not supported on Windows.")
	fmt.Println("Windows uses ACLs (Access Control Lists) for ownership.")
	fmt.Println("Use 'icacls' in Command Prompt or PowerShell instead.")
	return false
}

// Same — not supported on Windows
func handleLchown(args []string, ctx *Context) bool {
	fmt.Println("lchown is not supported on Windows.")
	fmt.Println("Windows uses ACLs (Access Control Lists) for ownership.")
	fmt.Println("Use 'icacls' in Command Prompt or PowerShell instead.")
	return false
}