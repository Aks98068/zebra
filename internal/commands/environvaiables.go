package commands

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
)

// ============================================================
// ENVIRONMENT VARIABLE
// ============================================================

type environmentVariable struct {
	Name  string
	Value string
}

// ============================================================
// TERMINAL COLORS
// ============================================================

const (
	colorReset   = "\033[0m"
	colorBold    = "\033[1m"
	colorDim     = "\033[2m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorWhite   = "\033[37m"
)

// ============================================================
// GET ENVIRONMENT VARIABLE
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

func zebraVersion(args []string, ctx *Context) bool {

	fmt.Println("Zebra 1.0 version")

	return false
}

func getEnvironmentVariableCommand(
	args []string,
	ctx *Context,
) bool {

	// --------------------------------------------------------
	// GET ALL
	// --------------------------------------------------------

	if len(args) == 1 &&
		(args[0] == "-A" ||
			args[0] == "--all") {

		getAllEnvironmentVariables()

		return false
	}

	// --------------------------------------------------------
	// GET ALL WITH LONG FORM
	// --------------------------------------------------------

	if len(args) == 2 &&
		args[0] == "-A" &&
		strings.EqualFold(
			args[1],
			"environmentvariables",
		) {

		getAllEnvironmentVariables()

		return false
	}

	// --------------------------------------------------------
	// MISSING ARGUMENT
	// --------------------------------------------------------

	if len(args) < 1 {

		fmt.Println()

		fmt.Println(
			colorYellow +
				"Usage:" +
				colorReset,
		)

		fmt.Println(
			"  get <environment-variable-name>",
		)

		fmt.Println(
			"  get -A",
		)

		fmt.Println(
			"  get --all",
		)

		fmt.Println(
			"  get -A environmentvariables",
		)

		fmt.Println()

		return false
	}

	// --------------------------------------------------------
	// GET SPECIFIC VARIABLE
	// --------------------------------------------------------

	name := strings.TrimSpace(args[0])

	getEnvironmentVariable(name)

	return false
}

// ============================================================
// GET ALL ENVIRONMENT VARIABLES
// ============================================================

func getAllEnvironmentVariables() {

	variables := os.Environ()

	if len(variables) == 0 {

		fmt.Println(
			colorYellow +
				"No environment variables found." +
				colorReset,
		)

		return
	}

	// --------------------------------------------------------
	// Parse variables
	// --------------------------------------------------------

	var envs []environmentVariable

	for _, variable := range variables {

		parts := strings.SplitN(
			variable,
			"=",
			2,
		)

		if len(parts) != 2 {
			continue
		}

		envs = append(
			envs,
			environmentVariable{
				Name:  parts[0],
				Value: parts[1],
			},
		)
	}

	// --------------------------------------------------------
	// Sort alphabetically
	// --------------------------------------------------------

	sort.Slice(
		envs,
		func(i, j int) bool {

			return strings.ToUpper(
				envs[i].Name,
			) < strings.ToUpper(
				envs[j].Name,
			)
		},
	)

	// --------------------------------------------------------
	// Calculate name width
	// --------------------------------------------------------

	nameWidth := len("NAME")

	for _, env := range envs {

		if len(env.Name) > nameWidth {
			nameWidth = len(env.Name)
		}
	}

	// Prevent an extremely wide NAME column.

	if nameWidth > 32 {
		nameWidth = 32
	}

	// --------------------------------------------------------
	// Header
	// --------------------------------------------------------

	fmt.Println()

	fmt.Println(
		colorCyan +
			"╭──────────────────────────────────────────────────────────────────────────────╮" +
			colorReset,
	)

	fmt.Printf(
		colorCyan+"│"+colorReset+
			" %-76s "+
			colorCyan+"│"+colorReset+
			"\n",
		"ENVIRONMENT VARIABLES",
	)

	fmt.Printf(
		colorCyan+"│"+colorReset+
			" %-76s "+
			colorCyan+"│"+colorReset+
			"\n",
		fmt.Sprintf(
			"%d variables",
			len(envs),
		),
	)

	fmt.Println(
		colorCyan +
			"╰──────────────────────────────────────────────────────────────────────────────╯" +
			colorReset,
	)

	fmt.Println()

	// --------------------------------------------------------
	// Table header
	// --------------------------------------------------------

	fmt.Printf(
		colorBold+
			colorCyan+
			"%-*s  %-s"+
			colorReset+
			"\n",
		nameWidth,
		"NAME",
		"VALUE",
	)

	fmt.Println(
		colorDim +
			strings.Repeat(
				"─",
				nameWidth+70,
			) +
			colorReset,
	)

	// --------------------------------------------------------
	// Display variables
	// --------------------------------------------------------

	for _, env := range envs {

		name := env.Name
		value := env.Value

		nameColor := colorCyan
		valueColor := colorWhite

		// ----------------------------------------------------
		// Special variables
		// ----------------------------------------------------

		switch strings.ToUpper(name) {

		case "PATH":

			nameColor = colorYellow
			valueColor = colorYellow

		case "HOME",
			"HOMEDRIVE",
			"HOMEPATH",
			"USERPROFILE":

			nameColor = colorGreen

		case "TEMP",
			"TMP":

			nameColor = colorMagenta

		case "OS",
			"COMPUTERNAME",
			"USERNAME",
			"USERDOMAIN":

			nameColor = colorBlue
		}

		// ----------------------------------------------------
		// Empty value
		// ----------------------------------------------------

		displayValue := value

		if displayValue == "" {

			displayValue =
				colorDim +
					"<empty>" +
					colorReset
		}

		// ----------------------------------------------------
		// Print row
		// ----------------------------------------------------

		fmt.Printf(
			"%s%-*s%s  %s%s%s\n",

			nameColor,

			nameWidth,

			name,

			colorReset,

			valueColor,

			displayValue,

			colorReset,
		)
	}

	// --------------------------------------------------------
	// Footer
	// --------------------------------------------------------

	fmt.Println()

	fmt.Println(
		colorDim +
			strings.Repeat(
				"─",
				nameWidth+70,
			) +
			colorReset,
	)

	fmt.Printf(
		colorGreen+
			"✓"+
			colorReset+
			" %d environment variables\n",
		len(envs),
	)

	fmt.Println()
}

// ============================================================
// GET ONE ENVIRONMENT VARIABLE
// ============================================================

func getEnvironmentVariable(name string) {

	name = strings.TrimSpace(name)

	if name == "" {

		fmt.Println(
			colorRed +
				"Environment variable name cannot be empty." +
				colorReset,
		)

		return
	}

	value, exists := os.LookupEnv(name)

	if !exists {

		fmt.Println()

		fmt.Println(
			colorRed +
				"✗ Variable does not exist: " +
				colorReset +
				name,
		)

		fmt.Println()

		return
	}

	fmt.Println()

	fmt.Println(
		colorCyan +
			"╭──────────────────────────────────────────────╮" +
			colorReset,
	)

	fmt.Printf(
		colorCyan+"│"+colorReset+
			" %-44s "+
			colorCyan+"│"+colorReset+
			"\n",
		"ENVIRONMENT VARIABLE",
	)

	fmt.Println(
		colorCyan +
			"╰──────────────────────────────────────────────╯" +
			colorReset,
	)

	fmt.Println()

	fmt.Printf(
		colorBold+"Name   :"+colorReset+
			" %s\n",
		name,
	)

	if value == "" {

		fmt.Printf(
			colorBold+"Value  :"+colorReset+
				" %s\n",
			colorDim+"<empty>"+colorReset,
		)

	} else {

		fmt.Printf(
			colorBold+"Value  :"+colorReset+
				" %s\n",
			value,
		)
	}

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
// PATH is intentionally protected.
//
// Use:
//
//	path add <directory>
//	path remove <directory>
//
// ============================================================

func setEnvironmentVariablesCommand(
	args []string,
	ctx *Context,
) bool {

	if len(args) < 2 {

		fmt.Println()

		fmt.Println(
			colorYellow +
				"Usage: set <environment-variable-name> <value>" +
				colorReset,
		)

		fmt.Println()

		return false
	}

	name := strings.TrimSpace(args[0])

	value := strings.Join(
		args[1:],
		" ",
	)

	if name == "" {

		fmt.Println(
			colorRed +
				"Environment variable name cannot be empty." +
				colorReset,
		)

		return false
	}

	// --------------------------------------------------------
	// Protect PATH
	// --------------------------------------------------------

	if strings.EqualFold(
		name,
		"PATH",
	) {

		fmt.Println()

		fmt.Println(
			colorRed +
				"Error: PATH cannot be modified using 'set'." +
				colorReset,
		)

		fmt.Println()

		fmt.Println(
			"Use:",
		)

		fmt.Println(
			`  path add "C:\Program Files\dotnet"`,
		)

		fmt.Println(
			`  path remove "C:\Program Files\dotnet"`,
		)

		fmt.Println()

		return false
	}

	// --------------------------------------------------------
	// Persist variable
	// --------------------------------------------------------

	if err :=
		setPersistentEnvironmentVariable(
			name,
			value,
		); err != nil {

		fmt.Println(
			colorRed +
				"Error: " +
				colorReset +
				err.Error(),
		)

		return false
	}

	// --------------------------------------------------------
	// Update current Zebra process
	// --------------------------------------------------------

	if err := os.Setenv(
		name,
		value,
	); err != nil {

		fmt.Println(
			colorYellow+
				"Warning: variable was persisted, but current process could not update it: "+
				colorReset,
			err,
		)

		return false
	}

	fmt.Println()

	fmt.Printf(
		colorGreen+
			"✓ Environment variable set: "+
			colorReset+
			"%s\n",
		name,
	)

	fmt.Println()

	return false
}

// ============================================================
// UNSET ENVIRONMENT VARIABLE
// ============================================================

func unsetEnvironmentVariableCommand(
	args []string,
	ctx *Context,
) bool {

	if len(args) != 1 {

		fmt.Println()

		fmt.Println(
			colorYellow +
				"Usage: unset <environment-variable-name>" +
				colorReset,
		)

		fmt.Println()

		return false
	}

	name := strings.TrimSpace(args[0])

	if name == "" {

		fmt.Println(
			colorRed +
				"Environment variable name cannot be empty." +
				colorReset,
		)

		return false
	}

	// --------------------------------------------------------
	// Protect PATH
	// --------------------------------------------------------

	if strings.EqualFold(
		name,
		"PATH",
	) {

		fmt.Println()

		fmt.Println(
			colorRed +
				"Error: PATH cannot be removed using 'unset'." +
				colorReset,
		)

		fmt.Println()

		fmt.Println(
			"Use:",
		)

		fmt.Println(
			`  path remove "C:\Program Files\dotnet"`,
		)

		fmt.Println()

		return false
	}

	// --------------------------------------------------------
	// Check variable
	// --------------------------------------------------------

	_, exists := os.LookupEnv(name)

	if !exists {

		fmt.Println()

		fmt.Println(
			colorYellow +
				"Variable does not exist: " +
				colorReset +
				name,
		)

		fmt.Println()

		return false
	}

	// --------------------------------------------------------
	// Remove persistent variable
	// --------------------------------------------------------

	if err :=
		unsetPersistentEnvironmentVariable(
			name,
		); err != nil {

		fmt.Println(
			colorRed +
				"Error: " +
				colorReset +
				err.Error(),
		)

		return false
	}

	// --------------------------------------------------------
	// Remove from current process
	// --------------------------------------------------------

	if err := os.Unsetenv(name); err != nil {

		fmt.Println(
			colorYellow+
				"Warning: persistent variable was removed, but current process could not update it: "+
				colorReset,
			err,
		)

		return false
	}

	fmt.Println()

	fmt.Printf(
		colorGreen+
			"✓ Environment variable removed: "+
			colorReset+
			"%s\n",
		name,
	)

	fmt.Println()

	return false
}

// ============================================================
// PATH COMMAND
// ============================================================
//
// Supported:
//
//	path
//	path add <directory>
//	path remove <directory>
//
// ============================================================

func pathCommand(
	args []string,
	ctx *Context,
) bool {

	// --------------------------------------------------------
	// SHOW PATH
	// --------------------------------------------------------

	if len(args) == 0 {

		showPath()

		return false
	}

	// --------------------------------------------------------
	// ADD PATH
	// --------------------------------------------------------

	if strings.EqualFold(
		args[0],
		"add",
	) {

		if len(args) < 2 {

			fmt.Println(
				colorYellow +
					`Usage: path add <directory>` +
					colorReset,
			)

			return false
		}

		entry := strings.Join(
			args[1:],
			" ",
		)

		if err := addPathEntry(
			entry,
		); err != nil {

			fmt.Println(
				colorRed +
					"Error: " +
					colorReset +
					err.Error(),
			)

			return false
		}

		return false
	}

	// --------------------------------------------------------
	// REMOVE PATH
	// --------------------------------------------------------

	if strings.EqualFold(
		args[0],
		"remove",
	) {

		if len(args) < 2 {

			fmt.Println(
				colorYellow +
					`Usage: path remove <directory>` +
					colorReset,
			)

			return false
		}

		entry := strings.Join(
			args[1:],
			" ",
		)

		if err := removePathEntry(
			entry,
		); err != nil {

			fmt.Println(
				colorRed +
					"Error: " +
					colorReset +
					err.Error(),
			)

			return false
		}

		return false
	}

	// --------------------------------------------------------
	// INVALID COMMAND
	// --------------------------------------------------------

	fmt.Println()

	fmt.Println(
		colorYellow +
			"Usage:" +
			colorReset,
	)

	fmt.Println(
		"  path",
	)

	fmt.Println(
		"  path add <directory>",
	)

	fmt.Println(
		"  path remove <directory>",
	)

	fmt.Println()

	return false
}

// ============================================================
// SHOW PATH
// ============================================================

func showPath() {

	pathValue := os.Getenv("PATH")

	if pathValue == "" {

		fmt.Println(
			colorYellow +
				"PATH is empty." +
				colorReset,
		)

		return
	}

	entries := strings.Split(
		pathValue,
		string(os.PathListSeparator),
	)

	fmt.Println()

	fmt.Println(
		colorYellow +
			"╭──────────────────────────────────────────────────────────────────────────────╮" +
			colorReset,
	)

	fmt.Printf(
		colorYellow+"│"+colorReset+
			" %-76s "+
			colorYellow+"│"+colorReset+
			"\n",
		"PATH",
	)

	fmt.Printf(
		colorYellow+"│"+colorReset+
			" %-76s "+
			colorYellow+"│"+colorReset+
			"\n",
		fmt.Sprintf(
			"%d entries",
			len(entries),
		),
	)

	fmt.Println(
		colorYellow +
			"╰──────────────────────────────────────────────────────────────────────────────╯" +
			colorReset,
	)

	fmt.Println()

	number := 1

	for _, entry := range entries {

		entry = strings.TrimSpace(entry)

		if entry == "" {
			continue
		}

		fmt.Printf(
			colorCyan+
				"%3d"+
				colorReset+
				"  %s\n",
			number,
			entry,
		)

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
		return fmt.Errorf(
			"PATH entry cannot be empty",
		)
	}

	// --------------------------------------------------------
	// Get current PATH
	// --------------------------------------------------------

	currentPath := os.Getenv("PATH")

	var entries []string

	if currentPath != "" {

		entries = strings.Split(
			currentPath,
			string(os.PathListSeparator),
		)
	}

	// --------------------------------------------------------
	// Check duplicate
	// --------------------------------------------------------

	for _, existing := range entries {

		existing = strings.TrimSpace(existing)

		if existing == "" {
			continue
		}

		if samePathEntry(
			existing,
			entry,
		) {

			fmt.Println()

			fmt.Println(
				colorYellow +
					"PATH entry already exists:" +
					colorReset,
			)

			fmt.Println(
				"  " + entry,
			)

			fmt.Println()

			return nil
		}
	}

	// --------------------------------------------------------
	// Add new entry
	// --------------------------------------------------------

	entries = append(
		entries,
		entry,
	)

	updatedPath := strings.Join(
		entries,
		string(os.PathListSeparator),
	)

	// --------------------------------------------------------
	// Persist
	// --------------------------------------------------------

	if err :=
		setPersistentEnvironmentVariable(
			"PATH",
			updatedPath,
		); err != nil {

		return err
	}

	// --------------------------------------------------------
	// Update current process
	// --------------------------------------------------------

	if err := os.Setenv(
		"PATH",
		updatedPath,
	); err != nil {

		return fmt.Errorf(
			"PATH was persisted but current process could not be updated: %w",
			err,
		)
	}

	fmt.Println()

	fmt.Println(
		colorGreen +
			"✓ PATH entry added:" +
			colorReset,
	)

	fmt.Println(
		"  " + entry,
	)

	fmt.Println()

	return nil
}

// ============================================================
// REMOVE PATH ENTRY
// ============================================================

func removePathEntry(entry string) error {

	entry = strings.TrimSpace(entry)

	if entry == "" {

		return fmt.Errorf(
			"PATH entry cannot be empty",
		)
	}

	currentPath := os.Getenv("PATH")

	if currentPath == "" {

		fmt.Println(
			colorYellow +
				"PATH is empty." +
				colorReset,
		)

		return nil
	}

	entries := strings.Split(
		currentPath,
		string(os.PathListSeparator),
	)

	var updatedEntries []string

	found := false

	// --------------------------------------------------------
	// Remove matching entry
	// --------------------------------------------------------

	for _, existing := range entries {

		existing = strings.TrimSpace(existing)

		if existing == "" {
			continue
		}

		if samePathEntry(
			existing,
			entry,
		) {

			found = true

			continue
		}

		updatedEntries = append(
			updatedEntries,
			existing,
		)
	}

	// --------------------------------------------------------
	// Entry not found
	// --------------------------------------------------------

	if !found {

		fmt.Println()

		fmt.Println(
			colorYellow +
				"PATH entry not found:" +
				colorReset,
		)

		fmt.Println(
			"  " + entry,
		)

		fmt.Println()

		return nil
	}

	// --------------------------------------------------------
	// Build new PATH
	// --------------------------------------------------------

	updatedPath := strings.Join(
		updatedEntries,
		string(os.PathListSeparator),
	)

	// --------------------------------------------------------
	// Persist
	// --------------------------------------------------------

	if err :=
		setPersistentEnvironmentVariable(
			"PATH",
			updatedPath,
		); err != nil {

		return err
	}

	// --------------------------------------------------------
	// Update current process
	// --------------------------------------------------------

	if err := os.Setenv(
		"PATH",
		updatedPath,
	); err != nil {

		return fmt.Errorf(
			"PATH was persisted but current process could not be updated: %w",
			err,
		)
	}

	fmt.Println()

	fmt.Println(
		colorGreen +
			"✓ PATH entry removed:" +
			colorReset,
	)

	fmt.Println(
		"  " + entry,
	)

	fmt.Println()

	return nil
}

// ============================================================
// PATH COMPARISON
// ============================================================

func samePathEntry(
	a string,
	b string,
) bool {

	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)

	if runtime.GOOS == "windows" {

		return strings.EqualFold(
			a,
			b,
		)
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

		return windowsSetEnvironmentVariable(
			name,
			value,
		)

	case "linux":

		return linuxSetEnvironmentVariable(
			name,
			value,
		)

	case "darwin":

		return macSetEnvironmentVariable(
			name,
			value,
		)

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
			strings.TrimSpace(
				string(output),
			),
		)
	}

	return nil
}

