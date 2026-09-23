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

	fmt.Println(BrightYellow + Bold + "Filesystem:" + Reset)

	fmt.Println("  folder <path>                         Create a folder")
	fmt.Println("  file <path>                           Create a file")
	fmt.Println("  cd <path>                             Change directory")
	fmt.Println("  pwd                                   Show current directory")
	fmt.Println("  ls [path]                             List files and folders")
	fmt.Println("  dir [path]                            Alias for ls")
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

	// ========================================================
	// ENVIRONMENT
	// ========================================================

	fmt.Println(BrightYellow + Bold + "Environment:" + Reset)

	fmt.Println("  get <name>                            Get environment variable")
	fmt.Println("  get -A                                Get all environment variables")
	fmt.Println("  get --all                             Get all environment variables")
	fmt.Println("  set <name> <value>                    Set persistent environment variable")
	fmt.Println("  unset <name>                          Remove persistent environment variable")
	fmt.Println("  path                                  View PATH")
	fmt.Println("  path add <directory>                  Add directory to PATH")
	fmt.Println("  path remove <directory>               Remove directory from PATH")

	fmt.Println()

	// ========================================================
	// PROCESS
	// ========================================================

	fmt.Println(BrightYellow + Bold + "Process:" + Reset)

	fmt.Println("  ps                                    List all running processes")
	fmt.Println("  procs                                 Alias for ps")
	fmt.Println("  processes                             Alias for ps")
	fmt.Println("  ppid <pid>                            Find a process by PID")
	fmt.Println("  pgrep <name>                          Search processes by name")
	fmt.Println("  kill <pid>                            Kill a process by PID")
	fmt.Println("  pkill <name>                          Kill all processes matching name")
	fmt.Println("  killself                              Terminate the current process")

	fmt.Println()

	// ========================================================
	// SYMLINKS & LINKS
	// ========================================================

	fmt.Println(BrightYellow + Bold + "Symlinks & Links:" + Reset)

	fmt.Println("  symlink <target> <link>               Create a symbolic link")
	fmt.Println("  ln-s <target> <link>                  Alias for symlink")
	fmt.Println("  readlink <link>                       Read symlink target")
	fmt.Println("  hardlink <source> <destination>       Create a hard link")
	fmt.Println("  ln <source> <destination>             Alias for hardlink")

	fmt.Println()

	// ========================================================
	// PERMISSIONS & OWNERSHIP
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
	// SIGNALS & EXIT
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
	fmt.Println("  systeminfo                            Display computer information")
	fmt.Println("  systeminformation                     Alias for systeminfo")
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

	// ========================================================
	// RECONNAISSANCE
	// ========================================================

	fmt.Println(BrightYellow + Bold + "Reconnaissance:" + Reset)

	fmt.Println("  dns <domain>                          Perform DNS lookup")
	fmt.Println("  dns reverse <ip>                      Perform reverse DNS lookup")
	fmt.Println("  dnslookup                             Alias for dns")

	fmt.Println("  whois <domain|ip>                     Perform WHOIS lookup")
	fmt.Println("  whoislookup <domain|ip>               Alias for whois")

	fmt.Println("  http <url>                            Perform HTTP/HTTPS reconnaissance")
	fmt.Println("  httprecon <url>                       Alias for http")
	fmt.Println("  http-recon <url>                      Alias for http")

	fmt.Println()

	// ========================================================
	// HTTP METHODS
	// ========================================================

	fmt.Println(BrightCyan + Bold + "HTTP Methods:" + Reset)

	fmt.Println("  GET                                   Retrieve a resource")
	fmt.Println("  POST                                  Submit data to a resource")
	fmt.Println("  PUT                                   Replace a resource")
	fmt.Println("  PATCH                                 Partially modify a resource")
	fmt.Println("  DELETE                                Delete a resource")
	fmt.Println("  HEAD                                  Retrieve response headers only")
	fmt.Println("  OPTIONS                               Query supported HTTP methods")
	fmt.Println("  TRACE                                 Perform HTTP diagnostic request")
	fmt.Println("  CONNECT                               Establish a tunnel")

	fmt.Println()

	// ========================================================
	// HTTP RECON OPTIONS
	// ========================================================

	fmt.Println(BrightCyan + Bold + "HTTP Recon Options:" + Reset)

	fmt.Println(
		"  http <url> -method <method>            Set HTTP method",
	)

	fmt.Println(
		"  http <url> -timeout <duration>         Set request timeout",
	)

	fmt.Println(
		"  http <url> -follow                     Follow redirects",
	)

	fmt.Println(
		"  http <url> -no-follow                  Disable redirect following",
	)

	fmt.Println(
		"  http <url> -max-redirects <number>     Set maximum redirects",
	)

	fmt.Println(
		"  http <url> -user-agent <string>        Set custom User-Agent",
	)

	fmt.Println(
		"  http <url> -header <key:value>         Add custom HTTP header",
	)

	fmt.Println(
		"  http <url> -proxy <url>                Use HTTP/HTTPS proxy",
	)

	fmt.Println(
		"  http <url> -insecure                   Disable TLS certificate verification",
	)

	fmt.Println(
		"  http <url> -body-limit <size>          Limit response body size",
	)

	fmt.Println(
		"  http <url> -o <file>                   Save reconnaissance results",
	)

	fmt.Println()

	// ========================================================
	// HTTP RECON INFORMATION
	// ========================================================

	fmt.Println(BrightCyan + Bold + "HTTP Recon Information:" + Reset)

	fmt.Println("  Status Code                           HTTP response status")
	fmt.Println("  Response Time                         Request/response duration")
	fmt.Println("  Protocol                              HTTP protocol version")
	fmt.Println("  Headers                               Response headers")
	fmt.Println("  Cookies                               Response cookies")
	fmt.Println("  Redirects                             Redirect chain")
	fmt.Println("  TLS                                   TLS version and cipher information")
	fmt.Println("  Certificates                          Peer certificate information")
	fmt.Println("  Security Headers                      Security header analysis")
	fmt.Println("  Page Title                            HTML page title")
	fmt.Println("  Technology Hints                      Detected web technologies")
	fmt.Println("  Response Body                         Response body information")

	fmt.Println()

	// ========================================================
	// HTTP RECON EXAMPLES
	// ========================================================

	fmt.Println(BrightCyan + Bold + "HTTP Recon Examples:" + Reset)

	fmt.Println(
		"  http example.com",
	)

	fmt.Println(
		"  http https://example.com",
	)

	fmt.Println(
		"  http example.com -method HEAD",
	)

	fmt.Println(
		"  http example.com -method OPTIONS",
	)

	fmt.Println(
		"  http example.com -method POST",
	)

	fmt.Println(
		"  http example.com -method PUT",
	)

	fmt.Println(
		"  http example.com -method PATCH",
	)

	fmt.Println(
		"  http example.com -method DELETE",
	)

	fmt.Println(
		"  http example.com -method TRACE",
	)

	fmt.Println(
		"  http example.com -timeout 30s",
	)

	fmt.Println(
		"  http example.com -header \"Authorization: Bearer TOKEN\"",
	)

	fmt.Println(
		"  http example.com -user-agent \"Zebra/1.0\"",
	)

	fmt.Println(
		"  http example.com -no-follow",
	)

	fmt.Println(
		"  http example.com -max-redirects 5",
	)

	fmt.Println(
		"  http example.com -insecure",
	)

	fmt.Println(
		"  http example.com -body-limit 5MB",
	)

	fmt.Println(
		"  http example.com -o report.json",
	)

	fmt.Println()

	// ========================================================
	// GENERAL
	// ========================================================

	fmt.Println(BrightYellow + Bold + "General:" + Reset)

	fmt.Println("  help                                  Show this help")
	fmt.Println("  terminal                              Open a new Zebra terminal")
	fmt.Println("  term                                  Alias for terminal")
	fmt.Println("  newterminal                           Alias for terminal")
	fmt.Println("  exit                                  Exit Zebra")
	fmt.Println("  quit                                  Alias for exit")
	fmt.Println("  --version                             Display Zebra version")
	fmt.Println("  version                               Alias for --version")

	fmt.Println()
}
