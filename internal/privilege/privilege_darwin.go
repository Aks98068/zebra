//go:build darwin

package privilege

import (
	"fmt"
	"os"
	"os/exec"
)

func EnsureElevated() error {
	if os.Geteuid() == 0 {
		return nil
	}

	return relaunchAsRoot()
}

func relaunchAsRoot() error {
	executable, err := os.Executable()

	if err != nil {
		return fmt.Errorf("unable to find Zebra executable: %w", err)
	}

	args := append(
		[]string{executable},
		os.Args[1:]...,
	)

	cmd := exec.Command(
		"sudo",
		args...,
	)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()

	if err != nil {
		return fmt.Errorf("failed to obtain root privileges: %w", err)
	}

	os.Exit(0)

	return nil
}