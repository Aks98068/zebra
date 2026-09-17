package commands

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func getAllEnvironmentVariables() {
	variables := os.Environ()

	for _, variable := range variables {
		parts := strings.SplitN(variable, "=", 2)

		if len(parts) == 2 {
			name := parts[0]
			value := parts[1]

			fmt.Println("Name:", name)
			fmt.Println("Value:", value)
			fmt.Println()
		}
	}
}

func getEnvironmentVariable(name string) {
	value, exists := os.LookupEnv(name)

	if exists {
		fmt.Println("Variable exists:", value)
	} else {
		fmt.Println("Variable does not exist:", name)
	}
}

// Command handler for: get -A environmentvariables
func getAllEnvironmentVariablesCommand(args []string, ctx *Context) bool {
	getAllEnvironmentVariables()
	return true
}

// Command handler for: get <environment-variable-name>
func getEnvironmentVariableCommand(args []string, ctx *Context) bool {
	if len(args) < 1 {
		fmt.Println("Usage: get <environment-variable-name>")
		return false
	}

	name := args[0]

	getEnvironmentVariable(name)

	return true
}

func setPersistentEnvironmentVariable(name string, value string) error {
	switch runtime.GOOS {
	case "windows":
		return winowsfun(name, value)

	case "linux":
		return linuxfun(name, value)

	case "darwin":
		return macfun(name, value)

	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func winowsfun(name string, value string) error {
	command := fmt.Sprintf(
		`[Environment]::SetEnvironmentVariable("%s","%s", "User")`,
		name,
		value,
	)

	cmd := exec.Command(
		"powershell",
		"-NoProfile",
		"-Command",
		command,
	)

	return cmd.Run()
}

func linuxfun(name string, value string) error {
	command := fmt.Sprintf(
		`echo 'export %s="%s"' >> ~/.bashrc`,
		name,
		value,
	)

	cmd := exec.Command("bash", "-c", command)

	return cmd.Run()
}

func macfun(name string, value string) error {
	command := fmt.Sprintf(
		`echo 'export %s="%s"' >> ~/.zshrc`,
		name,
		value,
	)

	cmd := exec.Command("zsh", "-c", command)

	return cmd.Run()
}


func unsetPersistentEnvironmentVariable(name string) error {
	switch runtime.GOOS {
	case "windows":
		return unsetWindowsEnvironmentVariable(name)

	case "linux":
		return unsetLinuxEnvironmentVariable(name)

	case "darwin":
		return unsetMacEnvironmentVariable(name)

	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func unsetWindowsEnvironmentVariable(name string) error {
	command := fmt.Sprintf(
		`[Environment]::SetEnvironmentVariable("%s", $null, "User")`,
		name,
	)

	cmd := exec.Command(
		"powershell",
		"-NoProfile",
		"-Command",
		command,
	)

	return cmd.Run()
}

func unsetLinuxEnvironmentVariable(name string) error {
	command := fmt.Sprintf(
		`sed -i '/^export %s=/d' ~/.bashrc`,
		name,
	)

	cmd := exec.Command("bash", "-c", command)

	return cmd.Run()
}

func unsetMacEnvironmentVariable(name string) error {
	command := fmt.Sprintf(
		`sed -i '' '/^export %s=/d' ~/.zshrc`,
		name,
	)

	cmd := exec.Command("zsh", "-c", command)

	return cmd.Run()
}

func unsetEnvironmentVariableCommand(args []string, ctx *Context) bool {
	if len(args) < 1 {
		fmt.Println("Usage: unset <environment-variable-name>")
		return false
	}

	name := args[0]

	err := unsetPersistentEnvironmentVariable(name)

	if err != nil {
		fmt.Println("Error:", err)
		return false
	}

	fmt.Println("Environment variable removed:", name)

	return true
}