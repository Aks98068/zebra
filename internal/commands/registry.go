package commands

import (
	"fmt"
	"sort"
	"zebra/internal/commands/recon/dnss"
	"zebra/internal/commands/recon/whois"
)

var registrys = make(map[string]Command)

func Register(command Command) {
	registrys[command.Name] = command

	for _, alias := range command.Aliases {
		registrys[alias] = command
	}
}

func Init() {
	registrys = make(map[string]Command)

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
	// PROCESS COMMANDS
	// =========================================================

	Register(Command{
		Name:        "ps",
		Aliases:     []string{"procs", "processes"},
		Description: "list all running processes",
		Usage:       "ps",
		Run:         handleProcessList,
	})

	Register(Command{
		Name:        "ppid",
		Description: "find a process by PID",
		Usage:       "ppid <pid>",
		Run:         handleProcessPid,
	})

	Register(Command{
		Name:        "pgrep",
		Description: "search processes by name",
		Usage:       "pgrep <name>",
		Run:         handleProcessGrep,
	})

	Register(Command{
		Name:        "kill",
		Description: "kill a process by PID",
		Usage:       "kill <pid>",
		Run:         handleKillPid,
	})

	Register(Command{
		Name:        "pkill",
		Description: "kill all processes matching a name",
		Usage:       "pkill <name>",
		Run:         handleKillName,
	})

	Register(Command{
		Name:        "killself",
		Description: "terminate the current process",
		Usage:       "killself",
		Run:         handleKillSelf,
	})

	// =========================================================
	// SYMLINKS & HARD LINKS
	// =========================================================

	Register(Command{
		Name:        "symlink",
		Aliases:     []string{"ln-s"},
		Description: "create a symbolic link",
		Usage:       "symlink <target> <link>",
		Run:         handleSymlink,
	})

	Register(Command{
		Name:        "readlink",
		Description: "read a symbolic link target",
		Usage:       "readlink <link>",
		Run:         handleReadlink,
	})

	Register(Command{
		Name:        "hardlink",
		Aliases:     []string{"ln"},
		Description: "create a hard link",
		Usage:       "hardlink <source> <destination>",
		Run:         handleHardlink,
	})

	// =========================================================
	// PERMISSIONS & OWNERSHIP
	// =========================================================

	Register(Command{
		Name:        "chmod",
		Description: "change file permissions",
		Usage:       "chmod <permissions> <file>",
		Run:         handleChmod,
	})

	Register(Command{
		Name:        "chown",
		Description: "change file owner",
		Usage:       "chown <uid> <gid> <file>",
		Run:         handleChown,
	})

	Register(Command{
		Name:        "lchown",
		Description: "change symlink owner",
		Usage:       "lchown <uid> <gid> <link>",
		Run:         handleLchown,
	})

	// =========================================================
	// STDIO
	// =========================================================

	Register(Command{
		Name:        "stdout",
		Description: "write to standard output",
		Usage:       "stdout <message>",
		Run:         handleStdout,
	})

	Register(Command{
		Name:        "stderr",
		Description: "write to standard error",
		Usage:       "stderr <message>",
		Run:         handleStderr,
	})

	Register(Command{
		Name:        "stdin",
		Description: "read a line from standard input",
		Usage:       "stdin",
		Run:         handleStdin,
	})

	// =========================================================
	// SIGNALS & EXIT
	// =========================================================

	Register(Command{
		Name:        "signals",
		Description: "listen for OS signals",
		Usage:       "signals",
		Run:         handleSignals,
	})

	Register(Command{
		Name:        "exitcode",
		Description: "exit with a specific code",
		Usage:       "exitcode <code>",
		Run:         handleExitCode,
	})

	// =========================================================
	// SYSTEM INFO
	// =========================================================

	Register(Command{
		Name:        "hostname",
		Description: "show machine hostname",
		Usage:       "hostname",
		Run:         handleHostname,
	})

	Register(Command{
		Name:        "homedir",
		Description: "show user home directory",
		Usage:       "homedir",
		Run:         handleHomedir,
	})

	Register(Command{
		Name:        "cachedir",
		Description: "show user cache directory",
		Usage:       "cachedir",
		Run:         handleCachedir,
	})

	Register(Command{
		Name:        "configdir",
		Description: "show user config directory",
		Usage:       "configdir",
		Run:         handleConfigdir,
	})

	Register(Command{
		Name:        "sysinfo",
		Description: "show all system info",
		Usage:       "sysinfo",
		Run:         handleSysinfo,
	})

	// =========================================================
	// TEMP FILES
	// =========================================================

	Register(Command{
		Name:        "tmpfile",
		Description: "create a temporary file",
		Usage:       "tmpfile [prefix]",
		Run:         handleTmpfile,
	})

	Register(Command{
		Name:        "tmpdir",
		Description: "create a temporary directory",
		Usage:       "tmpdir [prefix]",
		Run:         handleTmpdir,
	})

	Register(Command{
		Name:        "tempdir",
		Description: "show system temp directory",
		Usage:       "tempdir",
		Run:         handleTempdir,
	})

	// =========================================================
	// GENERAL COMMANDS
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
		Aliases:     []string{"version"},
		Description: "display a version",
		Usage:       "zebra --version",
		Run:         zebraVersion,
	})

	// ================================
	// RECON COMMANDS
	// ================================

	Register(Command{
		Name:        "dns",
		Aliases:     []string{"dnslookup"},
		Description: "perform DNS lookup and reverse DNS lookup",
		Usage:       "dns <domain> | dns reverse <ip>",
		Run: func(args []string, ctx *Context) bool {
			return dnss.DNSLookup(args)
		},
	})

	Register(Command{
		Name:        "dns",
		Aliases:     []string{"dnslookup"},
		Description: "perform DNS lookups",
		Usage:       "dns <domain> [-type A|AAAA|MX|NS|TXT|CNAME|ALL] [-o <file>]",
		Run: func(args []string, ctx *Context) bool {
			return dnss.DNSLookup(args)
		},
	})

	Register(Command{
		Name:        "whois",
		Aliases:     []string{"whoislookup"},
		Description: "perform WHOIS lookup",
		Usage:       "whois <domain> [-server <server>] [-timeout <duration>] [-o <file>]",
		Run: func(args []string, ctx *Context) bool {
			return whois.WHOISLookup(args)
		},
	})

}

// ============================================================
// EXECUTE — only returns true for exit/quit
// ============================================================

func Execute(tokens []string, ctx *Context) bool {
	if len(tokens) == 0 {
		return false
	}

	commandName := tokens[0]
	args := tokens[1:]

	command, exists := registrys[commandName]

	if !exists {
		fmt.Println("unknown command:", commandName)
		return false
	}

	// ─── only exit and quit signal the shell to stop ──────────
	isExitCommand := commandName == "exit" || commandName == "quit"

	command.Run(args, ctx)

	return isExitCommand
}

// ============================================================
// NAMES
// ============================================================

func Names() []string {
	names := make([]string, 0, len(registrys))

	seen := make(map[string]bool)

	for _, command := range registrys {
		if seen[command.Name] {
			continue
		}

		seen[command.Name] = true
		names = append(names, command.Name)
	}

	sort.Strings(names)

	return names
}
