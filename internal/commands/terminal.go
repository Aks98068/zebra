package commands

import (
"fmt"
"os"
"os/exec"
"path/filepath"
"runtime"
"strings"
"syscall"
)

const zebraTerminalChildEnv = "ZEBRA_TERMINAL_CHILD"

// ============================================================
// TERMINAL COMMAND
// ============================================================

func handleTerminal(args []string, ctx *Context) bool {
if len(args) != 0 {
fmt.Println("Usage: terminal")
return false
}


if ctx == nil || ctx.CurrentDir == nil {
	fmt.Println("terminal: invalid command context")
	return false
}

if err := OpenZebraTerminalAt(*ctx.CurrentDir); err != nil {
	fmt.Println("terminal:", err)
	return false
}

fmt.Println("New Zebra terminal opened.")

return false


}

// ============================================================
// OPEN TERMINAL
// ============================================================

func OpenZebraTerminal() error {
currentDir, err := os.Getwd()


if err != nil {
	return OpenZebraTerminalAt(".")
}

return OpenZebraTerminalAt(currentDir)


}

// ============================================================
// OPEN TERMINAL AT DIRECTORY
// ============================================================

func OpenZebraTerminalAt(workingDir string) error {
workingDir = strings.TrimSpace(workingDir)


if workingDir == "" {
	workingDir = "."
}

// ---------------------------------------------------------
// Resolve directory
// ---------------------------------------------------------

absoluteDir, err := filepath.Abs(workingDir)

if err != nil {
	return fmt.Errorf(
		"unable to resolve working directory: %w",
		err,
	)
}

workingDir = filepath.Clean(absoluteDir)

// ---------------------------------------------------------
// Verify directory
// ---------------------------------------------------------

info, err := os.Stat(workingDir)

if err != nil {
	return fmt.Errorf(
		"working directory does not exist: %s: %w",
		workingDir,
		err,
	)
}

if !info.IsDir() {
	return fmt.Errorf(
		"working path is not a directory: %s",
		workingDir,
	)
}

// ---------------------------------------------------------
// Find current Zebra executable.
//
// This is important because Zebra may be installed
// somewhere completely different from the current
// directory.
// ---------------------------------------------------------

executable, err := os.Executable()

if err != nil {
	return fmt.Errorf(
		"unable to find Zebra executable: %w",
		err,
	)
}

executable, err = filepath.Abs(executable)

if err != nil {
	return fmt.Errorf(
		"unable to resolve Zebra executable: %w",
		err,
	)
}

executable = filepath.Clean(executable)

// ---------------------------------------------------------
// Platform
// ---------------------------------------------------------

switch runtime.GOOS {

case "windows":
	return openWindowsTerminal(
		executable,
		workingDir,
	)

case "linux":
	return openLinuxTerminal(
		executable,
		workingDir,
	)

case "darwin":
	return openMacTerminal(
		executable,
		workingDir,
	)

default:
	return fmt.Errorf(
		"unsupported operating system: %s",
		runtime.GOOS,
	)
}


}

// ============================================================
// CHILD ENVIRONMENT
// ============================================================

func zebraChildEnvironment() []string {
env := append(
[]string{},
os.Environ()...,
)


// Remove old values first so we never end up with:

// ZEBRA_TERMINAL_CHILD=0
// ZEBRA_TERMINAL_CHILD=1

filtered := make([]string, 0, len(env)+1)

for _, value := range env {
	if strings.HasPrefix(
		value,
		zebraTerminalChildEnv+"=",
	) {
		continue
	}

	filtered = append(filtered, value)
}

filtered = append(
	filtered,
	zebraTerminalChildEnv+"=1",
)

return filtered


}

// ============================================================
// WINDOWS
// ============================================================

