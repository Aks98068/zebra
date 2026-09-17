package commands

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// ============================================================
// GET ENVIRONMENT VARIABLES
// ============================================================
//
// Supported:
//
//	get PATH
//	get HOME
//	get USERNAME
//	get -A
//	get --all
//	get -A environmentvariables
//
// ============================================================

// getEnvironmentVariableCommand handles the `get` command.
func getEnvironmentVariableCommand(args []string, ctx *Context) bool {

	// --------------------------------------------------------
	// GET ALL VARIABLES
	// --------------------------------------------------------

	if len(args) == 1 && args[0] == "-A" {
		getAllEnvironmentVariables()
		return false
	}

	if len(args) == 1 && args[0] == "--all" {
		getAllEnvironmentVariables()
		return false
	}

	if len(args) == 2 &&
		args[0] == "-A" &&
		strings.EqualFold(args[1], "environmentvariables") {

		getAllEnvironmentVariables()
		return false
	}

	// --------------------------------------------------------
	// MISSING ARGUMENT
	// --------------------------------------------------------

	if len(args) < 1 {
		fmt.Println("Usage:")
		fmt.Println("  get <environment-variable-name>")
		fmt.Println("  get -A")
		fmt.Println("  get --all")
		fmt.Println("  get -A environmentvariables")
		return false
	}

	// --------------------------------------------------------
	// GET SPECIFIC VARIABLE
	// --------------------------------------------------------

	name := args[0]

	getEnvironmentVariable(name)

	return false
}

// getAllEnvironmentVariables prints all environment variables.
func getAllEnvironmentVariables() {

	variables := os.Environ()

	if len(variables) == 0 {
		fmt.Println("No environment variables found.")
		return
	}

	fmt.Println()
	fmt.Println("ENVIRONMENT VARIABLES")
	fmt.Println("=====================")
	fmt.Println()

	for _, variable := range variables {

		parts := strings.SplitN(variable, "=", 2)

		if len(parts) != 2 {
			continue
		}

		name := parts[0]
		value := parts[1]

		fmt.Println("Name :", name)
		fmt.Println("Value:", value)
		fmt.Println()
	}
}

// getEnvironmentVariable prints one environment variable.
func getEnvironmentVariable(name string) {

	name = strings.TrimSpace(name)

	if name == "" {
		fmt.Println("Environment variable name cannot be empty.")
		return
	}

	value, exists := os.LookupEnv(name)

	if !exists {
		fmt.Println("Variable does not exist:", name)
		return
	}

	fmt.Println()
	fmt.Println("Environment Variable")
	fmt.Println("--------------------")
	fmt.Println("Name :", name)
	fmt.Println("Value:", value)
	fmt.Println()
}

// ============================================================
// SET ENVIRONMENT VARIABLE
// ============================================================
//
// Supported:
//
//	set ZEBRA_MODE production
//	set APP_ENV "development mode"
//
// PATH is intentionally blocked:
//
//	set PATH ...
//
// Use:
//
//	path add <directory>
//
// ============================================================

func setEnvironmentVariablesCommand(args []string, ctx *Context) bool {

	if len(args) < 2 {
		fmt.Println("Usage: set <environment-variable-name> <value>")
		return false
	}

	name := strings.TrimSpace(args[0])

	value := strings.Join(args[1:], " ")

	if name == "" {
		fmt.Println("Environment variable name cannot be empty.")
		return false
	}

	// --------------------------------------------------------
	// PROTECT PATH
	// --------------------------------------------------------

	if strings.EqualFold(name, "PATH") {

		fmt.Println("Error: PATH cannot be modified using 'set'.")
		fmt.Println()
		fmt.Println("Use:")
		fmt.Println(`  path add "C:\Program Files\dotnet"`)
		fmt.Println(`  path remove "C:\Program Files\dotnet"`)

		return false
	}

	// --------------------------------------------------------
	// PERSIST VARIABLE
	// --------------------------------------------------------

	if err := setPersistentEnvironmentVariable(name, value); err != nil {
		fmt.Println("Error:", err)
		return false
	}

	// --------------------------------------------------------
	// UPDATE CURRENT ZEBRA PROCESS
	// --------------------------------------------------------

	if err := os.Setenv(name, value); err != nil {

		fmt.Println(
			"Warning: variable was persisted, but current process could not update it:",
			err,
		)

		return false
	}

	fmt.Println("Environment variable set:", name)

	return false
}

// ============================================================
// UNSET ENVIRONMENT VARIABLE
// ============================================================
//
// Supported:
//
//	unset ZEBRA_MODE
//
// PATH is intentionally protected:
//
//	unset PATH
//
// Use:
//
//	path remove <directory>
//
// ============================================================

