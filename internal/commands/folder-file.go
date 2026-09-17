
package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ============================================================
// PATH RESOLUTION
// ============================================================
//
// Converts a Zebra path into an absolute filesystem path.
//
// Supported:
//
//     file.txt
//     ./file.txt
//     ../file.txt
//     ~/Documents
//     C:\Users\ACER
//     D:\Projects
//
// Relative paths are resolved against Zebra's virtual
// CurrentDir.
//
// ============================================================

func resolvePath(path string, ctx *Context) string {
	path = strings.TrimSpace(path)

	if path == "" {
		return *ctx.CurrentDir
	}

	// --------------------------------------------------------
	// Expand ~
	// --------------------------------------------------------

	if path == "~" {
		home, err := os.UserHomeDir()

		if err == nil {
			return filepath.Clean(home)
		}
	}

	// ~/something
	//
	// filepath.IsAbs does not treat "~" specially, so expand
	// it before checking for an absolute path.

	if strings.HasPrefix(path, "~/") ||
		strings.HasPrefix(path, `~\`) {

		home, err := os.UserHomeDir()

		if err == nil {
			path = filepath.Join(
				home,
				path[2:],
			)
		}
	}

	// --------------------------------------------------------
	// Absolute path
	// --------------------------------------------------------

	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}

	// --------------------------------------------------------
	// Relative path
	// --------------------------------------------------------

	return filepath.Clean(
		filepath.Join(
			*ctx.CurrentDir,
			path,
		),
	)
}

// ============================================================
// FOLDER / MKDIR
// ============================================================
//
// Examples:
//
//     folder test
//     folder one two three
//     mkdir src tests docs
//
// Creates every supplied directory.
//
// ============================================================

func handleFolder(args []string, ctx *Context) bool {
	if len(args) == 0 {
		fmt.Println("Usage: folder <path> [path ...]")
		return false
	}

	var created int
	var failed int

	for _, argument := range args {
		path := resolvePath(argument, ctx)

		err := os.MkdirAll(path, 0755)

		if err != nil {
			fmt.Printf(
				"✗ Failed: %s\n  %v\n",
				path,
				err,
			)

			failed++
			continue
		}

		fmt.Printf(
			"✓ Folder created: %s\n",
			path,
		)

		created++
	}

	fmt.Println()

	fmt.Printf(
		"Created: %d folder(s)",
		created,
	)

	if failed > 0 {
		fmt.Printf(
			"    Failed: %d",
			failed,
		)
	}

	fmt.Println()

	return false
}

// ============================================================
// FILE / TOUCH
// ============================================================
//
// Examples:
//
//     file a.txt
//     file a.txt b.txt c.txt
//     touch one.txt two.txt three.txt
//
// Creates every supplied file.
//
// Existing files are NOT deleted. os.OpenFile with O_CREATE
// and O_APPEND is used so an existing file remains intact.
//
// ============================================================

func handleFile(args []string, ctx *Context) bool {
	if len(args) == 0 {
		fmt.Println("Usage: file <path> [path ...]")
		return false
	}

	var created int
	var failed int

	for _, argument := range args {
		path := resolvePath(argument, ctx)

		file, err := os.OpenFile(
			path,
			os.O_CREATE|os.O_WRONLY,
			0644,
		)

		if err != nil {
			fmt.Printf(
				"✗ Failed: %s\n  %v\n",
				path,
				err,
			)

			failed++
			continue
		}

		if err := file.Close(); err != nil {
			fmt.Printf(
				"✗ Failed closing: %s\n  %v\n",
				path,
				err,
			)

			failed++
			continue
		}

		fmt.Printf(
			"✓ File created: %s\n",
			path,
		)

		created++
	}

	fmt.Println()

	fmt.Printf(
		"Created: %d file(s)",
		created,
	)

	if failed > 0 {
		fmt.Printf(
			"    Failed: %d",
			failed,
		)
	}

	fmt.Println()

	return false
}

// ============================================================
// CD
// ============================================================
//
//     cd folder
//     cd ..
//     cd ~
//     cd C:\
//     cd D:\Projects
//
// Changes Zebra's virtual current directory.
//
// ============================================================

func handleCd(args []string, ctx *Context) bool {
	if len(args) == 0 {
		fmt.Println("Usage: cd <path>")
		return false
	}

	if len(args) > 1 {
		fmt.Println("Usage: cd <path>")
		return false
	}

	path := resolvePath(
		args[0],
		ctx,
	)

	path = filepath.Clean(path)

	info, err := os.Stat(path)

	if err != nil {
		fmt.Println("Error:", err)
		return false
	}

	if !info.IsDir() {
		fmt.Println("Not a directory:", path)
		return false
	}

	absolutePath, err := filepath.Abs(path)

	if err != nil {
		fmt.Println("Error:", err)
		return false
	}

	*ctx.CurrentDir = filepath.Clean(
		absolutePath,
	)

	return false
}

// ============================================================
// REMOVE / RM
// ============================================================
//
// Examples:
//
//     remove file.txt
//     remove a.txt b.txt c.txt
//     remove folder1 folder2
//
// Directories are removed recursively.
//
// ============================================================

func handleRemove(args []string, ctx *Context) bool {
	if len(args) == 0 {
		fmt.Println("Usage: remove <path> [path ...]")
		return false
	}

	var removed int
	var failed int

	for _, argument := range args {
		path := resolvePath(argument, ctx)

		// ----------------------------------------------------
		// Safety check
		// ----------------------------------------------------

		if isDangerousRemovePath(
			path,
			ctx,
		) {
			fmt.Printf(
				"✗ Refused to remove protected path: %s\n",
				path,
			)

			failed++
			continue
		}

		// ----------------------------------------------------
		// Check existence
		// ----------------------------------------------------

		_, err := os.Lstat(path)

		if err != nil {
			fmt.Printf(
				"✗ Not found: %s\n",
				path,
			)

			failed++
			continue
		}

		// ----------------------------------------------------
		// Remove recursively
		// ----------------------------------------------------

		err = os.RemoveAll(path)

		if err != nil {
			fmt.Printf(
				"✗ Failed removing: %s\n  %v\n",
				path,
				err,
			)

			failed++
			continue
		}

		fmt.Printf(
			"✓ Removed: %s\n",
			path,
		)

		removed++
	}

	fmt.Println()

	fmt.Printf(
		"Removed: %d item(s)",
		removed,
	)

	if failed > 0 {
		fmt.Printf(
			"    Failed: %d",
			failed,
		)
	}

	fmt.Println()

	return false
}

// ============================================================
// COPY
// ============================================================
//
// Supported:
//
//     copy a.txt backup.txt
//
//     copy a.txt b.txt backup
//
//     copy folder1 backup
//
//     copy folder1 folder2 backup
//
// Multiple sources require the destination to be a directory.
//
// Directories are copied recursively.
//
// ============================================================

func handleCopy(args []string, ctx *Context) bool {
	if len(args) < 2 {
		fmt.Println(
			"Usage: copy <source> [source ...] <destination>",
		)

		return false
	}

	sources := args[:len(args)-1]

	destination := resolvePath(
		args[len(args)-1],
		ctx,
	)

	// --------------------------------------------------------
	// Resolve all source paths
	// --------------------------------------------------------

	resolvedSources := make([]string, 0, len(sources))

	for _, source := range sources {
		resolvedSources = append(
			resolvedSources,
			resolvePath(source, ctx),
		)
	}

	// --------------------------------------------------------
	// Multiple sources
	// --------------------------------------------------------

	if len(resolvedSources) > 1 {
		info, err := os.Stat(destination)

		if err != nil {
			fmt.Printf(
				"✗ Destination must be an existing directory when copying multiple sources:\n  %s\n",
				destination,
			)

			return false
		}

		if !info.IsDir() {
			fmt.Println(
				"✗ Destination must be a directory:",
				destination,
			)

			return false
		}
	}

	var copied int
	var failed int

	for _, source := range resolvedSources {
		target := destination

		// ----------------------------------------------------
		// If destination is a directory, place the source
		// inside it.
		// ----------------------------------------------------

		if info, err := os.Stat(destination); err == nil &&
			info.IsDir() {

			target = filepath.Join(
				destination,
				filepath.Base(source),
			)
		}

		// ----------------------------------------------------
		// Prevent copying an item into itself
		// ----------------------------------------------------

		if samePath(source, target) {
			fmt.Printf(
				"✗ Cannot copy item onto itself: %s\n",
				source,
			)

			failed++
			continue
		}

		// ----------------------------------------------------
		// Check source
		// ----------------------------------------------------

		info, err := os.Stat(source)

		if err != nil {
			fmt.Printf(
				"✗ Source not found: %s\n",
				source,
			)

			failed++
			continue
		}

		// ----------------------------------------------------
		// Copy
		// ----------------------------------------------------

		if info.IsDir() {
			err = copyDirectory(
				source,
				target,
			)
		} else {
			err = copyFile(
				source,
				target,
			)
		}

		if err != nil {
			fmt.Printf(
				"✗ Copy failed:\n  %s\n  → %s\n  %v\n",
				source,
				target,
				err,
			)

			failed++
			continue
		}

		fmt.Printf(
			"✓ Copied:\n  %s\n  → %s\n",
			source,
			target,
		)

		copied++
	}

	fmt.Println()

	fmt.Printf(
		"Copied: %d item(s)",
		copied,
	)

	if failed > 0 {
		fmt.Printf(
			"    Failed: %d",
			failed,
		)
	}

	fmt.Println()

	return false
}

// ============================================================
// MOVE
// ============================================================
//
// Supported:
//
//     move a.txt backup.txt
//
//     move a.txt b.txt backup
//
//     move folder1 folder2 archive
//
// Multiple sources require the destination to be a directory.
//
// ============================================================

func handleMove(args []string, ctx *Context) bool {
	if len(args) < 2 {
		fmt.Println(
			"Usage: move <source> [source ...] <destination>",
		)

		return false
	}

	sources := args[:len(args)-1]

	destination := resolvePath(
		args[len(args)-1],
		ctx,
	)

	// --------------------------------------------------------
	// Multiple sources require directory destination
	// --------------------------------------------------------

	if len(sources) > 1 {
		info, err := os.Stat(destination)

		if err != nil || !info.IsDir() {
			fmt.Println(
				"✗ Destination must be an existing directory:",
				destination,
			)

			return false
		}
	}

	var moved int
	var failed int

	for _, argument := range sources {
		source := resolvePath(
			argument,
			ctx,
		)

		target := destination

		// ----------------------------------------------------
		// Destination directory
		// ----------------------------------------------------

		if info, err := os.Stat(destination); err == nil &&
			info.IsDir() {

			target = filepath.Join(
				destination,
				filepath.Base(source),
			)
		}

		// ----------------------------------------------------
		// Prevent moving item into itself
		// ----------------------------------------------------

		if samePath(source, target) {
			fmt.Printf(
				"✗ Cannot move item onto itself: %s\n",
				source,
			)

			failed++
			continue
		}

		// ----------------------------------------------------
		// Check source
		// ----------------------------------------------------

		_, err := os.Lstat(source)

		if err != nil {
			fmt.Printf(
				"✗ Source not found: %s\n",
				source,
			)

			failed++
			continue
		}

		// ----------------------------------------------------
		// Move
		// ----------------------------------------------------

		err = os.Rename(
			source,
			target,
		)

		if err != nil {

			// ------------------------------------------------
			// Cross-device fallback.
			//
			// os.Rename can fail when source and destination
			// are on different drives/filesystems.
			// ------------------------------------------------

			err = moveFallback(
				source,
				target,
			)

			if err == nil {
				err = os.RemoveAll(source)
			}
		}

		if err != nil {
			fmt.Printf(
				"✗ Move failed:\n  %s\n  → %s\n  %v\n",
				source,
				target,
				err,
			)

			failed++
			continue
		}

		fmt.Printf(
			"✓ Moved:\n  %s\n  → %s\n",
			source,
			target,
		)

		moved++
	}

	fmt.Println()

	fmt.Printf(
		"Moved: %d item(s)",
		moved,
	)

	if failed > 0 {
		fmt.Printf(
			"    Failed: %d",
			failed,
		)
	}

	fmt.Println()

	return false
}

// ============================================================
// COPY FILE
// ============================================================

func copyFile(
	source string,
	destination string,
) error {

	sourceFile, err := os.Open(source)

	if err != nil {
		return err
	}

	defer sourceFile.Close()

	// --------------------------------------------------------
	// Ensure destination parent exists
	// --------------------------------------------------------

	parent := filepath.Dir(destination)

	if err := os.MkdirAll(
		parent,
		0755,
	); err != nil {
		return err
	}

	// --------------------------------------------------------
	// Get source permissions
	// --------------------------------------------------------

	info, err := sourceFile.Stat()

	if err != nil {
		return err
	}

	// --------------------------------------------------------
	// Create destination
	// --------------------------------------------------------

	destinationFile, err := os.OpenFile(
		destination,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		info.Mode().Perm(),
	)

	if err != nil {
		return err
	}

	// --------------------------------------------------------
	// Copy contents
	// --------------------------------------------------------

	_, copyErr := io.Copy(
		destinationFile,
		sourceFile,
	)

	closeErr := destinationFile.Close()

	if copyErr != nil {
		return copyErr
	}

	if closeErr != nil {
		return closeErr
	}

	return nil
}

// ============================================================
// COPY DIRECTORY
// ============================================================
//
// Recursive directory copy.
//
// Example:
//
//     copy project backup
//
// Produces:
//
//     backup/project/...
//
// ============================================================

func copyDirectory(
	source string,
	destination string,
) error {

	info, err := os.Stat(source)

	if err != nil {
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf(
			"source is not a directory",
		)
	}

	// --------------------------------------------------------
	// Create destination directory
	// --------------------------------------------------------

	if err := os.MkdirAll(
		destination,
		info.Mode().Perm(),
	); err != nil {
		return err
	}

	entries, err := os.ReadDir(source)

	if err != nil {
		return err
	}

	for _, entry := range entries {
		sourcePath := filepath.Join(
			source,
			entry.Name(),
		)

		destinationPath := filepath.Join(
			destination,
			entry.Name(),
		)

		entryInfo, err := entry.Info()

		if err != nil {
			return err
		}

		if entryInfo.IsDir() {
			if err := copyDirectory(
				sourcePath,
				destinationPath,
			); err != nil {
				return err
			}

			continue
		}

		if err := copyFile(
			sourcePath,
			destinationPath,
		); err != nil {
			return err
		}
	}

	return nil
}

// ============================================================
// MOVE FALLBACK
// ============================================================
//
// Used when os.Rename cannot move across filesystems/drives.
//
// Strategy:
//
//     file      → copy + remove
//     directory → recursive copy + remove
//
// ============================================================

func moveFallback(
	source string,
	destination string,
) error {

	info, err := os.Stat(source)

	if err != nil {
		return err
	}

	if info.IsDir() {
		return copyDirectory(
			source,
			destination,
		)
	}

	return copyFile(
		source,
		destination,
	)
}

// ============================================================
// SAME PATH
// ============================================================

func samePath(
	first string,
	second string,
) bool {

	firstAbs, err := filepath.Abs(first)

	if err != nil {
		firstAbs = filepath.Clean(first)
	}

	secondAbs, err := filepath.Abs(second)

	if err != nil {
		secondAbs = filepath.Clean(second)
	}

	firstAbs = filepath.Clean(firstAbs)
	secondAbs = filepath.Clean(secondAbs)

	if samePathCaseInsensitive() {
		return strings.EqualFold(
			firstAbs,
			secondAbs,
		)
	}

	return firstAbs == secondAbs
}

// ============================================================
// PLATFORM PATH COMPARISON
// ============================================================

func samePathCaseInsensitive() bool {
	return os.PathSeparator == '\\'
}

// ============================================================
// PROTECTED REMOVE PATH
// ============================================================
//
// Prevents:
//
//     remove C:\
//     remove /
//     remove ~
//
// This is deliberately conservative.
//
// ============================================================

func isDangerousRemovePath(
	path string,
	ctx *Context,
) bool {

	cleanPath := filepath.Clean(path)

	// --------------------------------------------------------
	// Zebra current directory
	// --------------------------------------------------------

	current := filepath.Clean(
		*ctx.CurrentDir,
	)

	if samePath(
		cleanPath,
		current,
	) {
		return true
	}

	// --------------------------------------------------------
	// User home
	// --------------------------------------------------------

	home, err := os.UserHomeDir()

	if err == nil &&
		samePath(cleanPath, home) {
		return true
	}

	// --------------------------------------------------------
	// Filesystem root
	// --------------------------------------------------------

	volumeRoot := filepath.VolumeName(cleanPath)

	if volumeRoot != "" {
		root := volumeRoot +
			string(os.PathSeparator)

		if samePath(cleanPath, root) {
			return true
		}
	}

	if cleanPath == string(os.PathSeparator) {
		return true
	}

	return false
}

// ============================================================
// LS / DIR / LIST
// ============================================================
//
//     ls
//     ls folder
//     dir
//     list
//
// ============================================================

func handleLs(args []string, ctx *Context) bool {

	path := *ctx.CurrentDir

	if len(args) > 0 {
		path = resolvePath(
			args[0],
			ctx,
		)
	}

	if len(args) > 1 {
		fmt.Println(
			"Usage: ls [path]",
		)

		return false
	}

	entries, err := os.ReadDir(path)

	if err != nil {
		fmt.Println(
			"Error:",
			err,
		)

		return false
	}

	fmt.Println()
	fmt.Println(
		"Directory:",
		path,
	)
	fmt.Println()

	fmt.Printf(
		"%-6s %-32s %-12s %-8s %-19s\n",
		"TYPE",
		"NAME",
		"SIZE",
		"PERM",
		"MODIFIED",
	)

	fmt.Println(
		strings.Repeat("─", 85),
	)

	var folders int
	var files int
	var totalSize int64

	for _, entry := range entries {

		info, err := entry.Info()

		if err != nil {
			continue
		}

		name := entry.Name()

		if entry.IsDir() {

			fmt.Printf(
				"%-6s %-32s %-12s %-8s %-19s\n",
				"DIR",
				name,
				"-",
				permissions(info),
				formatTime(info.ModTime()),
			)

			folders++

			continue
		}

		size := info.Size()

		fmt.Printf(
			"%-6s %-32s %-12s %-8s %-19s\n",
			"FILE",
			name,
			formatSize(size),
			permissions(info),
			formatTime(info.ModTime()),
		)

		files++
		totalSize += size
	}

	fmt.Println(
		strings.Repeat("─", 85),
	)

	fmt.Printf(
		"%d folders    %d files    %s\n",
		folders,
		files,
		formatSize(totalSize),
	)

	fmt.Println()

	return false
}

// ============================================================
// CAT / READ
// ============================================================

func handleCat(args []string, ctx *Context) bool {

	if len(args) == 0 {
		fmt.Println(
			"Usage: cat <file>",
		)

		return false
	}

	if len(args) > 1 {
		fmt.Println(
			"Usage: cat <file>",
		)

		return false
	}

	path := resolvePath(
		args[0],
		ctx,
	)

	data, err := os.ReadFile(path)

	if err != nil {
		fmt.Println(
			"Error reading file:",
			err,
		)

		return false
	}

	fmt.Println(
		string(data),
	)

	return false
}

// ============================================================
// WRITE
// ============================================================

func handleWrite(args []string, ctx *Context) bool {

	if len(args) == 0 {
		fmt.Println(
			"Usage: write <file>",
		)

		return false
	}

	if len(args) > 1 {
		fmt.Println(
			"Usage: write <file>",
		)

		return false
	}

	path := resolvePath(
		args[0],
		ctx,
	)

	fmt.Println(
		"Enter content.",
	)

	fmt.Println(
		"Type END on a new line to finish.",
	)

	var lines []string

	for ctx.Scanner.Scan() {

		line := ctx.Scanner.Text()

		if strings.TrimSpace(line) == "END" {
			break
		}

		lines = append(
			lines,
			line,
		)
	}

	if err := ctx.Scanner.Err(); err != nil {
		fmt.Println(
			"Input error:",
			err,
		)

		return false
	}

	content := strings.Join(
		lines,
		"\n",
	)

	err := os.WriteFile(
		path,
		[]byte(content),
		0644,
	)

	if err != nil {
		fmt.Println(
			"Error writing file:",
			err,
		)

		return false
	}

	fmt.Println(
		"File written:",
		path,
	)

	return false
}

// ============================================================
// FORMAT FILE SIZE
// ============================================================

func formatSize(size int64) string {

	if size < 1024 {
		return fmt.Sprintf(
			"%d B",
			size,
		)
	}

	if size < 1024*1024 {
		return fmt.Sprintf(
			"%.1f KB",
			float64(size)/1024,
		)
	}

	if size < 1024*1024*1024 {
		return fmt.Sprintf(
			"%.1f MB",
			float64(size)/(1024*1024),
		)
	}

	return fmt.Sprintf(
		"%.1f GB",
		float64(size)/(1024*1024*1024),
	)
}

// ============================================================
// FORMAT TIME
// ============================================================

func formatTime(t time.Time) string {
	return t.Format(
		"2006-01-02 15:04:05",
	)
}

// ============================================================
// FORMAT PERMISSIONS
// ============================================================
//
// NOTE:
//
// These are Go filesystem mode bits.
// On Windows they should NOT be interpreted as the complete
// NTFS ACL/security descriptor.
//
// ============================================================

func permissions(info os.FileInfo) string {

	mode := info.Mode()

	var result strings.Builder

	// Read
	if mode.Perm()&0400 != 0 {
		result.WriteByte('r')
	} else {
		result.WriteByte('-')
	}

	// Write
	if mode.Perm()&0200 != 0 {
		result.WriteByte('w')
	} else {
		result.WriteByte('-')
	}

	// Execute
	if mode.Perm()&0100 != 0 {
		result.WriteByte('x')
	} else {
		result.WriteByte('-')
	}

	return result.String()
}

