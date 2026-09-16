package commands

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func getAllEnvironmentVariables() {
	// Get all environment variables
	variables := os.Environ()

	// Loop through all environment variables
	for _, variable := range variables {

		// Split into name and value
		parts := strings.SplitN(variable, "=", 2)

		// Make sure we have both name and value
		if len(parts) == 2 {
			name := parts[0]
			value := parts[1]

			fmt.Println("Name:", name)
			fmt.Println("Value:", value)
			fmt.Println()
		}
	}
}

func getAllEnvironmentVariable(name string) {
	value, exists := os.LookupEnv(name)

	if exists {
		fmt.Println("variable exists: ", value)
	} else {
		fmt.Println("variabler doesnot exists", name)
	}
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
	// Windows implementation
	command := fmt.Sprintf(
		`[Environment]::SetEnvironmentVariable("%s","%s", "User")`,
		name,
		value,
	)

	cmd := exec.Command("powershell", "-NoProfile", "-Command", command)

	err := cmd.Run()
	return err
}

func linuxfun(name string, value string) error {
	// Linux implementation
	command := fmt.Sprintf(
		`echo 'export %s="%s"' >> ~/.bashrc`, name, value,
	)

	cmd := exec.Command("bash", "-c", command)

	err := cmd.Run()
	return err
}

func macfun(name string, value string) error {
	// macOS implementation
	command := fmt.Sprintf(`echo 'export %s="%s"' >> ~/.zshrc`, name, value)
	cmd := exec.Command("zsh", "-c", command)
	error := cmd.Run()

	return error
}
