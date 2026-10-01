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
	fmt.Println("  copy <source> <destination>           Copy file")
	fmt.Println("  move <source> <destination>           Move file/folder")
	fmt.Println("  cat <file>                            Read a file")
	fmt.Println("  write <file>                          Write a file")

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
	fmt.Println("  path                                  View PATH")
	fmt.Println("  path add <directory>                  Add directory to PATH")
	fmt.Println("  path remove <directory>               Remove directory from PATH")

	fmt.Println()

	// ========================================================
	// PROCESS
	// ========================================================

	fmt.Println(
		BrightYellow + Bold +
			"Process:" +
			Reset,
	)

	fmt.Println("  ps                                    List running processes")
	fmt.Println("  ppid <pid>                            Find process by PID")
	fmt.Println("  pgrep <name>                          Search processes by name")
	fmt.Println("  kill <pid>                            Kill process by PID")
	fmt.Println("  pkill <name>                          Kill processes by name")
	fmt.Println("  killself                              Terminate current process")

	fmt.Println()

	// ========================================================
	// LINKS
	// ========================================================

	fmt.Println(
		BrightYellow + Bold +
			"Links:" +
			Reset,
	)

	fmt.Println("  symlink <target> <link>               Create symbolic link")
	fmt.Println("  readlink <link>                       Read symbolic link target")
	fmt.Println("  hardlink <source> <destination>       Create hard link")

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
			"Signals & Exit Codes:" +
			Reset,
	)

	fmt.Println("  signals                               Listen for OS signals")
	fmt.Println("  exitcode <code>                       Exit with specific code")

	fmt.Println()

	// ========================================================
	// SYSTEM INFORMATION
	// ========================================================

	fmt.Println(
		BrightYellow + Bold +
			"System Information:" +
			Reset,
	)

	fmt.Println("  sysinfo                               Show system information")
	fmt.Println("  systeminfo                            Show detailed computer information")
	fmt.Println("  hostname                              Show machine hostname")
	fmt.Println("  homedir                               Show user home directory")
	fmt.Println("  cachedir                              Show user cache directory")
	fmt.Println("  configdir                             Show user config directory")

	fmt.Println()

	// ========================================================
	// TEMPORARY FILES
	// ========================================================

	fmt.Println(
		BrightYellow + Bold +
			"Temporary Files:" +
			Reset,
	)

	fmt.Println("  tmpfile [prefix]                      Create temporary file")
	fmt.Println("  tmpdir [prefix]                       Create temporary directory")
	fmt.Println("  tempdir                               Show system temp directory")

	fmt.Println()

	// ========================================================
	// RECONNAISSANCE
	// ========================================================

	fmt.Println(
		BrightYellow + Bold +
			"Reconnaissance:" +
			Reset,
	)

	fmt.Println(
		"  dns <domain>                          DNS reconnaissance",
	)

	fmt.Println(
		"  dns reverse <ip>                      Reverse DNS lookup",
	)

	fmt.Println(
		"  whois <domain|ip>                     WHOIS reconnaissance",
	)

	fmt.Println(
		"  http <url>                            HTTP/HTTPS reconnaissance",
	)

	fmt.Println(
		"  database <target> [--verify]          Database service reconnaissance",
	)

	fmt.Println(
		"  tls <host> [port]                     TLS/SSL certificate reconnaissance",
	)

	fmt.Println(
		"  subfinder <domain>                    Subdomain reconnaissance",
	)

	fmt.Println(
		"  osint <type> <target>                 OSINT reconnaissance",
	)

	fmt.Println()

	// ========================================================
	// RECON ALIASES
	// ========================================================

	fmt.Println(
		BrightCyan + Bold +
			"Recon Aliases:" +
			Reset,
	)

	fmt.Println("  dnslookup                             Alias for dns")
	fmt.Println("  whoislookup                           Alias for whois")
	fmt.Println("  httprecon                             Alias for http")
	fmt.Println("  http-recon                            Alias for http")
	fmt.Println("  databaserecon                         Alias for database")
	fmt.Println("  dbrecon                               Alias for database")
	fmt.Println("  ssl                                   Alias for tls")
	fmt.Println("  subdomain                             Alias for subfinder")
	fmt.Println("  subs                                  Alias for subfinder")
	fmt.Println("  recon                                 Alias for osint")

	fmt.Println()

	// ========================================================
	// HTTP METHODS
	// ========================================================

	fmt.Println(
		BrightCyan + Bold +
			"HTTP Methods:" +
			Reset,
	)

	fmt.Println("  GET                                   Retrieve a resource")
	fmt.Println("  POST                                  Submit data")
	fmt.Println("  PUT                                   Replace a resource")
	fmt.Println("  PATCH                                 Partially modify a resource")
	fmt.Println("  DELETE                                Delete a resource")
	fmt.Println("  HEAD                                  Retrieve headers only")
	fmt.Println("  OPTIONS                               Query supported methods")
	fmt.Println("  TRACE                                 Diagnostic HTTP request")
	fmt.Println("  CONNECT                               Establish HTTP tunnel")

	fmt.Println()

	// ========================================================
	// HTTP RECON OPTIONS
	// ========================================================

	fmt.Println(
		BrightCyan + Bold +
			"HTTP Recon Options:" +
			Reset,
	)

	fmt.Println("  -method <method>                      Set HTTP method")
	fmt.Println("  -timeout <duration>                   Set request timeout")
	fmt.Println("  -header <key:value>                   Add HTTP header")
	fmt.Println("  -user-agent <value>                   Set User-Agent")
	fmt.Println("  -proxy <url>                          Use HTTP/HTTPS proxy")
	fmt.Println("  -insecure                             Disable TLS verification")
	fmt.Println("  -follow                               Follow redirects")
	fmt.Println("  -no-follow                            Disable redirect following")
	fmt.Println("  -max-redirects <n>                    Maximum redirects")
	fmt.Println("  -body <data>                          Send request body")
	fmt.Println("  -body-file <file>                     Send body from file")
	fmt.Println("  -content-type <type>                  Set Content-Type")
	fmt.Println("  -body-limit <size>                    Limit response body")
	fmt.Println("  -o <file>                             Save reconnaissance report")

	fmt.Println()

	// ========================================================
	// HTTP RECON INFORMATION
	// ========================================================

	fmt.Println(
		BrightCyan + Bold +
			"HTTP Recon Information:" +
			Reset,
	)

	fmt.Println("  Status Code                           HTTP response status")
	fmt.Println("  Response Time                         Request duration")
	fmt.Println("  Protocol                              HTTP protocol version")
	fmt.Println("  Headers                               Response headers")
	fmt.Println("  Cookies                               Response cookies")
	fmt.Println("  Redirects                             Redirect chain")
	fmt.Println("  TLS                                   TLS version and cipher")
	fmt.Println("  Certificates                          Certificate information")
	fmt.Println("  Security Headers                      Security header analysis")
	fmt.Println("  Page Title                            HTML page title")
	fmt.Println("  Technology Hints                      Technology indicators")
	fmt.Println("  Response Body                         Response body information")

	fmt.Println()

	// ========================================================
	// DATABASE RECON
	// ========================================================

	fmt.Println(
		BrightCyan + Bold +
			"Database Recon:" +
			Reset,
	)

	fmt.Println(
		"  database <target>                     Passive database reconnaissance",
	)

	fmt.Println(
		"  database <target> --verify            Verify database endpoints",
	)

	fmt.Println(
		"  databaserecon <target>                Alias for database",
	)

	fmt.Println(
		"  dbrecon <target>                      Alias for database",
	)

	fmt.Println()

	fmt.Println(
		"  Supported services:",
	)

	fmt.Println(
		"    MySQL                               TCP 3306",
	)

	fmt.Println(
		"    PostgreSQL                          TCP 5432",
	)

	fmt.Println(
		"    Redis                               TCP 6379",
	)

	fmt.Println(
		"    MongoDB                             TCP 27017",
	)

	fmt.Println(
		"    Microsoft SQL Server                TCP 1433",
	)

	fmt.Println(
		"    Oracle                              TCP 1521",
	)

	fmt.Println(
		"    Cassandra                           TCP 9042",
	)

	fmt.Println(
		"    CouchDB                             TCP 5984",
	)

	fmt.Println(
		"    ArangoDB                            TCP 8529",
	)

	fmt.Println()

	// ========================================================
	// OSINT
	// ========================================================

	fmt.Println(
		BrightCyan + Bold +
			"OSINT Recon:" +
			Reset,
	)

	fmt.Println("  osint gravatar <email>                Gravatar lookup")
	fmt.Println("  osint hibp <email>                    HIBP lookup")
	fmt.Println("  osint username <name>                 Username reconnaissance")
	fmt.Println("  osint dork <query>                    Search-engine dork reconnaissance")
	fmt.Println("  recon <type> <target>                 Alias for osint")

	fmt.Println()

	// ========================================================
	// TLS / SSL
	// ========================================================

	fmt.Println(
		BrightCyan + Bold +
			"TLS / SSL Recon:" +
			Reset,
	)

	fmt.Println("  tls <host>                            TLS certificate reconnaissance")
	fmt.Println("  tls <host> <port>                     Specify TLS port")
	fmt.Println("  ssl <host>                            Alias for tls")

	fmt.Println()

	// ========================================================
	// SUBDOMAIN RECON
	// ========================================================

	fmt.Println(
		BrightCyan + Bold +
			"Subdomain Recon:" +
			Reset,
	)

	fmt.Println("  subfinder <domain>                    Discover subdomains")
	fmt.Println("  subdomain <domain>                    Alias for subfinder")
	fmt.Println("  subs <domain>                         Alias for subfinder")

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
	fmt.Println("  info                                  Show tool information")
	fmt.Println("  terminal                              Open a new Zebra terminal")
	fmt.Println("  exit                                  Exit Zebra")
	fmt.Println("  --version                             Display Zebra version")

	fmt.Println()

	// ========================================================
	// GENERAL ALIASES
	// ========================================================

	fmt.Println(
		BrightCyan + Bold +
			"General Aliases:" +
			Reset,
	)

	fmt.Println("  term                                  Alias for terminal")
	fmt.Println("  newterminal                           Alias for terminal")
	fmt.Println("  quit                                  Alias for exit")
	fmt.Println("  version                               Alias for --version")

	fmt.Println()

	fmt.Println(
		BrightGreen + Bold +
			"Network Reconnaissance:" +
			Reset,
	)

	fmt.Println(
		"  networkscan <target> [options]",
	)

	fmt.Println(
		"  network <target> [options]",
	)

	fmt.Println(
		"  nmap <target> [options]",
	)

	fmt.Println()

	fmt.Println(
		"  Scan Types:",
	)

	fmt.Println(
		"    -sT                 TCP connect scan",
	)

	fmt.Println(
		"    -sU                 UDP scan",
	)

	fmt.Println(
		"    -sT -sU             TCP and UDP scan",
	)

	fmt.Println()

	fmt.Println(
		"  Detection:",
	)

	fmt.Println(
		"    -sV                 Service and version detection",
	)

	fmt.Println(
		"    -O                  Operating system detection",
	)

	fmt.Println(
		"    -v                  Verbose scan output",
	)

	fmt.Println()

	fmt.Println(
		"  Port Selection:",
	)

	fmt.Println(
		"    -p <ports>          Scan specified ports",
	)

	fmt.Println(
		"    -p 22               Scan port 22",
	)

	fmt.Println(
		"    -p 22,80,443        Scan multiple ports",
	)

	fmt.Println(
		"    -p 1-1000           Scan port range",
	)

	fmt.Println(
		"    -p 22,80,443,8000-9000",
	)

	fmt.Println(
		"                        Multiple ports and ranges",
	)

	fmt.Println(
		"    -p-                 Scan ports 1-65535",
	)

	fmt.Println(
		"    --all-ports         Scan ports 1-65535",
	)

	fmt.Println()

	fmt.Println(
		"  Result Filtering:",
	)

	fmt.Println(
		"    --open              Display only open ports",
	)

	fmt.Println(
		"    --reason            Display reason for port state",
	)

	fmt.Println()

	fmt.Println(
		"  Timing and Performance:",
	)

	fmt.Println(
		"    --timeout <time>    Connection timeout",
	)

	fmt.Println(
		"    -T1                 Very slow timing",
	)

	fmt.Println(
		"    -T2                 Slow timing",
	)

	fmt.Println(
		"    -T3                 Normal timing",
	)

	fmt.Println(
		"    -T4                 Faster timing",
	)

	fmt.Println(
		"    -T5                 Fast timing",
	)

	fmt.Println()

	fmt.Println(
		"  Output:",
	)

	fmt.Println(
		"    -o <file>           Write text report",
	)

	fmt.Println(
		"    -oJ <file>          Write JSON report",
	)

	fmt.Println(
		"    -oX <file>          Write XML report",
	)

	fmt.Println(
		"    -oH <file>          Write HTML report",
	)

	fmt.Println()

	fmt.Println(
		"  Examples:",
	)

	fmt.Println(
		"    nmap 192.168.1.1",
	)

	fmt.Println(
		"    nmap 192.168.1.1 -sT",
	)

	fmt.Println(
		"    nmap 192.168.1.1 -sU",
	)

	fmt.Println(
		"    nmap 192.168.1.1 -sT -sU",
	)

	fmt.Println(
		"    nmap 192.168.1.1 -sT -sV",
	)

	fmt.Println(
		"    nmap 192.168.1.1 -sT -sV -O",
	)

	fmt.Println(
		"    nmap 192.168.1.1 -p 22,80,443",
	)

	fmt.Println(
		"    nmap 192.168.1.1 -p 1-1000",
	)

	fmt.Println(
		"    nmap 192.168.1.1 -p-",
	)

	fmt.Println(
		"    nmap 192.168.1.1 --open",
	)

	fmt.Println(
		"    nmap 192.168.1.1 --reason",
	)

	fmt.Println(
		"    nmap 192.168.1.1 -v",
	)

	fmt.Println(
		"    nmap 192.168.1.1 -T4 --timeout 2s",
	)

	fmt.Println(
		"    nmap 192.168.1.1 -sV -o report.txt",
	)

	fmt.Println(
		"    nmap 192.168.1.1 -sV -oJ report.json",
	)

	fmt.Println(
		"    nmap 192.168.1.1 -sV -oX report.xml",
	)

	fmt.Println(
		"    nmap 192.168.1.1 -sV -oH report.html",
	)

	fmt.Println()

	fmt.Println(
		BrightGreen + Bold +
			"Network Diagnostics:" +
			Reset,
	)

	fmt.Println(
		"  ping <target> [-c <count>] [-W <timeout>]",
	)

	fmt.Println()

	fmt.Println(
		"  ping Options:",
	)

	fmt.Println(
		"    -c <count>          Number of ICMP echo requests",
	)

	fmt.Println(
		"    -W <duration>       Timeout for each request",
	)

	fmt.Println(
		"    --timeout <duration>",
	)

	fmt.Println(
		"                        Timeout for each request",
	)

	fmt.Println()

	fmt.Println(
		"  ping Examples:",
	)

	fmt.Println(
		"    ping 192.168.1.1",
	)

	fmt.Println(
		"    ping google.com",
	)

	fmt.Println(
		"    ping 192.168.1.1 -c 5",
	)

	fmt.Println(
		"    ping 192.168.1.1 -W 1s",
	)

	fmt.Println(
		"    ping 192.168.1.1 -c 10 -W 2s",
	)

	fmt.Println()
}
