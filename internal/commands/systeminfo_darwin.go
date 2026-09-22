//go:build darwin

package commands

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// ============================================================
// SHARED SYSTEM INFORMATION STRUCTURES
// ============================================================

type Systeminfo struct {
	Hostname     string
	OS           string
	OsVersion    string
	Architecture string
	Kernel       string
	Uptime       string

	CPU     CPUInfo
	Memory  MemoryInfo
	Network []NetworkInfo
	Disks   []DiskInfo

	Platform Platforminfo
}

type CPUInfo struct {
	Cores             int64
	LogicalProcessors int64
	Architecture      string
	Model             string
	Vendor            string
	Frequency         string
}

type MemoryInfo struct {
	Total     string
	Available string
	Used      string
	Free      string
	Usage     string
}

type NetworkInfo struct {
	Name      string
	MAC       string
	Status    string
	IPAddress []string
}

type DiskInfo struct {
	Device     string
	MountPoint string
	FileSystem string
	Total      string
	Used       string
	Free       string
	Usage      string
}

type Platforminfo struct {
	Manufacturer string
	Model        string
	BIOSVersion  string
	OSBuild      string
	InstallDate  string
}

func SystemInfo(args []string, ctx *Context) bool {
	var info Systeminfo

	info.Hostname, _ = os.Hostname()

	info.OS = "macOS"
	info.OsVersion = darwinCommand("sw_vers", "-productVersion")
	info.Architecture = runtime.GOARCH
	info.Kernel = darwinCommand("uname", "-r")
	info.Uptime = darwinUptime()

	info.CPU = darwinCPU()
	info.Memory = darwinMemory()
	info.Network = darwinNetwork()
	info.Disks = darwinDisks()

	info.Platform.Manufacturer = "Apple"
	info.Platform.Model = darwinCommand("sysctl", "-n", "hw.model")
	info.Platform.OSBuild = darwinCommand("sw_vers", "-buildVersion")

	fmt.Println()
	fmt.Println("============================================================")
	fmt.Println("                     SYSTEM INFORMATION")
	fmt.Println("============================================================")

	fmt.Printf("Hostname       : %s\n", info.Hostname)
	fmt.Printf("OS             : %s\n", info.OS)
	fmt.Printf("OS Version     : %s\n", info.OsVersion)
	fmt.Printf("Architecture   : %s\n", info.Architecture)
	fmt.Printf("Kernel         : %s\n", info.Kernel)
	fmt.Printf("Uptime         : %s\n", info.Uptime)

	fmt.Println()
	fmt.Println("------------------------- CPU -------------------------------")

	fmt.Printf("Model          : %s\n", info.CPU.Model)
	fmt.Printf("Vendor         : %s\n", info.CPU.Vendor)
	fmt.Printf("Cores          : %d\n", info.CPU.Cores)
	fmt.Printf("Logical CPUs   : %d\n", info.CPU.LogicalProcessors)
	fmt.Printf("Architecture   : %s\n", info.CPU.Architecture)
	fmt.Printf("Frequency      : %s\n", info.CPU.Frequency)

	fmt.Println()
	fmt.Println("------------------------ MEMORY -----------------------------")

	fmt.Printf("Total          : %s\n", info.Memory.Total)
	fmt.Printf("Used           : %s\n", info.Memory.Used)
	fmt.Printf("Available      : %s\n", info.Memory.Available)
	fmt.Printf("Free           : %s\n", info.Memory.Free)
	fmt.Printf("Usage          : %s\n", info.Memory.Usage)

	fmt.Println()
	fmt.Println("------------------------ NETWORK ----------------------------")

	for _, network := range info.Network {
		fmt.Printf("Interface      : %s\n", network.Name)
		fmt.Printf("Status         : %s\n", network.Status)
		fmt.Printf("MAC            : %s\n", network.MAC)

		for _, ip := range network.IPAddress {
			fmt.Printf("IP Address     : %s\n", ip)
		}

		fmt.Println()
	}

	fmt.Println("------------------------- DISKS -----------------------------")

	for _, disk := range info.Disks {
		fmt.Printf("Device         : %s\n", disk.Device)
		fmt.Printf("Mount Point    : %s\n", disk.MountPoint)
		fmt.Printf("File System    : %s\n", disk.FileSystem)
		fmt.Printf("Total          : %s\n", disk.Total)
		fmt.Printf("Used           : %s\n", disk.Used)
		fmt.Printf("Free           : %s\n", disk.Free)
		fmt.Printf("Usage          : %s\n", disk.Usage)
		fmt.Println()
	}

	fmt.Println("----------------------- PLATFORM ----------------------------")

	fmt.Printf("Manufacturer   : %s\n", info.Platform.Manufacturer)
	fmt.Printf("Model          : %s\n", info.Platform.Model)
	fmt.Printf("BIOS Version   : %s\n", info.Platform.BIOSVersion)
	fmt.Printf("OS Build       : %s\n", info.Platform.OSBuild)
	fmt.Printf("Install Date   : %s\n", info.Platform.InstallDate)

	fmt.Println("============================================================")
	fmt.Println()

	return true
}

func darwinCommand(command string, args ...string) string {
	output, err := exec.Command(command, args...).Output()

	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}

