//go:build linux

package privilege

import (
	"fmt"
	"os"
	"os/exec"
)

// ============================================================
// ENSURE ELEVATED
// ============================================================
//
// If Zebra is already running as root:
//     return nil
//
// Otherwise:
//     relaunch Zebra through sudo
//     terminate the original process
//
// This ensures that only ONE Zebra process continues running.
// ============================================================

func EnsureElevated() error {
	// Already running as root.
	if os.Geteuid() == 0 {
		return nil
	}

	return relaunchAsRoot()
}

// ============================================================
// RELAUNCH AS ROOT
// ============================================================

func relaunchAsRoot() error {
	// --------------------------------------------------------
	// FIND ZEBRA EXECUTABLE
	// --------------------------------------------------------

	executable, err := os.Executable()

	if err != nil {
		return fmt.Errorf(
			"unable to find Zebra executable: %w",
			err,
		)
	}

	// --------------------------------------------------------
	// MAKE SURE SUDO EXISTS
	// --------------------------------------------------------

	sudoPath, err := exec.LookPath("sudo")

	if err != nil {
		return fmt.Errorf(
			"sudo is not installed or not available in PATH: %w",
			err,
		)
	}

	// --------------------------------------------------------
	// PRESERVE ORIGINAL ARGUMENTS
	// --------------------------------------------------------

	args := make([]string, 0, len(os.Args))

	// First argument is the Zebra executable.
	args = append(args, executable)

	// Preserve all original Zebra arguments.
	args = append(args, os.Args[1:]...)

	// --------------------------------------------------------
	// START ELEVATED ZEBRA
	// --------------------------------------------------------

	cmd := exec.Command(
		sudoPath,
		args...,
	)

	// Use the same terminal.
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// --------------------------------------------------------
	// RUN ELEVATED PROCESS
	// --------------------------------------------------------

	if err := cmd.Run(); err != nil {
		return fmt.Errorf(
			"failed to obtain root privileges: %w",
			err,
		)
	}

	// --------------------------------------------------------
	// IMPORTANT
	//
	// The sudo/root Zebra process has finished.
	//
	// The original non-root Zebra process must terminate.
	// --------------------------------------------------------

	os.Exit(0)

	return nil
}