// ============================================================
// WINDOWS ESCAPING
// ============================================================

func escapePowerShellSingleQuotes(
	value string,
) string {

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
			strings.TrimSpace(
				string(output),
			),
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
			strings.TrimSpace(
				string(output),
			),
		)
	}

	return nil
}

// ============================================================
// SHELL ESCAPING
// ============================================================

func escapeShellDoubleQuotes(
	value string,
) string {

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

	value = strings.ReplaceAll(
		value,
		`$`,
		`\$`,
	)

	value = strings.ReplaceAll(
		value,
		"`",
		"\\`",
	)

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

		return windowsUnsetEnvironmentVariable(
			name,
		)

	case "linux":

		return linuxUnsetEnvironmentVariable(
			name,
		)

	case "darwin":

		return macUnsetEnvironmentVariable(
			name,
		)

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
			strings.TrimSpace(
				string(output),
			),
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
			strings.TrimSpace(
				string(output),
			),
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
			strings.TrimSpace(
				string(output),
			),
		)
	}

	return nil
}

// ============================================================
// SED ESCAPING
// ============================================================

func escapeSedPattern(
	value string,
) string {

	value = strings.ReplaceAll(
		value,
		`\`,
		`\\`,
	)

	value = strings.ReplaceAll(
		value,
		`/`,
		`\/`,
	)

	value = strings.ReplaceAll(
		value,
		`.`,
		`\.`)

	value = strings.ReplaceAll(
		value,
		`*`,
		`\*`,
	)

	value = strings.ReplaceAll(
		value,
		`[`,
		`\[`,
	)

	value = strings.ReplaceAll(
		value,
		`]`,
		`\]`,
	)

	value = strings.ReplaceAll(
		value,
		`^`,
		`\^`,
	)

	value = strings.ReplaceAll(
		value,
		`$`,
		`\$`,
	)

	return value
}
