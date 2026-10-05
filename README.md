# 🦓 Zebra

**Zebra** is a custom-built, cross-platform **cybersecurity CLI toolkit written in Go**.

It provides a unified command-line environment for **system operations, reconnaissance, networking, filesystem operations, and security-focused utilities** while helping the developer build a deeper understanding of **Go, operating systems, networking, cybersecurity, and security-tool development**.

Zebra is being developed from scratch with a focus on:

* Learning by implementation
* Modular architecture
* Cross-platform compatibility
* System-level programming
* Networking and reconnaissance
* Security-focused tooling
* Maintainable CLI design

> Zebra is an independent project developed by **Abhishekh Kumar Sah (`Aks98068`)** as a practical exploration of Go programming, systems programming, cybersecurity, networking, and security tooling.

---

## 🚀 Features

Zebra provides an interactive terminal environment with commands organized into multiple categories.

### 📁 Filesystem

Zebra provides filesystem and file-management operations including:

* Create files and directories
* Navigate directories
* List files and folders
* Copy files
* Move files and directories
* Remove files and directories
* Read files
* Write files
* Create symbolic links
* Create hard links
* Read symbolic-link targets
* Inspect file permissions and ownership

---

### ⚙️ Environment

Environment management functionality includes:

* Read environment variables
* List environment variables
* Set persistent environment variables
* Remove environment variables
* Manage the system `PATH`

---

### 🔄 Process Management

Zebra provides process-management functionality including:

* List running processes
* Search processes by name
* Find processes by PID
* Terminate processes
* Terminate processes by name
* Inspect parent-child process relationships

---

### 🖥️ System Information

System information commands provide access to information such as:

* Operating-system information
* Hostname
* Home directory
* Cache directory
* Configuration directory
* Temporary directory information

---

# 🌐 Reconnaissance

Zebra includes security-focused reconnaissance capabilities intended for **authorized testing, security research, system administration, and controlled laboratory environments**.

## DNS Reconnaissance

Zebra supports common DNS record and lookup operations.

Supported functionality includes:

* A records
* AAAA records
* MX records
* NS records
* TXT records
* CNAME records
* Reverse DNS lookups
* Multiple record types

### Examples

```text
dns example.com
```

```text
dns example.com -type A
```

```text
dns example.com -type MX
```

```text
dns reverse 8.8.8.8
```

---

## WHOIS Reconnaissance

Zebra supports domain and IP WHOIS lookups.

Features include:

* Domain WHOIS lookups
* IP WHOIS lookups
* Custom WHOIS servers
* Request timeouts
* WHOIS referral following
* Saving results to files

### Examples

```text
whois example.com
```

```text
whois example.com -timeout 20s
```

```text
whois 8.8.8.8
```

---

## HTTP/HTTPS Reconnaissance

Zebra's HTTP reconnaissance command performs an all-in-one analysis of HTTP and HTTPS endpoints.

It can collect information including:

* HTTP status code
* Response status
* Final URL
* Redirect chain
* HTTP protocol
* Response timing
* Response headers
* Cookies
* TLS information
* TLS certificate information
* Security-related headers
* Content type
* Content length
* HTML page title
* Technology hints
* Bounded response body

### Basic HTTP reconnaissance

```text
http example.com
```

### Follow redirects

```text
http https://example.com -follow
```

### Custom User-Agent

```text
http example.com -user-agent "Zebra/1.0"
```

### Custom header

```text
http example.com -header "X-Test: Zebra"
```

### Save results

```text
http example.com -o report.json
```

---

# 🛠️ Installation

## Requirements

* Go 1.XX or newer
* Git
* A supported operating system

Zebra is being developed with cross-platform support in mind.

### Target platforms

* Windows
* Linux
* macOS

> Replace `Go 1.XX` with the actual minimum Go version supported by the current release.

---

## Clone the Repository

```bash
git clone https://github.com/Aks98068/zebra.git
cd zebra
```

Install dependencies:

```bash
go mod download
```

Build the project:

```bash
go build ./...
```

Build the Zebra executable:

```bash
go build -o zebra .
```

On Windows:

```powershell
go build -o zebra.exe .
```

---

# ▶️ Usage

Start Zebra:

```bash
zebra
```

You will enter the interactive Zebra terminal.

Example session:

```text
zebra> help

zebra> pwd

zebra> ls

zebra> dns example.com

zebra> whois example.com

zebra> http https://example.com

zebra> ps

zebra> hostname

zebra> exit
```

Use:

```text
help
```

inside Zebra to display the available commands.

---

# 📖 Command Overview

