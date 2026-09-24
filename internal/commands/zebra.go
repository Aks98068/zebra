package commands

import "fmt"

func Info(args []string, ctx *Context) bool {

	toolname := "Zebra"
	Developer := "Abhishekh kumar sah"
	version := "1.0.0"
	github := "https://github.com/Aks98068"
	infos := `
	
### 📁 Filesystem

* Create files and directories
* Navigate directories
* List files and folders
* Copy files
* Move files and directories
* Remove files and directories
* Read files
* Write files
* Symbolic links
* Hard links
* Read symbolic-link targets
* File permissions and ownership

### ⚙️ Environment

* Read environment variables
* List all environment variables
* Set persistent environment variables
* Remove environment variables
* Manage the system PATH

### 🔄 Process Management

* List running processes
* Search processes by name
* Find processes by PID
* Terminate processes
* Terminate processes by name
* Inspect parent-child process relationships

### 🖥️ System Information

* Operating-system information
* Hostname
* Home directory
* Cache directory
* Configuration directory
* Temporary directory information

### 🌐 Reconnaissance

Zebra includes security-focused reconnaissance capabilities for authorized testing and research.

#### DNS Reconnaissance

Supports:

* A records
* AAAA records
* MX records
* NS records
* TXT records
* CNAME records
* Reverse DNS lookups
* Multiple record types`

	fmt.Println("ToolName:       ", toolname)
	fmt.Println("Developer Name: ", Developer)
	fmt.Println("Github:         ", github)
	fmt.Println("Version:        ", version)
	fmt.Println(infos)
	return false
}
