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

	fmt.Println("  folder <path>                         Create a folder")
	fmt.Println("  file <path>                           Create a file")
	fmt.Println("  cd <path>                             Change directory")
	fmt.Println("  pwd                                   Show current directory")
	fmt.Println("  ls [path]                             List files and folders")
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

	fmt.Println(
		BrightYellow + Bold +
			"Environment:" +
			Reset,
	)

	fmt.Println("  get <name>                            Get environment variable")
	fmt.Println("  get -A                                Get all environment variables")
	fmt.Println("  set <name> <value>                    Set persistent environment variable")
	fmt.Println("  unset <name>                          Remove persistent environment variable")

	fmt.Println()

	// ========================================================
	// PROCESS
	// ========================================================

	fmt.Println(
		BrightYellow + Bold +
			"Process:" +
			Reset,
	)

	fmt.Println("  ps                                    List all running processes")
	fmt.Println("  procs                                 Alias for ps")
	fmt.Println("  ppid <pid>                            Find a process by PID")
	fmt.Println("  pgrep <name>                          Search processes by name")
	fmt.Println("  kill <pid>                            Kill a process by PID")
	fmt.Println("  pkill <name>                          Kill all processes matching name")
	fmt.Println("  killself                              Terminate the current process")

	fmt.Println()

	// ========================================================
	// GENERAL
	// ========================================================

	fmt.Println(
		BrightYellow + Bold +
			"General:" +
			Reset,
	)

	fmt.Println("  help                                  Show this help")
	fmt.Println("  terminal                              Open a new Zebra terminal")
	fmt.Println("  exit                                  Exit Zebra")
	fmt.Println("  quit                                  Alias for exit")

	fmt.Println()

	// ========================================================
	// SYMLINKS
	// ========================================================

	fmt.Println(
		BrightYellow + Bold +
			"Symlinks & Links:" +
			Reset,
	)

	fmt.Println("  symlink <target> <link>               Create a symbolic link")
	fmt.Println("  ln-s <target> <link>                  Alias for symlink")
	fmt.Println("  readlink <link>                       Read symlink target")
	fmt.Println("  hardlink <source> <destination>       Create a hard link")
	fmt.Println("  ln <source> <destination>              Alias for hardlink")

	fmt.Println()

	// ========================================================
	// PERMISSIONS
	// ========================================================

	fmt.Println(
		BrightYellow + Bold +
			"Permissions & Ownership:" +
			Reset,
	)

	fmt.Println("  chmod <permissions> <file>            Change file permissions")
	fmt.Println("  chown <uid> <gid> <file>              Change file owner")
	fmt.Println("  lchown <uid> <gid> <link>             Change symlink owner")

	fmt.Println()

	// ========================================================
	// STDIO
	// ========================================================

	fmt.Println(
		BrightYellow + Bold +
			"Stdio:" +
			Reset,
	)

	fmt.Println("  stdout <message>                      Write to standard output")
	fmt.Println("  stderr <message>                      Write to standard error")
	fmt.Println("  stdin                                 Read from standard input")

	fmt.Println()

	// ========================================================
	// SIGNALS
	// ========================================================

	fmt.Println(
		BrightYellow + Bold +
			"Signals & Exit:" +
			Reset,
	)

	fmt.Println("  signals                               Listen for OS signals")
	fmt.Println("  exitcode <code>                       Exit with specific code")

	fmt.Println()

	// ========================================================
	// SYSTEM INFO
	// ========================================================

	fmt.Println(
		BrightYellow + Bold +
			"System Info:" +
			Reset,
	)

	fmt.Println("  systeminfo                            Show complete system information")
	fmt.Println("  sysinfo                               Show minimal system information")
	fmt.Println("  hostname                              Show machine hostname")
	fmt.Println("  homedir                               Show user home directory")
	fmt.Println("  cachedir                              Show user cache directory")
	fmt.Println("  configdir                             Show user config directory")

	fmt.Println()

	// ========================================================
	// TEMP FILES
	// ========================================================

	fmt.Println(
		BrightYellow + Bold +
			"Temp Files:" +
			Reset,
	)

	fmt.Println("  tmpfile [prefix]                      Create a temporary file")
	fmt.Println("  tmpdir [prefix]                       Create a temporary directory")
	fmt.Println("  tempdir                               Show system temp directory")

	fmt.Println()

	// ========================================================
	// DNS
	// ========================================================

	fmt.Println(
		BrightYellow + Bold +
			"DNS:" +
			Reset,
	)

	fmt.Println("  dns <domain>                          Perform DNS lookup")
	fmt.Println("  dns <domain> -type A                  Lookup IPv4 records")
	fmt.Println("  dns <domain> -type AAAA               Lookup IPv6 records")
	fmt.Println("  dns <domain> -type MX                 Lookup mail records")
	fmt.Println("  dns <domain> -type NS                 Lookup nameservers")
	fmt.Println("  dns <domain> -type TXT                Lookup TXT records")
	fmt.Println("  dns <domain> -type CNAME              Lookup CNAME records")
	fmt.Println("  dns <domain> -type ALL                Lookup all supported records")
	fmt.Println("  dns <domain> -o <file>                Save DNS results")
	fmt.Println("  dns reverse <ip>                      Perform reverse DNS lookup")
	fmt.Println("  dns reverse <ip> -o <file>            Save reverse DNS results")

	fmt.Println()

	// ========================================================
	// RECONNAISSANCE
	// ========================================================

	fmt.Println(
		BrightYellow + Bold +
			"Reconnaissance:" +
			Reset,
	)

	fmt.Println("  whois <domain|ip>                     Perform WHOIS lookup")
	fmt.Println("  whoislookup <domain|ip>               Alias for whois")

	fmt.Println("  http <url>                            Perform HTTP/HTTPS reconnaissance")
	fmt.Println("  httprecon <url>                       Alias for http")
	fmt.Println("  http-recon <url>                      Alias for http")
	fmt.Println("  web <url>                             Alias for http")
	fmt.Println("  webrecon <url>                        Alias for http")

	fmt.Println()

	// ========================================================
	// HTTP RECON OPTIONS
	// ========================================================

	fmt.Println(
		BrightCyan + Bold +
			"HTTP Recon Options:" +
			Reset,
	)

	fmt.Println(
		"  http <url> [-method <method>]         HTTP method:",
	)

	fmt.Println(
		"                                        GET, POST, PUT, PATCH,",
	)

	fmt.Println(
		"                                        DELETE, HEAD, OPTIONS,",
	)

	fmt.Println(
		"                                        TRACE, CONNECT",
	)

	fmt.Println(
		"  http <url> [-timeout <duration>]      Set request timeout",
	)

	fmt.Println(
		"  http <url> [-follow]                  Follow redirects",
	)

	fmt.Println(
		"  http <url> [-no-follow]               Disable redirect following",
	)

	fmt.Println(
		"  http <url> [-max-redirects <number>]  Set maximum redirects",
	)

	fmt.Println(
		"  http <url> [-user-agent <string>]     Set custom User-Agent",
	)

	fmt.Println(
		"  http <url> [-header <key:value>]      Add custom HTTP header",
	)

	fmt.Println(
		"  http <url> [-proxy <url>]             Use HTTP/HTTPS proxy",
	)

	fmt.Println(
		"  http <url> [-insecure]                Disable TLS verification",
	)

	fmt.Println(
		"  http <url> [-body <data>]             Send request body",
	)

	fmt.Println(
		"  http <url> [-body-file <file>]        Send body from file",
	)

	fmt.Println(
		"  http <url> [-content-type <type>]     Set request Content-Type",
	)

	fmt.Println(
		"  http <url> [-body-limit <size>]       Limit response body size",
	)

	fmt.Println(
		"  http <url> [-o <file>]                Save reconnaissance results",
	)

	fmt.Println()

	// ========================================================
	// HTTP RECON INFORMATION
	// ========================================================

	fmt.Println(
		BrightCyan + Bold +
			"HTTP Recon Information:" +
			Reset,
	)

	fmt.Println(
		"  URL Information                       Scheme, host, port, path and query",
	)

	fmt.Println(
		"  DNS Information                      IPv4 and IPv6 resolution",
	)

	fmt.Println(
		"  Server Information                   Server and X-Powered-By detection",
	)

	fmt.Println(
		"  HTTP Timing                           DNS, TCP, TLS, TTFB and total time",
	)

	fmt.Println(
		"  HTTP Methods                          Advertised Allow methods",
	)

	fmt.Println(
		"  Redirects                             Complete redirect chain",
	)

	fmt.Println(
		"  TLS                                   TLS version, cipher and certificates",
	)

	fmt.Println(
		"  Security Headers                     CSP, HSTS, XFO, Referrer-Policy, etc.",
	)

	fmt.Println(
		"  Technologies                          Technology detection hints",
	)

	fmt.Println(
		"  Cookies                               Response cookies and attributes",
	)

	fmt.Println(
		"  Page Title                            HTML page title detection",
	)

	fmt.Println(
		"  Response Body                         Response body with size limit",
	)

	fmt.Println()

	// ========================================================
	// HTTP RECON EXAMPLES
	// ========================================================

	fmt.Println(
		BrightCyan + Bold +
			"HTTP Recon Examples:" +
			Reset,
	)

	fmt.Println(
		"  http example.com                     Run HTTP reconnaissance",
	)

	fmt.Println(
		"  http https://example.com             Scan HTTPS target",
	)

	fmt.Println(
		"  http example.com -method HEAD        Send HEAD request",
	)

	fmt.Println(
		"  http example.com -method OPTIONS     Inspect HTTP methods",
	)

	fmt.Println(
		"  http example.com -method POST \\      Send POST request",
	)

	fmt.Println(
		"      -body '{\"test\":\"hello\"}'",
	)

	fmt.Println(
		"  http example.com -method PUT \\       Send PUT request",
	)

	fmt.Println(
		"      -body-file request.json \\",
	)

	fmt.Println(
		"      -content-type application/json",
	)

	fmt.Println(
		"  http example.com -follow              Follow redirects",
	)

	fmt.Println(
		"  http example.com -timeout 30s        Set 30 second timeout",
	)

	fmt.Println(
		"  http example.com -header 'X-Test: 1' Add custom header",
	)

	fmt.Println(
		"  http example.com -proxy http://127.0.0.1:8080",
	)

	fmt.Println(
		"  http example.com -insecure            Disable TLS verification",
	)

	fmt.Println(
		"  http example.com -o report.json      Save results as JSON",
	)

	fmt.Println()

	// ========================================================
	// SYSTEM INFORMATION EXAMPLE
	// ========================================================

	fmt.Println(
		BrightCyan + Bold +
			"System Info Examples:" +
			Reset,
	)

	fmt.Println(
		"  sysinfo                               Show complete system information",
	)

	fmt.Println(
		"  hostname                              Show hostname",
	)

	fmt.Println(
		"  homedir                               Show user home directory",
	)

	fmt.Println(
		"  cachedir                              Show user cache directory",
	)

	fmt.Println(
		"  configdir                             Show user config directory",
	)

	fmt.Println()
}
