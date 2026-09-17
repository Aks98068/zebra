package util

import "path/filepath"

// ResolvePath decides whether input is already a full/absolute path
// (e.g. C:\Users\abishek\test.txt or /home/abishek/test.txt) or just
// a bare name/relative path (e.g. test.txt, sub/test.txt).
//
// If it's absolute, it's used exactly as given -- no changes.
// If it's relative, it's joined onto currentDir, same as how cmd.exe
// or bash resolves a relative path against your current directory.
//
// Every command (folder, file, remove, rename, move, cd, ...) should
// run each incoming path argument through this before touching the
// filesystem, instead of each file reimplementing the same check.
func ResolvePath(currentDir, input string) string {
	if filepath.IsAbs(input) {
		return input
	}
	return filepath.Join(currentDir, input)
}
