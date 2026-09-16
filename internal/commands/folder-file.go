package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"zebra/internal/util"
)

// handleFolder creates one or more folders.
//
// Usage:
//
//	folder <name> [name2 ...]
func handleFolder(args []string, ctx *Context) bool {
	if len(args) < 1 {
		fmt.Println("usage: folder <name> [name2 ...]")
		return false
	}

	for _, name := range args {
		path := util.ResolvePath(*ctx.CurrentDir, name)

		if err := os.MkdirAll(path, 0755); err != nil {
			fmt.Println("error creating folder:", err)
			continue
		}

		fmt.Println("created folder:", path)
	}

	// Preserve your existing behavior:
	// after creating multiple folders, enter the first one.
	*ctx.CurrentDir = util.ResolvePath(*ctx.CurrentDir, args[0])

	return false
}

// handleFile creates one or more files.
//
// Usage:
//
//	file <name> [name2 ...]
func handleFile(args []string, ctx *Context) bool {
	if len(args) < 1 {
		fmt.Println("usage: file <name> [name2 ...]")
		return false
	}

	for _, name := range args {
		path := util.ResolvePath(*ctx.CurrentDir, name)

		f, err := os.Create(path)
		if err != nil {
			fmt.Println("error creating file:", err)
			continue
		}

		if err := f.Close(); err != nil {
			fmt.Println("error closing file:", err)
			continue
		}

		fmt.Println("created file:", path)
	}

	return false
}

// handleCd changes the current Zebra directory.
//
// Usage:
//
//	cd <name>
func handleCd(args []string, ctx *Context) bool {
	if len(args) < 1 {
		fmt.Println("usage: cd <name>")
		return false
	}

	path := util.ResolvePath(*ctx.CurrentDir, args[0])

	// Make sure the destination actually exists and is a directory.
	info, err := os.Stat(path)
	if err != nil {
		fmt.Println("error:", err)
		return false
	}

	if !info.IsDir() {
		fmt.Println("not a directory:", path)
		return false
	}

	*ctx.CurrentDir = path

	return false
}

// handleRemove deletes one or more files or folders.
//
// Files use os.Remove.
// Directories use os.RemoveAll.
func handleRemove(args []string, ctx *Context) bool {
	if len(args) < 1 {
		fmt.Println("usage: remove <name> [name2 ...]")
		return false
	}

	for _, name := range args {
		path := util.ResolvePath(*ctx.CurrentDir, name)

		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("does not exist:", path)
			} else {
				fmt.Println("error checking:", path, "-", err)
			}

			continue
		}

		if info.IsDir() {
			if err := os.RemoveAll(path); err != nil {
				fmt.Println("error removing folder:", path, "-", err)
				continue
			}

			fmt.Println("removed folder:", path)
			continue
		}

		if err := os.Remove(path); err != nil {
			fmt.Println("error removing file:", path, "-", err)
			continue
		}

		fmt.Println("removed file:", path)
	}

	return false
}

// handleCopy supports:
//
//	copy <src> <dest>
//	copy <src> <dest> <startLine> <endLine>
func handleCopy(args []string, ctx *Context) bool {
	switch len(args) {
	case 2:
		copyWhole(*ctx.CurrentDir, args[0], args[1])

	case 4:
		copyLineRange(
			*ctx.CurrentDir,
			args[0],
			args[1],
			args[2],
			args[3],
		)

	default:
		fmt.Println("usage: copy <src> <dest>")
		fmt.Println("       copy <src> <dest> <startLine> <endLine>")
	}

	return false
}

// copyWhole copies either a file or an entire directory.
func copyWhole(currentDir, srcArg, destArg string) {
	src := util.ResolvePath(currentDir, srcArg)
	dest := util.ResolvePath(currentDir, destArg)

	info, err := os.Stat(src)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	if info.IsDir() {
		if err := copyDir(src, dest); err != nil {
			fmt.Println("error copying folder:", err)
			return
		}

		fmt.Println("copied folder:", src, "->", dest)
		return
	}

	if err := copyFile(src, dest); err != nil {
		fmt.Println("error copying file:", err)
		return
	}

	fmt.Println("copied file:", src, "->", dest)
}

// copyFile streams bytes from source to destination.
func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)

	return err
}

// copyDir recursively copies a directory tree.
func copyDir(src, dest string) error {
	return filepath.WalkDir(
		src,
		func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			relPath, err := filepath.Rel(src, path)
			if err != nil {
				return err
			}

			targetPath := filepath.Join(dest, relPath)

			if d.IsDir() {
				return os.MkdirAll(targetPath, 0755)
			}

			return copyFile(path, targetPath)
		},
	)
}

