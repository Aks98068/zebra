package commands

import "fmt"

var registry = make(map[string]Command)

// Register adds a command to the registry.
func Register(command Command) {
	registry[command.Name] = command

	for _, alias := range command.Aliases {
		registry[alias] = command
	}
}

// Execute finds and runs a registered command.
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

// Init registers all built-in commands.
func Init() {
	Register(Command{
		Name:        "folder",
		Description: "create one or more folders",
		Usage:       "folder <name> [name2 ...]",
		Run:         handleFolder,
	})

	Register(Command{
		Name:        "file",
		Description: "create one or more files",
		Usage:       "file <name> [name2 ...]",
		Run:         handleFile,
	})

	Register(Command{
		Name:        "cd",
		Description: "move into a folder",
		Usage:       "cd <name>",
		Run:         handleCd,
	})

	Register(Command{
		Name:        "remove",
		Aliases:     []string{"rm"},
		Description: "delete a file or folder recursively",
		Usage:       "remove <name> [name2 ...]",
		Run:         handleRemove,
	})

	Register(Command{
		Name:        "copy",
		Aliases:     []string{"cp"},
		Description: "copy a file or folder",
		Usage:       "copy <src> <dest> [start end]",
		Run:         handleCopy,
	})

	Register(Command{
		Name:        "cat",
		Aliases:     []string{"read"},
		Description: "print a file's content",
		Usage:       "cat <file> [-n]",
		Run:         handleCat,
	})

	Register(Command{
		Name:        "write",
		Aliases:     []string{"nano", "vim"},
		Description: "open the mini editor",
		Usage:       "write <file>",
		Run:         handleWrite,
	})

	Register(Command{
		Name:        "pwd",
		Description: "show current folder",
		Usage:       "pwd",
		Run:         handlePwd,
	})

	Register(Command{
		Name:        "help",
		Description: "show available commands",
		Usage:       "help",
		Run:         handleHelp,
	})

	Register(Command{
		Name:        "exit",
		Aliases:     []string{"quit"},
		Description: "leave Zebra",
		Usage:       "exit",
		Run:         handleExit,
	})

	Register(Command{
		Name:        "move",
		Aliases:     []string{"mv"},
		Description: "move a file or folder",
		Usage:       "move <source> <destination>",
		Run:         handleMove,
	})
	Register(Command{
	Name:        "get -A environmentvariables",
	Description: "get all environment variables of the system",
	Usage:       "get -A environmentvariables",
	Run:         getAllEnvironmentVariablesCommand,
})

Register(Command{
	Name:        "get",
	Description: "get a particular environment variable of the system",
	Usage:       "get <environment-variable-name>",
	Run:         getEnvironmentVariableCommand,
})
}

Register(Command{
	Name:        "unset",
	Description: "remove a persistent environment variable",
	Usage:       "unset <environment-variable-name>",
	Run:         unsetEnvironmentVariableCommand,
})
