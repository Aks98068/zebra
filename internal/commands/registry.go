package commands

import (
	"fmt"
	"sort"
)

var registry = make(map[string]Command)

func Register(command Command) {
	registry[command.Name] = command

	for _, alias := range command.Aliases {
		registry[alias] = command
	}
}

func Init() {
	registry = make(map[string]Command)

	// =========================================================
	// Filesystem commands
	// =========================================================

	Register(Command{
		Name:        "folder",
		Description: "create a folder",
		Usage:       "folder <path>",
		Run:         handleFolder,
	})

	Register(Command{
		Name:        "file",
		Description: "create a file",
		Usage:       "file <path>",
		Run:         handleFile,
	})

	Register(Command{
		Name:        "cd",
		Description: "change current directory",
		Usage:       "cd <path>",
		Run:         handleCd,
	})

	Register(Command{
		Name:        "pwd",
		Description: "show current directory",
		Usage:       "pwd",
		Run:         handlePwd,
	})

	Register(Command{
		Name:        "remove",
		Aliases:     []string{"rm"},
		Description: "remove a file or folder",
		Usage:       "remove <path>",
		Run:         handleRemove,
	})

	Register(Command{
		Name:        "copy",
		Aliases:     []string{"cp"},
		Description: "copy a file",
		Usage:       "copy <source> <destination>",
		Run:         handleCopy,
	})

	Register(Command{
		Name:        "move",
		Aliases:     []string{"mv"},
		Description: "move a file or folder",
		Usage:       "move <source> <destination>",
		Run:         handleMove,
	})

	Register(Command{
		Name:        "cat",
		Aliases:     []string{"read"},
		Description: "read a file",
		Usage:       "cat <file>",
		Run:         handleCat,
	})

	Register(Command{
		Name:        "write",
		Aliases:     []string{"nano", "vim"},
		Description: "write content to a file",
		Usage:       "write <file>",
		Run:         handleWrite,
	})

	Register(Command{
		Name:        "ls",
		Aliases:     []string{"list"},
		Description: "list files and folders",
		Usage:       "ls [path]",
		Run:         handleLs,
	})

	Register(Command{
		Name:        "dir",
		Description: "list files and folders",
		Usage:       "dir [path]",
		Run:         handleLs,
	})

	// =========================================================
	// ENVIRONMENT COMMANDS
	// =========================================================

	Register(Command{
		Name:        "get",
		Description: "get an environment variable",
		Usage:       "get <name> | get -A | get --all",
		Run:         getEnvironmentVariableCommand,
	})

	Register(Command{
		Name:        "set",
		Description: "set a persistent environment variable",
		Usage:       "set <name> <value>",
		Run:         setEnvironmentVariablesCommand,
	})

	Register(Command{
		Name:        "unset",
		Description: "remove a persistent environment variable",
		Usage:       "unset <name>",
		Run:         unsetEnvironmentVariableCommand,
	})

	Register(Command{
		Name:        "path",
		Description: "view and manage PATH",
		Usage:       "path | path add <directory> | path remove <directory>",
		Run:         pathCommand,
	})

	// =========================================================
	// General commands
	// =========================================================

	Register(Command{
		Name:        "help",
		Description: "show available commands",
		Usage:       "help",
		Run:         handleHelp,
	})

	Register(Command{
		Name:        "exit",
		Aliases:     []string{"quit"},
		Description: "exit Zebra",
		Usage:       "exit",
		Run:         handleExit,
	})

	Register(Command{
		Name:        "terminal",
		Aliases:     []string{"term", "newterminal"},
		Description: "open a new Zebra terminal",
		Usage:       "terminal",
		Run:         handleTerminal,
	})

	Register(Command{
		Name:        "--version",
		Aliases:     []string{"term", "version"},
		Description: "dispaly a version",
		Usage:       "zebra --version",
		Run:         zebraVersion,
	})
}

func Execute(tokens []string, ctx *Context) bool {
	if len(tokens) == 0 {
		return false
	}

	commandName := tokens[0]
	args := tokens[1:]

	command, exists := registry[commandName]

	if !exists {
		fmt.Println("unknown command:", commandName)
		return false
	}

	return command.Run(args, ctx)
}

func Names() []string {
	names := make([]string, 0, len(registry))

	seen := make(map[string]bool)

	for _, command := range registry {
		if seen[command.Name] {
			continue
		}

		seen[command.Name] = true
		names = append(names, command.Name)
	}

	sort.Strings(names)

	return names
}
