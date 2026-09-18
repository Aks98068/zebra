package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
)

type Process struct {
	PID  uint32
	PPID uint32
	Name string
}

// ─── List all processes ───────────────────────────────────────────────────────

func handleProcessList(args []string, ctx *Context) bool {
	pids, err := process.Pids() // ❌ was: process.pids() — lowercase p
	if err != nil {
		fmt.Println("Error getting PIDs:", err)
		return false
	}

	var processes []Process // ❌ was: []process — wrong type, lowercase

	for _, pid := range pids {
		p, err := process.NewProcess(pid)
		if err != nil {
			continue
		}

		name, err := p.Name()
		if err != nil {
			name = "unknown"
		}

		ppid, err := p.Ppid()
		if err != nil {
			ppid = 0
		}

		processes = append(processes, Process{ // ❌ was: process{} — lowercase
			PID:  uint32(pid),
			PPID: uint32(ppid),
			Name: name,
		})
	}

	for _, proc := range processes {
		fmt.Printf("PID: %-6d PPID: %-6d Name: %s\n", proc.PID, proc.PPID, proc.Name)
	}

	return true
}

// ─── Search by PID ────────────────────────────────────────────────────────────

func handleProcessPid(args []string, ctx *Context) bool { // ❌ was: wrong signature
	if len(args) < 1 {
		fmt.Println("Usage: ppid <pid>")
		return false
	}

	var pid int32
	fmt.Sscanf(args[0], "%d", &pid)

	p, err := process.NewProcess(pid)
	if err != nil {
		fmt.Printf("No process found with PID: %d\n", pid)
		return false
	}

	name, err := p.Name()
	if err != nil {
		name = "unknown"
	}

	ppid, err := p.Ppid()
	if err != nil {
		ppid = 0
	}

	result := &Process{
		PID:  uint32(pid),
		PPID: uint32(ppid),
		Name: name,
	}

	fmt.Printf("PID: %-6d PPID: %-6d Name: %s\n", result.PID, result.PPID, result.Name)
	return true
}

// ─── Search by Name ───────────────────────────────────────────────────────────

func handleProcessGrep(args []string, ctx *Context) bool { // ❌ was: wrong signature
	if len(args) < 1 {
		fmt.Println("Usage: pgrep <name>")
		return false
	}

	query := strings.ToLower(args[0])
	pids, err := process.Pids()
	if err != nil {
		fmt.Println("Error getting PIDs:", err)
		return false
	}

	found := 0
	for _, pid := range pids {
		p, err := process.NewProcess(pid)
		if err != nil {
			continue
		}

		procName, err := p.Name()
		if err != nil {
			continue
		}

		if strings.Contains(strings.ToLower(procName), query) {
			ppid, _ := p.Ppid()
			fmt.Printf("PID: %-6d PPID: %-6d Name: %s\n", pid, ppid, procName)
			found++
		}
	}

	if found == 0 {
		fmt.Printf("No process found matching: %s\n", args[0])
		return false
	}

	return true
}

// ─── Kill by PID ──────────────────────────────────────────────────────────────

func handleKillPid(args []string, ctx *Context) bool { // ❌ was: wrong signature
	if len(args) < 1 {
		fmt.Println("Usage: kill <pid>")
		return false
	}

	var pid int32
	fmt.Sscanf(args[0], "%d", &pid)

	p, err := process.NewProcess(pid)
	if err != nil {
		fmt.Printf("No process found with PID: %d\n", pid)
		return false
	}

	name, _ := p.Name()

	if err := p.Kill(); err != nil {
		fmt.Printf("Failed to kill PID %d (%s): %v\n", pid, name, err)
		return false
	}

	fmt.Printf("Successfully killed PID: %d (%s)\n", pid, name)
	return true
}

// ─── Kill by Name ─────────────────────────────────────────────────────────────

func handleKillName(args []string, ctx *Context) bool { // ❌ was: wrong signature
	if len(args) < 1 {
		fmt.Println("Usage: pkill <name>")
		return false
	}

	query := strings.ToLower(args[0])
	pids, err := process.Pids()
	if err != nil {
		fmt.Println("Error getting PIDs:", err)
		return false
	}

	killed := 0
	for _, pid := range pids {
		p, err := process.NewProcess(pid)
		if err != nil {
			continue
		}

		procName, err := p.Name()
		if err != nil {
			continue
		}

		if strings.Contains(strings.ToLower(procName), query) {
			if err := p.Kill(); err != nil {
				fmt.Printf("Failed to kill PID %d (%s): %v\n", pid, procName, err)
				continue
			}
			fmt.Printf("Killed PID: %d (%s)\n", pid, procName)
			killed++
		}
	}

	if killed == 0 {
		fmt.Printf("No process found matching: %s\n", args[0])
		return false
	}

	fmt.Printf("Total killed: %d process(es)\n", killed)
	return true
}

// ─── Kill self ────────────────────────────────────────────────────────────────

func handleKillSelf(args []string, ctx *Context) bool { // ❌ was: missing args []string
	fmt.Println("Terminating self, PID:", os.Getpid())
	os.Exit(0)
	return true
}