func openWindowsTerminal(
executable string,
workingDir string,
) error {


env := zebraChildEnvironment()

// ---------------------------------------------------------
// CREATE_NEW_CONSOLE
//
// Windows creates a completely separate console window.
// ---------------------------------------------------------

const createNewConsole uint32 = 0x00000010

cmd := exec.Command(executable)

cmd.Dir = workingDir
cmd.Env = env

cmd.SysProcAttr = &syscall.SysProcAttr{
	CreationFlags: createNewConsole,
}

// ---------------------------------------------------------
// Start asynchronously.
//
// We deliberately DO NOT call cmd.Run().
//
// Run() would make Zebra wait for the child.
// Start() allows the current Zebra shell to continue.
// ---------------------------------------------------------

if err := cmd.Start(); err != nil {
	return fmt.Errorf(
		"unable to start new Zebra console: %w",
		err,
	)
}

// ---------------------------------------------------------
// Release the process handle.
//
// The new console owns the child process.
// ---------------------------------------------------------

_ = cmd.Process.Release()

return nil


}

// ============================================================
// LINUX
// ============================================================

func openLinuxTerminal(
executable string,
workingDir string,
) error {


env := zebraChildEnvironment()

// ---------------------------------------------------------
// x-terminal-emulator
// ---------------------------------------------------------

if terminal, err := exec.LookPath(
	"x-terminal-emulator",
); err == nil {

	cmd := exec.Command(
		terminal,
		"--working-directory",
		workingDir,
		"--",
		executable,
	)

	cmd.Dir = workingDir
	cmd.Env = env

	if err := cmd.Start(); err == nil {
		_ = cmd.Process.Release()
		return nil
	}
}

// ---------------------------------------------------------
// GNOME Terminal
// ---------------------------------------------------------

if terminal, err := exec.LookPath(
	"gnome-terminal",
); err == nil {

	cmd := exec.Command(
		terminal,
		"--working-directory",
		workingDir,
		"--",
		executable,
	)

	cmd.Env = env

	if err := cmd.Start(); err == nil {
		_ = cmd.Process.Release()
		return nil
	}
}

// ---------------------------------------------------------
// KDE Konsole
// ---------------------------------------------------------

if terminal, err := exec.LookPath(
	"konsole",
); err == nil {

	cmd := exec.Command(
		terminal,
		"--workdir",
		workingDir,
		"-e",
		executable,
	)

	cmd.Env = env

	if err := cmd.Start(); err == nil {
		_ = cmd.Process.Release()
		return nil
	}
}

// ---------------------------------------------------------
// XFCE Terminal
// ---------------------------------------------------------

if terminal, err := exec.LookPath(
	"xfce4-terminal",
); err == nil {

	cmd := exec.Command(
		terminal,
		"--working-directory",
		workingDir,
		"--command",
		executable,
	)

	cmd.Env = env

	if err := cmd.Start(); err == nil {
		_ = cmd.Process.Release()
		return nil
	}
}

// ---------------------------------------------------------
// xterm
// ---------------------------------------------------------

if terminal, err := exec.LookPath("xterm"); err == nil {

	cmd := exec.Command(
		terminal,
		"-e",
		executable,
	)

	cmd.Dir = workingDir
	cmd.Env = env

	if err := cmd.Start(); err == nil {
		_ = cmd.Process.Release()
		return nil
	}
}

return fmt.Errorf(
	"no supported Linux terminal emulator was found",
)


}

// ============================================================
// MACOS
// ============================================================

func openMacTerminal(
executable string,
workingDir string,
) error {


env := zebraChildEnvironment()

script := fmt.Sprintf(
	`tell application "Terminal"
		activate
		do script "cd %s && %s"
	end tell`,
	escapeAppleScript(
		shellQuote(workingDir),
	),
	escapeAppleScript(
		shellQuote(executable),
	),
)

cmd := exec.Command(
	"osascript",
	"-e",
	script,
)

cmd.Env = env

if err := cmd.Start(); err != nil {
	return fmt.Errorf(
		"unable to start macOS Terminal: %w",
		err,
	)
}

_ = cmd.Process.Release()

return nil


}

// ============================================================
// SHELL QUOTING
// ============================================================

func shellQuote(value string) string {
value = strings.ReplaceAll(
value,
`'`,
`'\''`,
)


return "'" + value + "'"


}

// ============================================================
// APPLESCRIPT ESCAPING
// ============================================================

func escapeAppleScript(value string) string {
value = strings.ReplaceAll(
value,
`\`,
`\\`,
)


value = strings.ReplaceAll(
	value,
	`"`,
	`\"`,
)

return value


}