// copyLineRange copies lines [start,end] from a text file.
func copyLineRange(
	currentDir,
	srcArg,
	destArg,
	startArg,
	endArg string,
) {
	src := util.ResolvePath(currentDir, srcArg)
	dest := util.ResolvePath(currentDir, destArg)

	start, err1 := strconv.Atoi(startArg)
	end, err2 := strconv.Atoi(endArg)

	if err1 != nil ||
		err2 != nil ||
		start < 1 ||
		end < start {
		fmt.Println("invalid line range:", startArg, endArg)
		return
	}

	in, err := os.Open(src)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer out.Close()

	writer := bufio.NewWriter(out)
	defer writer.Flush()

	scanner := bufio.NewScanner(in)

	lineNum := 0
	copied := 0

	for scanner.Scan() {
		lineNum++

		if lineNum >= start && lineNum <= end {
			fmt.Fprintln(writer, scanner.Text())
			copied++
		}

		if lineNum > end {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("read error:", err)
		return
	}

	fmt.Printf(
		"copied lines %d-%d (%d lines) from %s -> %s\n",
		start,
		end,
		copied,
		src,
		dest,
	)
}

// handleCat prints the contents of a file.
//
// Usage:
//
//	cat <file>
//	cat -n <file>
func handleCat(args []string, ctx *Context) bool {
	if len(args) < 1 {
		fmt.Println("usage: cat <file>")
		return false
	}

	showLineNumbers := false
	fileArg := args[0]

	if args[0] == "-n" {
		if len(args) < 2 {
			fmt.Println("usage: cat -n <file>")
			return false
		}

		showLineNumbers = true
		fileArg = args[1]
	}

	path := util.ResolvePath(*ctx.CurrentDir, fileArg)

	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("error:", err)
		return false
	}

	if !showLineNumbers {
		fmt.Println(string(data))
		return false
	}

	lines := strings.Split(string(data), "\n")

	for i, line := range lines {
		fmt.Printf("%4d  %s\n", i+1, line)
	}

	return false
}

// handleWrite is Zebra's minimal interactive editor.
//
// :wq = save
// :q  = discard
func handleWrite(args []string, ctx *Context) bool {
	if len(args) < 1 {
		fmt.Println("usage: write <file>")
		return false
	}

	path := util.ResolvePath(*ctx.CurrentDir, args[0])

	fmt.Println(
		"-- writing",
		path,
		"-- type :wq to save, :q to cancel --",
	)

	var lines []string
	lineNum := 1

	for {
		fmt.Printf("%3d | ", lineNum)

		if !ctx.Scanner.Scan() {
			break
		}

		text := ctx.Scanner.Text()

		if text == ":wq" {
			break
		}

		if text == ":q" {
			fmt.Println("discarded, nothing saved")
			return false
		}

		lines = append(lines, text)
		lineNum++
	}

	content := strings.Join(lines, "\n")

	if len(lines) > 0 {
		content += "\n"
	}

	if err := os.WriteFile(
		path,
		[]byte(content),
		0644,
	); err != nil {
		fmt.Println("error saving:", err)
		return false
	}

	fmt.Printf(
		"saved %d lines to %s\n",
		len(lines),
		path,
	)

	return false
}

// handleMove moves a file or directory from source to destination.
//
// Usage:
//
//	move <source> <destination>
//	mv    <source> <destination>
func handleMove(args []string, ctx *Context) bool {
	if len(args) != 2 {
		fmt.Println("usage: move <source> <destination>")
		return false
	}

	src := util.ResolvePath(*ctx.CurrentDir, args[0])
	dest := util.ResolvePath(*ctx.CurrentDir, args[1])

	// Check that the source exists.
	info, err := os.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("source does not exist:", src)
		} else {
			fmt.Println("error checking source:", err)
		}
		return false
	}

	// Prevent moving a directory into itself.
	if info.IsDir() {
		srcAbs, err := filepath.Abs(src)
		if err != nil {
			fmt.Println("error resolving source:", err)
			return false
		}

		destAbs, err := filepath.Abs(dest)
		if err != nil {
			fmt.Println("error resolving destination:", err)
			return false
		}

		rel, err := filepath.Rel(srcAbs, destAbs)
		if err == nil && rel != "." && rel != ".." &&
			!strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			fmt.Println("cannot move a folder into itself")
			return false
		}
	}

	// Make sure the destination's parent directory exists.
	parent := filepath.Dir(dest)

	if err := os.MkdirAll(parent, 0755); err != nil {
		fmt.Println("error creating destination directory:", err)
		return false
	}

	// If destination already exists, don't silently overwrite it.
	if _, err := os.Stat(dest); err == nil {
		fmt.Println("destination already exists:", dest)
		return false
	}

	if err := os.Rename(src, dest); err != nil {
		fmt.Println("error moving:", err)
		return false
	}

	if info.IsDir() {
		fmt.Println("moved folder:", src, "->", dest)
	} else {
		fmt.Println("moved file:", src, "->", dest)
	}

	return false
}
