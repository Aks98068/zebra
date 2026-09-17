package ui

import "fmt"

func PrintHelp() {

	fmt.Println()
	fmt.Println("ZEBRA COMMANDS")
	fmt.Println()

	fmt.Println("Filesystem:")
	fmt.Println("  folder <path>                         Create a folder")
	fmt.Println("  file <path>                           Create a file")
	fmt.Println("  cd <path>                             Change directory")
	fmt.Println("  pwd                                   Show current directory")
	fmt.Println("  remove <path>                         Remove file/folder")
	fmt.Println("  rm <path>                             Alias for remove")
	fmt.Println("  copy <source> <destination>           Copy file")
	fmt.Println("  cp <source> <destination>             Alias for copy")
	fmt.Println("  move <source> <destination>           Move file/folder")
	fmt.Println("  mv <source> <destination>             Alias for move")
	fmt.Println("  cat <file>                            Read file")
	fmt.Println("  read <file>                           Alias for cat")
	fmt.Println("  write <file>                          Write file")
	fmt.Println("  nano <file>                           Alias for write")
	fmt.Println("  vim <file>                            Alias for write")

	fmt.Println()

	fmt.Println("Environment:")
	fmt.Println("  get <name>                             Get environment variable")
	fmt.Println("  get -A environmentvariables            Get all environment variables")
	fmt.Println("  set <name> <value>                     Set persistent environment variable")
	fmt.Println("  unset <name>                           Remove persistent environment variable")

	fmt.Println()

	fmt.Println("General:")
	fmt.Println("  help                                   Show this help")
	fmt.Println("  exit                                   Exit Zebra")
	fmt.Println("  quit                                   Alias for exit")

	fmt.Println()
}