| Category        | Commands                                                                      |
| --------------- | ----------------------------------------------------------------------------- |
| Filesystem      | `folder`, `file`, `cd`, `pwd`, `ls`, `remove`, `copy`, `move`, `cat`, `write` |
| Environment     | `get`, `set`, `unset`, `path`                                                 |
| Processes       | `ps`, `procs`, `ppid`, `pgrep`, `kill`, `pkill`, `killself`                   |
| Symlinks        | `symlink`, `readlink`, `hardlink`                                             |
| Permissions     | `chmod`, `chown`, `lchown`                                                    |
| Stdio           | `stdin`, `stdout`, `stderr`                                                   |
| Signals         | `signals`, `exitcode`                                                         |
| System          | `sysinfo`, `hostname`, `homedir`, `cachedir`, `configdir`                     |
| Temporary Files | `tmpfile`, `tmpdir`, `tempdir`                                                |
| Reconnaissance  | `dns`, `whois`, `http`                                                        |
| General         | `help`, `terminal`, `exit`, `quit`                                            |

Run:

```text
help
```

inside Zebra for the complete command reference.

---

# 🏗️ Architecture

Zebra follows a modular command architecture designed to allow new functionality to be added without turning the project into one large implementation.

```text
                         ┌──────────────────┐
                         │      Zebra       │
                         │   Interactive    │
                         │     Terminal     │
                         └────────┬─────────┘
                                  │
                                  ▼
                         ┌──────────────────┐
                         │ Command Registry │
                         └────────┬─────────┘
                                  │
              ┌───────────────────┼───────────────────┐
              │                   │                   │
              ▼                   ▼                   ▼
        Filesystem             System              Recon
        Commands              Commands            Commands
              │                   │                   │
              ▼                   ▼                   ▼
        ┌───────────┐       ┌───────────┐       ┌─────────────┐
        │ Files     │       │ Processes │       │ DNS         │
        │ Folders   │       │ System    │       │ WHOIS       │
        │ Links     │       │ Signals   │       │ HTTP/HTTPS  │
        │ Permissions│      │           │       │             │
        └───────────┘       └───────────┘       └─────────────┘
```

Commands are registered through a centralized command registry:

```text
Command
   │
   ├── Name
   ├── Aliases
   ├── Description
   ├── Usage
   └── Run()
```

This architecture allows Zebra to grow through independent command modules while keeping command registration and execution consistent.

---

# 📂 Project Structure

The project is organized around internal packages and command modules.

```text
zebra/
│
├── cmd/
│
├── internal/
│   ├── commands/
│   │   ├── filesystem/
│   │   ├── process/
│   │   ├── recon/
│   │   │   ├── dns/
│   │   │   ├── whois/
│   │   │   └── http/
│   │   │
│   │   ├── system/
│   │   └── ...
│   │
│   ├── output/
│   ├── privilege/
│   ├── ui/
│   └── util/
│
├── main.go
├── go.mod
└── README.md
```

The exact project structure may evolve as Zebra develops.

---

# 📤 Output

Zebra supports saving command results to files where supported.

### DNS

```text
dns example.com -o dns.txt
```

### WHOIS

```text
whois example.com -o whois.txt
```

### HTTP

```text
http example.com -o report.json
```

The output layer is separated from command logic so commands can produce structured results and send them to different output destinations.

```text
                 Command
                    │
                    ▼
               Collector
                    │
                    ▼
               Result Data
                /       \
               /         \
              ▼           ▼
          Terminal       File
                        /    \
                       ▼      ▼
                     TXT     JSON
```

This separation is intended to make future output formats and reporting functionality easier to implement.

---

# 🔐 Security & Responsible Use

Zebra is intended for legitimate and authorized purposes, including:

* Authorized security testing
* Security research
* Educational purposes
* System administration
* Network reconnaissance
* Application reconnaissance
* Controlled laboratory environments

Only use Zebra against systems, applications, networks, domains, and infrastructure that you own or have explicit permission to test.

Do not use the toolkit to disrupt, damage, overload, or gain unauthorized access to systems.

The developer is not responsible for unauthorized or illegal use of the toolkit.

---

# 🧪 Development Philosophy

Zebra is intentionally being built from the ground up rather than relying heavily on existing security frameworks.

The project focuses on understanding:

* Go programming
* Operating-system interfaces
* Systems programming
* Networking
* HTTP
* DNS
* Processes
* Filesystems
* Cross-platform development
* Security tooling architecture
* CLI design
* Modular software architecture
* Structured output and reporting

The goal is not simply to create a collection of commands.

The goal is to understand **how the underlying functionality works and how security-focused software can be designed, implemented, tested, and maintained**.

---

# 🗺️ Roadmap

Zebra is actively evolving.

## Phase 1 — Core CLI

* [x] Interactive terminal
* [x] Command registry
* [x] Command aliases
* [x] Help system
* [x] Cross-platform foundation
* [x] Privilege handling

---

## Phase 2 — System & Filesystem