func darwinCPU() CPUInfo {
	var info CPUInfo

	info.Architecture = runtime.GOARCH
	info.Model = darwinCommand("sysctl", "-n", "machdep.cpu.brand_string")
	info.Vendor = darwinCommand("sysctl", "-n", "machdep.cpu.vendor")

	info.Cores = int64(
		darwinInt("hw.physicalcpu"),
	)

	info.LogicalProcessors = int64(
		darwinInt("hw.logicalcpu"),
	)

	frequency := darwinInt64("hw.cpufrequency")

	if frequency > 0 {
		info.Frequency = fmt.Sprintf(
			"%.2f GHz",
			float64(frequency)/1_000_000_000,
		)
	}

	return info
}

func darwinMemory() MemoryInfo {
	var info MemoryInfo

	total := uint64(
		darwinInt64("hw.memsize"),
	)

	if total == 0 {
		return info
	}

	info.Total = darwinFormatBytes(total)

	// vm_stat reports page statistics.
	vmstat := darwinCommand("vm_stat")

	pageSize := uint64(4096)

	var freePages uint64
	var inactivePages uint64
	var speculativePages uint64

	for _, line := range strings.Split(vmstat, "\n") {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Pages free:") {
			freePages = darwinVMStatValue(line)
		}

		if strings.HasPrefix(line, "Pages inactive:") {
			inactivePages = darwinVMStatValue(line)
		}

		if strings.HasPrefix(line, "Pages speculative:") {
			speculativePages = darwinVMStatValue(line)
		}
	}

	available := (freePages + inactivePages + speculativePages) * pageSize

	if available > total {
		available = total
	}

	used := total - available

	info.Available = darwinFormatBytes(available)
	info.Free = darwinFormatBytes(available)
	info.Used = darwinFormatBytes(used)

	info.Usage = fmt.Sprintf(
		"%.2f%%",
		float64(used)/float64(total)*100,
	)

	return info
}

func darwinVMStatValue(line string) uint64 {
	parts := strings.SplitN(line, ":", 2)

	if len(parts) != 2 {
		return 0
	}

	value := strings.TrimSpace(
		strings.TrimSuffix(
			parts[1],
			".",
		),
	)

	result, err := strconv.ParseUint(value, 10, 64)

	if err != nil {
		return 0
	}

	return result
}

func darwinInt(name string) int {
	return int(darwinInt64(name))
}

func darwinInt64(name string) int64 {
	value := darwinCommand(
		"sysctl",
		"-n",
		name,
	)

	result, err := strconv.ParseInt(
		strings.TrimSpace(value),
		10,
		64,
	)

	if err != nil {
		return 0
	}

	return result
}

func darwinUptime() string {
	output := darwinCommand(
		"sysctl",
		"-n",
		"kern.boottime",
	)

	if output == "" {
		return ""
	}

	start := strings.Index(output, "sec = ")

	if start == -1 {
		return ""
	}

	value := output[start+6:]

	end := strings.Index(value, ",")

	if end != -1 {
		value = value[:end]
	}

	seconds, err := strconv.ParseInt(
		strings.TrimSpace(value),
		10,
		64,
	)

	if err != nil {
		return ""
	}

	uptime := time.Since(
		time.Unix(seconds, 0),
	)

	return darwinFormatDuration(uptime)
}

func darwinNetwork() []NetworkInfo {
	interfaces, err := net.Interfaces()

	if err != nil {
		return nil
	}

	var result []NetworkInfo

	for _, iface := range interfaces {
		addresses, err := iface.Addrs()

		if err != nil {
			continue
		}

		var ips []string

		for _, address := range addresses {
			ip := ""

			switch value := address.(type) {
			case *net.IPNet:
				ip = value.IP.String()

			case *net.IPAddr:
				ip = value.IP.String()
			}

			if ip == "" {
				continue
			}

			if strings.HasPrefix(ip, "169.254.") {
				continue
			}

			ips = append(ips, ip)
		}

		if len(ips) == 0 {
			continue
		}

		status := "DOWN"

		if iface.Flags&net.FlagUp != 0 {
			status = "UP"
		}

		result = append(result, NetworkInfo{
			Name:      iface.Name,
			MAC:       iface.HardwareAddr.String(),
			Status:    status,
			IPAddress: ips,
		})
	}

	return result
}

func darwinDisks() []DiskInfo {
	output := darwinCommand(
		"df",
		"-kP",
	)

	var result []DiskInfo

	lines := strings.Split(output, "\n")

	for _, line := range lines[1:] {
		fields := strings.Fields(line)

		if len(fields) < 6 {
			continue
		}

		device := fields[0]

		totalKB, err1 := strconv.ParseUint(fields[1], 10, 64)
		usedKB, err2 := strconv.ParseUint(fields[2], 10, 64)
		freeKB, err3 := strconv.ParseUint(fields[3], 10, 64)

		if err1 != nil || err2 != nil || err3 != nil {
			continue
		}

		mountPoint := strings.Join(fields[5:], " ")

		total := totalKB * 1024
		used := usedKB * 1024
		free := freeKB * 1024

		usage := float64(used) / float64(total) * 100

		result = append(result, DiskInfo{
			Device:     device,
			MountPoint: mountPoint,
			FileSystem: "APFS/HFS+",
			Total:      darwinFormatBytes(total),
			Used:       darwinFormatBytes(used),
			Free:       darwinFormatBytes(free),
			Usage:      fmt.Sprintf("%.2f%%", usage),
		})
	}

	return result
}

func darwinFormatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))

	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))

	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))

	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))

	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func darwinFormatDuration(duration time.Duration) string {
	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	return fmt.Sprintf(
		"%dd %02dh %02dm %02ds",
		days,
		hours,
		minutes,
		seconds,
	)
}
