//go:build linux

package commands

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
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

	info.OS = linuxOSName()
	info.OsVersion = linuxOSVersion()
	info.Architecture = runtime.GOARCH
	info.Kernel = linuxKernel()
	info.Uptime = linuxUptime()

	info.CPU = linuxCPU()
	info.Memory = linuxMemory()
	info.Network = linuxNetwork()
	info.Disks = linuxDisks()

	info.Platform.Manufacturer = linuxReadFile(
		"/sys/class/dmi/id/sys_vendor",
	)

	info.Platform.Model = linuxReadFile(
		"/sys/class/dmi/id/product_name",
	)

	info.Platform.BIOSVersion = linuxReadFile(
		"/sys/class/dmi/id/bios_version",
	)

	info.Platform.OSBuild = info.OsVersion

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

func linuxReadFile(path string) string {
	data, err := os.ReadFile(path)

	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(data))
}

func linuxOSName() string {
	file, err := os.Open("/etc/os-release")

	if err != nil {
		return "Linux"
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "PRETTY_NAME=") {
			return strings.Trim(
				strings.TrimPrefix(line, "PRETTY_NAME="),
				`"`,
			)
		}
	}

	return "Linux"
}

func linuxOSVersion() string {
	file, err := os.Open("/etc/os-release")

	if err != nil {
		return ""
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "VERSION_ID=") {
			return strings.Trim(
				strings.TrimPrefix(line, "VERSION_ID="),
				`"`,
			)
		}
	}

	return ""
}

func linuxKernel() string {
	return linuxReadFile("/proc/sys/kernel/osrelease")
}

func linuxUptime() string {
	data, err := os.ReadFile("/proc/uptime")

	if err != nil {
		return ""
	}

	fields := strings.Fields(string(data))

	if len(fields) == 0 {
		return ""
	}

	seconds, err := strconv.ParseFloat(fields[0], 64)

	if err != nil {
		return ""
	}

	return linuxFormatDuration(
		time.Duration(seconds * float64(time.Second)),
	)
}

func linuxCPU() CPUInfo {
	var info CPUInfo

	info.Architecture = runtime.GOARCH

	file, err := os.Open("/proc/cpuinfo")

	if err != nil {
		return info
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	physicalCores := make(map[string]bool)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "model name") {
			if info.Model == "" {
				info.Model = strings.TrimSpace(
					strings.SplitN(line, ":", 2)[1],
				)
			}
		}

		if strings.HasPrefix(line, "vendor_id") {
			if info.Vendor == "" {
				info.Vendor = strings.TrimSpace(
					strings.SplitN(line, ":", 2)[1],
				)
			}
		}

		if strings.HasPrefix(line, "processor") {
			info.LogicalProcessors++
		}

		if strings.HasPrefix(line, "core id") {
			value := strings.TrimSpace(
				strings.SplitN(line, ":", 2)[1],
			)

			physicalCores[value] = true
		}

		if strings.HasPrefix(line, "cpu MHz") {
			if info.Frequency == "" {
				value := strings.TrimSpace(
					strings.SplitN(line, ":", 2)[1],
				)

				info.Frequency = value + " MHz"
			}
		}
	}

	info.Cores = int64(len(physicalCores))

	if info.Cores == 0 {
		info.Cores = info.LogicalProcessors
	}

	return info
}

func linuxMemory() MemoryInfo {
	var info MemoryInfo

	file, err := os.Open("/proc/meminfo")

	if err != nil {
		return info
	}

	defer file.Close()

	var total uint64
	var available uint64

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())

		if len(fields) < 2 {
			continue
		}

		value, err := strconv.ParseUint(fields[1], 10, 64)

		if err != nil {
			continue
		}

		value *= 1024

		switch fields[0] {
		case "MemTotal:":
			total = value

		case "MemAvailable:":
			available = value
		}
	}

	used := total - available

	info.Total = linuxFormatBytes(total)
	info.Available = linuxFormatBytes(available)
	info.Free = linuxFormatBytes(available)
	info.Used = linuxFormatBytes(used)

	if total > 0 {
		info.Usage = fmt.Sprintf(
			"%.2f%%",
			float64(used)/float64(total)*100,
		)
	}

	return info
}

func linuxNetwork() []NetworkInfo {
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

func linuxDisks() []DiskInfo {
	file, err := os.Open("/proc/mounts")

	if err != nil {
		return nil
	}

	defer file.Close()

	var result []DiskInfo

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())

		if len(fields) < 3 {
			continue
		}

		device := fields[0]
		mountPoint := fields[1]
		fileSystem := fields[2]

		var stat syscall.Statfs_t

		err := syscall.Statfs(
			mountPoint,
			&stat,
		)

		if err != nil {
			continue
		}

		total := stat.Blocks * uint64(stat.Bsize)
		free := stat.Bavail * uint64(stat.Bsize)
		used := total - free

		if total == 0 {
			continue
		}

		usage := float64(used) / float64(total) * 100

		result = append(result, DiskInfo{
			Device:     device,
			MountPoint: mountPoint,
			FileSystem: fileSystem,
			Total:      linuxFormatBytes(total),
			Used:       linuxFormatBytes(used),
			Free:       linuxFormatBytes(free),
			Usage:      fmt.Sprintf("%.2f%%", usage),
		})
	}

	return result
}

func linuxFormatBytes(bytes uint64) string {
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

func linuxFormatDuration(duration time.Duration) string {
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