* [x] Filesystem operations
* [x] Environment variables
* [x] Process management
* [x] Links
* [x] Permissions
* [x] System information
* [x] Temporary files
* [ ] Advanced process inspection
* [ ] Advanced filesystem search
* [ ] File hashing

---

## Phase 3 — Passive Reconnaissance

* [x] DNS reconnaissance
* [x] WHOIS reconnaissance
* [x] HTTP/HTTPS reconnaissance
* [ ] Email/MX reconnaissance
* [ ] SPF analysis
* [ ] DMARC analysis
* [ ] TLS analysis improvements
* [ ] Additional HTTP fingerprinting

---

## Phase 4 — Active Reconnaissance

* [ ] ICMP/ping
* [ ] TCP port scanning
* [ ] UDP reconnaissance
* [ ] Service detection
* [ ] Banner grabbing
* [ ] Traceroute
* [ ] Network discovery

---

## Phase 5 — Security Analysis

* [ ] HTTP security configuration checks
* [ ] TLS configuration analysis
* [ ] Vulnerability intelligence integration
* [ ] CVE lookup
* [ ] Configuration analysis
* [ ] Additional security checks

---

## Phase 6 — OSINT

* [ ] Username reconnaissance
* [ ] Public-source reconnaissance
* [ ] Metadata analysis
* [ ] Domain intelligence
* [ ] Organization reconnaissance

---

# 🔭 Future Direction

The long-term goal is to evolve Zebra into a modular, cross-platform **cybersecurity toolkit written in Go**.

```text
                    ZEBRA
                      │
       ┌──────────────┼──────────────┐
       │              │              │
       ▼              ▼              ▼
    SYSTEM          RECON          NETWORK
       │              │              │
       ▼              ▼              ▼
  Processes        DNS             Ports
  Filesystem       WHOIS           Services
  Users            HTTP            Banners
  Network          OSINT           Traceroute
       │              │              │
       └──────────────┼──────────────┘
                      │
                      ▼
                SECURITY ANALYSIS
                      │
                      ▼
             VULNERABILITY / CONFIG
                    CHECKS
```

Future development may introduce additional capabilities for:

* Network analysis
* Security configuration analysis
* TLS analysis
* OSINT
* Digital forensics
* Evidence-related analysis
* Vulnerability intelligence
* Structured security reporting

These capabilities will be introduced progressively as the project architecture evolves.

---

# 🧭 Design Goals

Zebra is being developed around several long-term engineering goals.

### Modularity

Commands should remain independent and reusable.

### Cross-Platform Support

Core functionality should work consistently across Windows, Linux, and macOS where the underlying operating-system APIs permit it.

### Structured Results

Commands should produce structured result data where practical so the same data can be displayed in the terminal or exported to files.

### Maintainability

The project should remain understandable as the number of commands increases.

### Security

Security-sensitive functionality should include validation, error handling, timeouts, and responsible-use considerations.

### Learning Through Implementation

Zebra is not intended to simply wrap existing tools. Building functionality from the ground up is part of the learning process.

---

# 🧪 Development & Quality Checks

Before submitting changes, run:

```bash
go fmt ./...
```

```bash
go vet ./...
```

```bash
go test ./...
```

```bash
go build ./...
```

When adding new commands:

1. Keep the implementation modular.
2. Follow the existing command architecture.
3. Validate user input.
4. Handle errors explicitly.
5. Avoid unnecessary global state.
6. Keep platform-specific functionality isolated where necessary.
7. Add tests where practical.

---

# 🤝 Contributing

Contributions, ideas, bug reports, documentation improvements, and feature suggestions are welcome.

When contributing:

* Follow the existing project structure.
* Keep new functionality modular.
* Follow the command registry architecture.
* Run formatting and validation tools before submitting changes.
* Document new commands and important behavior.
* Keep security-sensitive functionality focused on authorized use.

---

# 📜 License

License information will be added as the project develops.

Until a license is explicitly added to the repository, the project should not be assumed to be freely licensed for redistribution or commercial use.

---

# 👨‍💻 Author

**Abhishekh Kumar Sah (`Aks98068`)**

Zebra is an independent cybersecurity toolkit developed by **Abhishekh Kumar Sah**, also known online as **Aks98068**.

The project explores:

* **Go programming**
* **Systems programming**
* **Cybersecurity**
* **Networking**
* **Reconnaissance**
* **Digital forensics**
* **Security-tool development**
* **Cross-platform CLI development**

GitHub: **Aks98068**

---

## ⭐ Project Status

**Zebra is actively under development.**

The architecture and command set will continue to evolve as new system, networking, reconnaissance, cybersecurity, and analysis capabilities are implemented.

> Built from scratch with Go.
> Built to learn.
> Built to understand security tooling from the inside out.

---

## 🦓 Zebra

**Abhishekh Kumar Sah · Aks98068 · Go · Cybersecurity · Networking · Security Research**