func unsetEnvironmentVariableCommand(args []string, ctx *Context) bool {

	if len(args) != 1 {
		fmt.Println("Usage: unset <environment-variable-name>")
		return false
	}

	name := strings.TrimSpace(args[0])

	if name == "" {
		fmt.Println("Environment variable name cannot be empty.")
		return false
	}

	// --------------------------------------------------------
	// PROTECT PATH
	// --------------------------------------------------------

	if strings.EqualFold(name, "PATH") {

		fmt.Println("Error: PATH cannot be removed using 'unset'.")
		fmt.Println()
		fmt.Println("Use:")
		fmt.Println(`  path remove "C:\Program Files\dotnet"`)

		return false
	}

	// --------------------------------------------------------
	// CHECK VARIABLE
	// --------------------------------------------------------

	_, exists := os.LookupEnv(name)

	if !exists {
		fmt.Println("Variable does not exist:", name)
		return false
	}

	// --------------------------------------------------------
	// REMOVE PERSISTENT VARIABLE
	// --------------------------------------------------------

	if err := unsetPersistentEnvironmentVariable(name); err != nil {
		fmt.Println("Error:", err)
		return false
	}

	// --------------------------------------------------------
	// REMOVE FROM CURRENT PROCESS
	// --------------------------------------------------------

	if err := os.Unsetenv(name); err != nil {

		fmt.Println(
			"Warning: persistent variable was removed, but current process could not update it:",
			err,
		)

		return false
	}

	fmt.Println("Environment variable removed:", name)

	return false
}

// ============================================================
// PATH COMMAND
// ============================================================
//
// Supported:
//
//	path
//	path add "C:\Program Files\dotnet"
//	path remove "C:\Program Files\dotnet"
//
// ============================================================

func pathCommand(args []string, ctx *Context) bool {

	// --------------------------------------------------------
	// SHOW PATH
	// --------------------------------------------------------

	if len(args) == 0 {
		showPath()
		return false
	}

	// --------------------------------------------------------
	// ADD PATH ENTRY
	// --------------------------------------------------------

	if strings.EqualFold(args[0], "add") {

		if len(args) < 2 {
			fmt.Println(`Usage: path add <directory>`)
			return false
		}

		entry := strings.Join(args[1:], " ")

		if err := addPathEntry(entry); err != nil {
			fmt.Println("Error:", err)
			return false
		}

		return false
	}

	// --------------------------------------------------------
	// REMOVE PATH ENTRY
	// --------------------------------------------------------

	if strings.EqualFold(args[0], "remove") {

		if len(args) < 2 {
			fmt.Println(`Usage: path remove <directory>`)
			return false
		}

		entry := strings.Join(args[1:], " ")

		if err := removePathEntry(entry); err != nil {
			fmt.Println("Error:", err)
			return false
		}

		return false
	}

	// --------------------------------------------------------
	// INVALID COMMAND
	// --------------------------------------------------------

	fmt.Println("Usage:")
	fmt.Println("  path")
	fmt.Println("  path add <directory>")
	fmt.Println("  path remove <directory>")

	return false
}

// ============================================================
// SHOW PATH
// ============================================================

func showPath() {

	pathValue := os.Getenv("PATH")

	if pathValue == "" {
		fmt.Println("PATH is empty.")
		return
	}

	entries := strings.Split(
		pathValue,
		string(os.PathListSeparator),
	)

	fmt.Println()
	fmt.Println("PATH")
	fmt.Println("====")
	fmt.Println()

	number := 1

	for _, entry := range entries {

		entry = strings.TrimSpace(entry)

		if entry == "" {
			continue
		}

		fmt.Printf("%d. %s\n", number, entry)

		number++
	}

	fmt.Println()
}

// ============================================================
// ADD PATH ENTRY
// ============================================================

