package commands

import (
	"fmt"
	"os"
)

// os.CreateTemp() — create a temp file
func handleTmpfile(args []string, ctx *Context) bool {
	prefix := "zebra-"
	if len(args) > 0 {
		prefix = args[0]
	}

	f, err := os.CreateTemp("", prefix)
	if err != nil {
		fmt.Println("Error creating temp file:", err)
		return false
	}
	defer f.Close()

	fmt.Println("Temp file created:", f.Name())
	return true
}

// os.MkdirTemp() — create a temp directory
func handleTmpdir(args []string, ctx *Context) bool {
	prefix := "zebra-"
	if len(args) > 0 {
		prefix = args[0]
	}

	dir, err := os.MkdirTemp("", prefix)
	if err != nil {
		fmt.Println("Error creating temp directory:", err)
		return false
	}

	fmt.Println("Temp directory created:", dir)
	return true
}

// os.TempDir() — show system temp directory
func handleTempdir(args []string, ctx *Context) bool {
	fmt.Println("System temp directory:", os.TempDir())
	return true
}