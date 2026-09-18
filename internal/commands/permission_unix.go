//go:build !windows

package commands

import (
	"fmt"
	"os"
	"strconv"
)

func handleChmod(args []string, ctx *Context) bool {
	if len(args) < 2 {
		fmt.Println("Usage: chmod <permissions> <file>")
		fmt.Println("Example: chmod 755 file.txt")
		return false
	}

	perm, err := strconv.ParseUint(args[0], 8, 32)
	if err != nil {
		fmt.Println("Invalid permission format. Use octal e.g: 755, 644")
		return false
	}

	err = os.Chmod(args[1], os.FileMode(perm))
	if err != nil {
		fmt.Println("Error changing permissions:", err)
		return false
	}

	fmt.Printf("Permissions of %s changed to %s\n", args[1], args[0])
	return true
}

func handleChown(args []string, ctx *Context) bool {
	if len(args) < 3 {
		fmt.Println("Usage: chown <uid> <gid> <file>")
		fmt.Println("Example: chown 1000 1000 file.txt")
		return false
	}

	uid, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Invalid UID:", args[0])
		return false
	}

	gid, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Println("Invalid GID:", args[1])
		return false
	}

	err = os.Chown(args[2], uid, gid)
	if err != nil {
		fmt.Println("Error changing owner:", err)
		return false
	}

	fmt.Printf("Owner of %s changed to UID:%d GID:%d\n", args[2], uid, gid)
	return true
}

func handleLchown(args []string, ctx *Context) bool {
	if len(args) < 3 {
		fmt.Println("Usage: lchown <uid> <gid> <link>")
		return false
	}

	uid, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Invalid UID:", args[0])
		return false
	}

	gid, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Println("Invalid GID:", args[1])
		return false
	}

	err = os.Lchown(args[2], uid, gid)
	if err != nil {
		fmt.Println("Error changing symlink owner:", err)
		return false
	}

	fmt.Printf("Symlink owner of %s changed to UID:%d GID:%d\n", args[2], uid, gid)
	return true
}