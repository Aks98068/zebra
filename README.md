# 🦓 Zebra

**Zebra** is a custom-built, cross-platform cybersecurity CLI toolkit written in **Go**.

It is designed to provide a unified command-line environment for **system operations, reconnaissance, networking, filesystem operations, and security-focused utilities** while giving the developer a deeper understanding of Go, operating systems, networking, and cybersecurity.

> Zebra is being developed from scratch with a focus on learning, modularity, cross-platform compatibility, and eventually becoming a comprehensive security toolkit.

---

## 🚀 Features

Zebra provides an interactive terminal environment with commands organized into multiple categories.

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
* Manage the system `PATH`

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
* Multiple record types

Example:

```text
dns example.com
dns example.com -type A
dns example.com -type MX
dns reverse 8.8.8.8
```

#### WHOIS Reconnaissance

Supports:

* Domain WHOIS lookups
* IP WHOIS lookups
* Custom WHOIS servers
* Request timeouts
* WHOIS referral following
* Output to files

Example:

```text
whois example.com
whois example.com -timeout 20s
whois 8.8.8.8
```

#### HTTP/HTTPS Reconnaissance

Zebra's HTTP reconnaissance command performs an all-in-one HTTP/HTTPS analysis.

It can collect:

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

Example:

```text
http example.com
```

Follow redirects:

```text
http https://example.com -follow
```

Custom User-Agent:

```text
http example.com -user-agent "Zebra/1.0"
```

Custom header:

```text
http example.com -header "X-Test: Zebra"
```

Save results:

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

Target platforms include:

* Windows
* Linux
* macOS

---

## Clone the repository

```bash
git clone https://github.com/Aks98068/zebra.git
cd zebra
```

Install dependencies:

```bash
go mod download
```

Build Zebra:

```bash
go build ./...
```

To build the executable:

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

Example:

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
| Recon           | `dns`, `whois`, `http`                                                        |
| General         | `help`, `terminal`, `exit`, `quit`                                            |

Run:

```text
help
```

inside Zebra for the complete command reference.

---

# 🏗️ Architecture

Zebra follows a modular command architecture.

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

Commands are registered through a centralized registry:

```text
Command
   │
   ├── Name
   ├── Aliases
   ├── Description
   ├── Usage
   └── Run()
```

This allows Zebra to grow by adding independent command modules without turning the entire project into a single large implementation.

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

The exact structure may evolve as Zebra grows.

---

# 📤 Output

Zebra supports saving command results to files where supported.

Example:

```text
dns example.com -o dns.txt
```

```text
whois example.com -o whois.txt
```

```text
http example.com -o report.json
```

The output layer is separated from command logic so commands can produce structured results and then send those results to:

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

---

# 🔐 Security & Responsible Use

Zebra is intended for:

* Authorized security testing
* Security research
* Educational purposes
* System administration
* Network and application reconnaissance
* Controlled laboratory environments

Only use Zebra against systems, applications, networks, and domains that you own or have explicit permission to test.

The developer is not responsible for unauthorized or illegal use of the toolkit.

---

# 🧪 Development Philosophy

Zebra is intentionally being built from the ground up rather than relying heavily on existing security frameworks.

The project focuses on understanding:

* Go programming
* Operating-system interfaces
* Networking
* HTTP
* DNS
* Processes
* Filesystems
* Cross-platform development
* Security tooling architecture
* CLI design
* Modular software architecture

The goal is not simply to create a collection of commands, but to understand how the underlying functionality works.

---

# 🗺️ Roadmap

Zebra is actively evolving.

### Phase 1 — Core CLI

* [x] Interactive terminal
* [x] Command registry
* [x] Command aliases
* [x] Help system
* [x] Cross-platform foundation
* [x] Privilege handling

### Phase 2 — System & Filesystem

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

### Phase 3 — Passive Reconnaissance

* [x] DNS reconnaissance
* [x] WHOIS reconnaissance
* [x] HTTP/HTTPS reconnaissance
* [ ] Email/MX reconnaissance
* [ ] SPF analysis
* [ ] DMARC analysis
* [ ] TLS analysis improvements
* [ ] Additional HTTP fingerprinting

### Phase 4 — Active Reconnaissance

* [ ] ICMP/ping
* [ ] TCP port scanning
* [ ] UDP reconnaissance
* [ ] Service detection
* [ ] Banner grabbing
* [ ] Traceroute
* [ ] Network discovery

### Phase 5 — Security Analysis

* [ ] HTTP security configuration checks
* [ ] TLS configuration analysis
* [ ] Vulnerability intelligence integration
* [ ] CVE lookup
* [ ] Configuration analysis
* [ ] Additional security checks

### Phase 6 — OSINT

* [ ] Username reconnaissance
* [ ] Public-source reconnaissance
* [ ] Metadata analysis
* [ ] Domain intelligence
* [ ] Organization reconnaissance

### Future

The long-term goal is to evolve Zebra into a modular, cross-platform cybersecurity toolkit containing:

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

---

# 🤝 Contributing

Contributions, ideas, bug reports, and improvements are welcome.

Before submitting changes:

```bash
go fmt ./...
go vet ./...
go test ./...
go build ./...
```

Keep new functionality modular and follow the existing command architecture.

---

# 📜 License

License information will be added as the project develops.

---

# 👨‍💻 Author

**Abhishekh Kumar Sah**

Zebra is an independent project built to explore Go programming, systems programming, cybersecurity, networking, and security-tool development.

---

## ⭐ Project Status

**Zebra is actively under development.**

The architecture and command set will continue to evolve as new security, networking, and system capabilities are implemented.
