package commands

import (
	"fmt"
	"os"
	"path/filepath"
)

func HandleFolder(currentDir *string, args []string) {
	if len(args) < 1 {
		fmt.Println("usage: folder <name> [name2 ...]")
		return
	}
	for _, name := range args {
		path := filepath.Join(*currentDir, name)
		if err := os.MkdirAll(path, 0755); err != nil {
			fmt.Println("error creating folder:", err)
			continue
		}
		fmt.Println("created folder:", path)
	}
	*currentDir = filepath.Join(*currentDir, args[0])
}

func HandleFile(currentDir string, args []string) {
	if len(args) < 1 {
		fmt.Println("usage: file <name> [name2 ...]")
		return
	}
	for _, name := range args {
		path := filepath.Join(currentDir, name)
		f, err := os.Create(path)
		if err != nil {
			fmt.Println("error creating file:", err)
			continue
		}
		f.Close()
		fmt.Println("created file:", path)
	}
}

func HandleCd(currentDir *string, args []string) {
	if len(args) < 1 {
		return
	}
	*currentDir = filepath.Join(*currentDir, args[0])
}

// HandleRemove deletes one or more files or folders by name.
// It checks each path first to decide which removal method fits:
// a plain file (or empty folder) uses os.Remove, a folder with
// contents needs os.RemoveAll to delete recursively.
func HandleRemove(currentDir string, args []string) {
	if len(args) < 1 {
		fmt.Println("usage: remove <name> [name2 ...]")
		return
	}

	for _, name := range args {
		path := filepath.Join(currentDir, name)

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
		} else {
			if err := os.Remove(path); err != nil {
				fmt.Println("error removing file:", path, "-", err)
				continue
			}
			fmt.Println("removed file:", path)
		}
	}
}