func addPathEntry(entry string) error {

	entry = strings.TrimSpace(entry)

	if entry == "" {
		return fmt.Errorf("PATH entry cannot be empty")
	}

	// --------------------------------------------------------
	// GET CURRENT PATH
	// --------------------------------------------------------

	currentPath := os.Getenv("PATH")

	entries := []string{}

	if currentPath != "" {

		entries = strings.Split(
			currentPath,
			string(os.PathListSeparator),
		)
	}

	// --------------------------------------------------------
	// CHECK DUPLICATE
	// --------------------------------------------------------

	for _, existing := range entries {

		existing = strings.TrimSpace(existing)

		if existing == "" {
			continue
		}

		if samePathEntry(existing, entry) {

			fmt.Println("PATH entry already exists:")
			fmt.Println(entry)

			return nil
		}
	}

	// --------------------------------------------------------
	// ADD NEW ENTRY
	// --------------------------------------------------------

	entries = append(entries, entry)

	updatedPath := strings.Join(
		entries,
		string(os.PathListSeparator),
	)

	// --------------------------------------------------------
	// PERSIST PATH
	// --------------------------------------------------------

	if err := setPersistentEnvironmentVariable(
		"PATH",
		updatedPath,
	); err != nil {
		return err
	}

	// --------------------------------------------------------
	// UPDATE CURRENT ZEBRA PROCESS
	// --------------------------------------------------------

	if err := os.Setenv("PATH", updatedPath); err != nil {

		return fmt.Errorf(
			"PATH was persisted but current process could not be updated: %w",
			err,
		)
	}

	fmt.Println("PATH entry added:")
	fmt.Println(entry)

	return nil
}

// ============================================================
// REMOVE PATH ENTRY
// ============================================================

func removePathEntry(entry string) error {

	entry = strings.TrimSpace(entry)

	if entry == "" {
		return fmt.Errorf("PATH entry cannot be empty")
	}

	currentPath := os.Getenv("PATH")

	if currentPath == "" {
		fmt.Println("PATH is empty.")
		return nil
	}

	entries := strings.Split(
		currentPath,
		string(os.PathListSeparator),
	)

	var updatedEntries []string

	found := false

	// --------------------------------------------------------
	// REMOVE MATCHING ENTRY
	// --------------------------------------------------------

	for _, existing := range entries {

		existing = strings.TrimSpace(existing)

		if existing == "" {
			continue
		}

		if samePathEntry(existing, entry) {

			found = true

			continue
		}

		updatedEntries = append(
			updatedEntries,
			existing,
		)
	}

	// --------------------------------------------------------
	// ENTRY NOT FOUND
	// --------------------------------------------------------

	if !found {

		fmt.Println("PATH entry not found:")
		fmt.Println(entry)

		return nil
	}

	// --------------------------------------------------------
	// BUILD NEW PATH
	// --------------------------------------------------------

	updatedPath := strings.Join(
		updatedEntries,
		string(os.PathListSeparator),
	)

	// --------------------------------------------------------
	// PERSIST PATH
	// --------------------------------------------------------

	if err := setPersistentEnvironmentVariable(
		"PATH",
		updatedPath,
	); err != nil {
		return err
	}

	// --------------------------------------------------------
	// UPDATE CURRENT ZEBRA PROCESS
	// --------------------------------------------------------

	if err := os.Setenv("PATH", updatedPath); err != nil {

		return fmt.Errorf(
			"PATH was persisted but current process could not be updated: %w",
			err,
		)
	}

	fmt.Println("PATH entry removed:")
	fmt.Println(entry)

	return nil
}

// ============================================================
// PATH COMPARISON
// ============================================================

func samePathEntry(a string, b string) bool {

	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)

	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}

	return a == b
}

// ============================================================
// PERSISTENT ENVIRONMENT VARIABLE
// ============================================================

func setPersistentEnvironmentVariable(
	name string,
	value string,
) error {

	switch runtime.GOOS {

	case "windows":
		return windowsSetEnvironmentVariable(name, value)

	case "linux":
		return linuxSetEnvironmentVariable(name, value)

	case "darwin":
		return macSetEnvironmentVariable(name, value)

	default:
		return fmt.Errorf(
			"unsupported operating system: %s",
			runtime.GOOS,
		)
	}
}

// ============================================================
// WINDOWS SET
// ============================================================

func windowsSetEnvironmentVariable(
	name string,
	value string,
) error {

	// Use PowerShell only for the persistent operation.
	//
	// The current process is updated separately using os.Setenv.

	command := fmt.Sprintf(
		`[Environment]::SetEnvironmentVariable('%s','%s','Machine')`,
		escapePowerShellSingleQuotes(name),
		escapePowerShellSingleQuotes(value),
	)

	cmd := exec.Command(
		"powershell.exe",
		"-NoProfile",
		"-NonInteractive",
		"-Command",
		command,
	)

	output, err := cmd.CombinedOutput()

	if err != nil {

		return fmt.Errorf(
			"failed to persist Windows environment variable: %v: %s",
			err,
			strings.TrimSpace(string(output)),
		)
	}

	return nil
}

// ============================================================
// WINDOWS ESCAPING
// ============================================================

func escapePowerShellSingleQuotes(value string) string {

	return strings.ReplaceAll(
		value,
		"'",
		"''",
	)
}

