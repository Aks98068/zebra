package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func handleFolder(args []string, ctx *Context) bool {

	if len(args) < 1 {
		fmt.Println("Usage: folder <path>")
		return false
	}

	path := args[0]

	if !filepath.IsAbs(path) {
		path = filepath.Join(*ctx.CurrentDir, path)
	}

	err := os.MkdirAll(path, 0755)

	if err != nil {
		fmt.Println("Error creating folder:", err)
		return false
	}

	absolutePath, err := filepath.Abs(path)

	if err == nil {
		*ctx.CurrentDir = absolutePath
	}

	fmt.Println("Folder created:", path)

	return false
}

func handleFile(args []string, ctx *Context) bool {

	if len(args) < 1 {
		fmt.Println("Usage: file <path>")
		return false
	}

	path := args[0]

	if !filepath.IsAbs(path) {
		path = filepath.Join(*ctx.CurrentDir, path)
	}

	file, err := os.Create(path)

	if err != nil {
		fmt.Println("Error creating file:", err)
		return false
	}

	file.Close()

	fmt.Println("File created:", path)

	return false
}

func handleCd(args []string, ctx *Context) bool {

	if len(args) < 1 {
		fmt.Println("Usage: cd <path>")
		return false
	}

	path := args[0]

	if path == "~" {
		home, err := os.UserHomeDir()

		if err != nil {
			fmt.Println("Error:", err)
			return false
		}

		path = home
	} else if !filepath.IsAbs(path) {
		path = filepath.Join(*ctx.CurrentDir, path)
	}

	absolutePath, err := filepath.Abs(path)

	if err != nil {
		fmt.Println("Error:", err)
		return false
	}

	info, err := os.Stat(absolutePath)

	if err != nil {
		fmt.Println("Error:", err)
		return false
	}

	if !info.IsDir() {
		fmt.Println("Not a directory:", absolutePath)
		return false
	}

	*ctx.CurrentDir = absolutePath

	return false
}

func handleRemove(args []string, ctx *Context) bool {

	if len(args) < 1 {
		fmt.Println("Usage: remove <path>")
		return false
	}

	path := args[0]

	if !filepath.IsAbs(path) {
		path = filepath.Join(*ctx.CurrentDir, path)
	}

	err := os.RemoveAll(path)

	if err != nil {
		fmt.Println("Error removing:", err)
		return false
	}

	fmt.Println("Removed:", path)

	return false
}

func handleCopy(args []string, ctx *Context) bool {

	if len(args) < 2 {
		fmt.Println("Usage: copy <source> <destination>")
		return false
	}

	source := args[0]
	destination := args[1]

	if !filepath.IsAbs(source) {
		source = filepath.Join(*ctx.CurrentDir, source)
	}

	if !filepath.IsAbs(destination) {
		destination = filepath.Join(*ctx.CurrentDir, destination)
	}

	sourceData, err := os.ReadFile(source)

	if err != nil {
		fmt.Println("Error reading source:", err)
		return false
	}

	err = os.WriteFile(destination, sourceData, 0644)

	if err != nil {
		fmt.Println("Error writing destination:", err)
		return false
	}

	fmt.Println("Copied:", source, "->", destination)

	return false
}

func handleMove(args []string, ctx *Context) bool {

	if len(args) < 2 {
		fmt.Println("Usage: move <source> <destination>")
		return false
	}

	source := args[0]
	destination := args[1]

	if !filepath.IsAbs(source) {
		source = filepath.Join(*ctx.CurrentDir, source)
	}

	if !filepath.IsAbs(destination) {
		destination = filepath.Join(*ctx.CurrentDir, destination)
	}

	err := os.Rename(source, destination)

	if err != nil {
		fmt.Println("Error moving:", err)
		return false
	}

	fmt.Println("Moved:", source, "->", destination)

	return false
}

func handleCat(args []string, ctx *Context) bool {

	if len(args) < 1 {
		fmt.Println("Usage: cat <file>")
		return false
	}

	path := args[0]

	if !filepath.IsAbs(path) {
		path = filepath.Join(*ctx.CurrentDir, path)
	}

	data, err := os.ReadFile(path)

	if err != nil {
		fmt.Println("Error reading file:", err)
		return false
	}

	fmt.Println(string(data))

	return false
}

func handleWrite(args []string, ctx *Context) bool {

	if len(args) < 1 {
		fmt.Println("Usage: write <file>")
		return false
	}

	path := args[0]

	if !filepath.IsAbs(path) {
		path = filepath.Join(*ctx.CurrentDir, path)
	}

	fmt.Println("Enter content.")
	fmt.Println("Type END on a new line to finish.")

	var lines []string

	for ctx.Scanner.Scan() {

		line := ctx.Scanner.Text()

		if strings.TrimSpace(line) == "END" {
			break
		}

		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")

	err := os.WriteFile(
		path,
		[]byte(content),
		0644,
	)

	if err != nil {
		fmt.Println("Error writing file:", err)
		return false
	}

	fmt.Println("File written:", path)

	return false
}