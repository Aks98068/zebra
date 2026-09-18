package ui

import "fmt"

func PrintHelp() {
	fmt.Println()

	fmt.Println(
		BrightCyan + Bold +
			"ZEBRA COMMANDS" +
			Reset,
	)

	fmt.Println()

	// ========================================================
	// FILESYSTEM
	// ========================================================

	fmt.Println(
		BrightYellow + Bold +
			"Filesystem:" +
			Reset,
	)

	fmt.Println(
		"  folder <path>                         Create a folder",
	)

	fmt.Println(
		"  file <path>                           Create a file",
	)

	fmt.Println(
		"  cd <path>                             Change directory",
	)

	fmt.Println(
		"  pwd                                   Show current directory",
	)

	fmt.Println(
		"  ls [path]                             List files and folders",
	)

	fmt.Println(
		"  remove <path>                         Remove file/folder",
	)

	fmt.Println(
		"  rm <path>                             Alias for remove",
	)

	fmt.Println(
		"  copy <source> <destination>           Copy file",
	)

	fmt.Println(
		"  cp <source> <destination>             Alias for copy",
	)

	fmt.Println(
		"  move <source> <destination>           Move file/folder",
	)

	fmt.Println(
		"  mv <source> <destination>             Alias for move",
	)

	fmt.Println(
		"  cat <file>                            Read file",
	)

	fmt.Println(
		"  read <file>                           Alias for cat",
	)

	fmt.Println(
		"  write <file>                          Write file",
	)

	fmt.Println(
		"  nano <file>                           Alias for write",
	)

	fmt.Println(
		"  vim <file>                            Alias for write",
	)

	fmt.Println()

	// ========================================================
	// ENVIRONMENT
	// ========================================================

	fmt.Println(
		BrightYellow + Bold +
			"Environment:" +
			Reset,
	)

	fmt.Println(
		"  get <name>                            Get environment variable",
	)

	fmt.Println(
		"  get -A                                Get all environment variables",
	)

	fmt.Println(
		"  set <name> <value>                    Set persistent environment variable",
	)

	fmt.Println(
		"  unset <name>                          Remove persistent environment variable",
	)

	fmt.Println()

	// ========================================================
	// PROCESS
	// ========================================================

	fmt.Println(
		BrightYellow + Bold +
			"Process:" +
			Reset,
	)

	fmt.Println(
		"  ps                                    List all running processes",
	)

	fmt.Println(
		"  procs                                 Alias for ps",
	)

	fmt.Println(
		"  ppid <pid>                            Find a process by PID",
	)

	fmt.Println(
		"  pgrep <name>                          Search processes by name",
	)

	fmt.Println(
		"  kill <pid>                            Kill a process by PID",
	)

	fmt.Println(
		"  pkill <name>                          Kill all processes matching name",
	)

	fmt.Println(
		"  killself                              Terminate the current process",
	)

	fmt.Println()

	// ========================================================
	// GENERAL
	// ========================================================

	fmt.Println(
		BrightYellow + Bold +
			"General:" +
			Reset,
	)

	fmt.Println(
		"  help                                  Show this help",
	)

	fmt.Println(
		"  terminal                              Open a new Zebra terminal",
	)

	fmt.Println(
		"  exit                                  Exit Zebra",
	)

	fmt.Println(
		"  quit                                  Alias for exit",
	)


	// ========================================================
// SYMLINKS
// ========================================================

fmt.Println(BrightYellow + Bold + "Symlinks & Links:" + Reset)

fmt.Println("  symlink <target> <link>               Create a symbolic link")
fmt.Println("  ln-s <target> <link>                  Alias for symlink")
fmt.Println("  readlink <link>                       Read symlink target")
fmt.Println("  hardlink <source> <destination>       Create a hard link")
fmt.Println("  ln <source> <destination>             Alias for hardlink")

fmt.Println()

// ========================================================
// PERMISSIONS
// ========================================================

fmt.Println(BrightYellow + Bold + "Permissions & Ownership:" + Reset)

fmt.Println("  chmod <permissions> <file>            Change file permissions")
fmt.Println("  chown <uid> <gid> <file>              Change file owner")
fmt.Println("  lchown <uid> <gid> <link>             Change symlink owner")

fmt.Println()

// ========================================================
// STDIO
// ========================================================

fmt.Println(BrightYellow + Bold + "Stdio:" + Reset)

fmt.Println("  stdout <message>                      Write to standard output")
fmt.Println("  stderr <message>                      Write to standard error")
fmt.Println("  stdin                                 Read from standard input")

fmt.Println()

// ========================================================
// SIGNALS
// ========================================================

fmt.Println(BrightYellow + Bold + "Signals & Exit:" + Reset)

fmt.Println("  signals                               Listen for OS signals")
fmt.Println("  exitcode <code>                       Exit with specific code")

fmt.Println()

// ========================================================
// SYSTEM INFO
// ========================================================

fmt.Println(BrightYellow + Bold + "System Info:" + Reset)

fmt.Println("  sysinfo                               Show all system info")
fmt.Println("  hostname                              Show machine hostname")
fmt.Println("  homedir                               Show user home directory")
fmt.Println("  cachedir                              Show user cache directory")
fmt.Println("  configdir                             Show user config directory")

fmt.Println()

// ========================================================
// TEMP FILES
// ========================================================

fmt.Println(BrightYellow + Bold + "Temp Files:" + Reset)

fmt.Println("  tmpfile [prefix]                      Create a temporary file")
fmt.Println("  tmpdir [prefix]                       Create a temporary directory")
fmt.Println("  tempdir                               Show system temp directory")

fmt.Println()

}