// ============================================================
// LINUX SET
// ============================================================

func linuxSetEnvironmentVariable(
	name string,
	value string,
) error {

	command := fmt.Sprintf(
		`export %s="%s"`,
		name,
		escapeShellDoubleQuotes(value),
	)

	cmd := exec.Command(
		"bash",
		"-c",
		fmt.Sprintf(
			`printf '\n%s\n' >> ~/.bashrc`,
			command,
		),
	)

	output, err := cmd.CombinedOutput()

	if err != nil {

		return fmt.Errorf(
			"failed to persist Linux environment variable: %v: %s",
			err,
			strings.TrimSpace(string(output)),
		)
	}

	return nil
}

// ============================================================
// MACOS SET
// ============================================================

func macSetEnvironmentVariable(
	name string,
	value string,
) error {

	command := fmt.Sprintf(
		`export %s="%s"`,
		name,
		escapeShellDoubleQuotes(value),
	)

	cmd := exec.Command(
		"zsh",
		"-c",
		fmt.Sprintf(
			`printf '\n%s\n' >> ~/.zshrc`,
			command,
		),
	)

	output, err := cmd.CombinedOutput()

	if err != nil {

		return fmt.Errorf(
			"failed to persist macOS environment variable: %v: %s",
			err,
			strings.TrimSpace(string(output)),
		)
	}

	return nil
}

// ============================================================
// SHELL ESCAPING
// ============================================================

func escapeShellDoubleQuotes(value string) string {

	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	value = strings.ReplaceAll(value, `$`, `\$`)
	value = strings.ReplaceAll(value, "`", "\\`")

	return value
}

// ============================================================
// UNSET PERSISTENT ENVIRONMENT VARIABLE
// ============================================================

func unsetPersistentEnvironmentVariable(
	name string,
) error {

	switch runtime.GOOS {

	case "windows":
		return windowsUnsetEnvironmentVariable(name)

	case "linux":
		return linuxUnsetEnvironmentVariable(name)

	case "darwin":
		return macUnsetEnvironmentVariable(name)

	default:
		return fmt.Errorf(
			"unsupported operating system: %s",
			runtime.GOOS,
		)
	}
}

// ============================================================
// WINDOWS UNSET
// ============================================================

func windowsUnsetEnvironmentVariable(
	name string,
) error {

	command := fmt.Sprintf(
		`[Environment]::SetEnvironmentVariable('%s',$null,'Machine')`,
		escapePowerShellSingleQuotes(name),
	)

	cmd := exec.Command(
		"powershell.exe",
		"-NoProfile",
		"-NonInteractive",
		"-Command",
		command,
	)

	output, err := cmd.CombinedOutput()

	if err != nil {

		return fmt.Errorf(
			"failed to remove Windows environment variable: %v: %s",
			err,
			strings.TrimSpace(string(output)),
		)
	}

	return nil
}

// ============================================================
// LINUX UNSET
// ============================================================

func linuxUnsetEnvironmentVariable(
	name string,
) error {

	command := fmt.Sprintf(
		`sed -i '/^export %s=/d' ~/.bashrc`,
		escapeSedPattern(name),
	)

	cmd := exec.Command(
		"bash",
		"-c",
		command,
	)

	output, err := cmd.CombinedOutput()

	if err != nil {

		return fmt.Errorf(
			"failed to remove Linux environment variable: %v: %s",
			err,
			strings.TrimSpace(string(output)),
		)
	}

	return nil
}

// ============================================================
// MACOS UNSET
// ============================================================

func macUnsetEnvironmentVariable(
	name string,
) error {

	command := fmt.Sprintf(
		`sed -i '' '/^export %s=/d' ~/.zshrc`,
		escapeSedPattern(name),
	)

	cmd := exec.Command(
		"zsh",
		"-c",
		command,
	)

	output, err := cmd.CombinedOutput()

	if err != nil {

		return fmt.Errorf(
			"failed to remove macOS environment variable: %v: %s",
			err,
			strings.TrimSpace(string(output)),
		)
	}

	return nil
}

// ============================================================
// SED ESCAPING
// ============================================================

func escapeSedPattern(value string) string {

	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `/`, `\/`)
	value = strings.ReplaceAll(value, `.`, `\.`)
	value = strings.ReplaceAll(value, `*`, `\*`)
	value = strings.ReplaceAll(value, `[`, `\[`)
	value = strings.ReplaceAll(value, `]`, `\]`)
	value = strings.ReplaceAll(value, `^`, `\^`)
	value = strings.ReplaceAll(value, `$`, `\$`)

	return value
}