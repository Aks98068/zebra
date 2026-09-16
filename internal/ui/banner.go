package ui

import (
	"fmt"
	"strings"
)

// Exported (capital letters) so main.go and the commands package
// can read them -- e.g. commands.Dispatch might want the version
// string later, or another file in ui itself uses them.
const Version = "1.0.0"
const Author = "Abhishekh Kumar Sah"
const Github = "github.com/Aks98068"

func PrintBanner() {
	art := `
    ███████╗███████╗██████╗ ██████╗  █████╗
    ╚══███╔╝██╔════╝██╔══██╗██╔══██╗██╔══██╗
      ███╔╝ █████╗  ██████╔╝██████╔╝███████║
     ███╔╝  ██╔══╝  ██╔══██╗██╔══██╗██╔══██║
    ███████╗███████╗██████╔╝██║  ██║██║  ██║
    ╚══════╝╚══════╝╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝
`
	fmt.Println(art)

	width := 44
	line := strings.Repeat("─", width)
	fmt.Println("┌" + line + "┐")
	printBoxLine(width, fmt.Sprintf("Zebra CLI Tool  v%s", Version))
	printBoxLine(width, "")
	printBoxLine(width, "Author : "+Author)
	printBoxLine(width, "GitHub : "+Github)
	fmt.Println("└" + line + "┘")
	fmt.Println()
	fmt.Println("Type 'help' to see available commands.")
	fmt.Println()
}

// printBoxLine stays lowercase (unexported) on purpose -- it's a
// helper only PrintBanner needs, nothing outside the ui package
// should ever call it directly. Same package, so PrintBanner can
// still call it with no import needed.
func printBoxLine(width int, text string) {
	padding := width - len(text) - 2
	if padding < 0 {
		padding = 0
	}
	fmt.Printf("│ %s%s │\n", text, strings.Repeat(" ", padding))
}

func PrintHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  folder <name> [name2 ...]   create one or more folders")
	fmt.Println("  file <name> [name2 ...]     create one or more files in current folder")
	fmt.Println("  cd <name>                   move into a folder")
	fmt.Println("  remove <name> [name2 ...]   delete a file or folder (recursive)")
	fmt.Println("  copy <src> <dest>           copy a file or folder")
	fmt.Println("  copy <src> <dest> <a> <b>   copy only lines a-b of a file")
	fmt.Println("  read | cat <file>           print a file's content")
	fmt.Println("  write <file>                open the mini editor")
	fmt.Println("  pwd                         show current folder")
	fmt.Println("  exit / quit                 leave zebra")
}
