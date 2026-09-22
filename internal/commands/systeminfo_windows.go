//go:build windows

package commands

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/yusufpapurcu/wmi"
	"golang.org/x/sys/windows/registry"
)

// ============================================================
// WINDOWS SYSTEM INFO
// ============================================================

type Systeminfo struct {
	Hostname     string
	OS           string
	Architecture string
	OsVersion    string
	Kernel       string
	Uptime       string
	CPU          CPUInfo
	Memory       MemoryInfo
	Network      []NetworkInfo
	Disks        []DiskInfo
	Platform     PlatformInfo
}

type CPUInfo struct {
	Model             string
	Vendor            string
	Cores             int64
	LogicalProcessors int64
	Architecture      string
	Frequency         string
}

type MemoryInfo struct {
	Total     string
	Used      string
	Available string
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

type PlatformInfo struct {
	Manufacturer string
	Model        string
	BIOSVersion  string
	OSBuild      string
	InstallDate  string
}

func SystemInfo(args []string, ctx *Context) bool {
	info, err := Windowsinfo()

	if err != nil {
		fmt.Println("Error collecting system information:", err)
		return false
	}

	printWindowsSystemInfo(info)

	return true
}

// ============================================================
// WINDOWS INFORMATION COLLECTION
// ============================================================

func Windowsinfo() (Systeminfo, error) {
	var info Systeminfo

	// --------------------------------------------------------
	// Hostname
	// --------------------------------------------------------

	hostname, err := os.Hostname()

	if err != nil {
		hostname = "Unknown"
	}

	info.Hostname = hostname

	// --------------------------------------------------------
	// Operating System
	// --------------------------------------------------------

	info.OS = "Windows"
	info.Architecture = runtime.GOARCH

	info.OsVersion = windowsRegistryValue(
		registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows NT\CurrentVersion`,
		"DisplayVersion",
	)

	if info.OsVersion == "" {
		info.OsVersion = windowsRegistryValue(
			registry.LOCAL_MACHINE,
			`SOFTWARE\Microsoft\Windows NT\CurrentVersion`,
			"ReleaseId",
		)
	}

	// --------------------------------------------------------
	// Windows build
	// --------------------------------------------------------

	build := windowsRegistryValue(
		registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows NT\CurrentVersion`,
		"CurrentBuild",
	)

	ubr := windowsRegistryDWORD(
		registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows NT\CurrentVersion`,
		"UBR",
	)

	if build != "" {
		if ubr != "" {
			info.Platform.OSBuild = build + "." + ubr
		} else {
			info.Platform.OSBuild = build
		}
	}

	// --------------------------------------------------------
	// Operating system WMI information
	// --------------------------------------------------------

	var operatingSystems []struct {
		Version        string
		InstallDate    string
		LastBootUpTime string
	}

	err = wmi.Query(
		"SELECT Version, InstallDate, LastBootUpTime FROM Win32_OperatingSystem",
		&operatingSystems,
	)

	if err == nil && len(operatingSystems) > 0 {
		info.Kernel = operatingSystems[0].Version

		if operatingSystems[0].InstallDate != "" {
			info.Platform.InstallDate = windowsWMIDate(
				operatingSystems[0].InstallDate,
			)
		}

		if operatingSystems[0].LastBootUpTime != "" {
			info.Uptime = windowsUptimeFromBootTime(
				operatingSystems[0].LastBootUpTime,
			)
		}
	}

	// --------------------------------------------------------
	// CPU
	// --------------------------------------------------------

	var processors []struct {
		Name                      string
		Manufacturer              string
		NumberOfCores             uint32
		NumberOfLogicalProcessors uint32
		MaxClockSpeed             uint32
	}

	err = wmi.Query(
		"SELECT Name, Manufacturer, NumberOfCores, NumberOfLogicalProcessors, MaxClockSpeed FROM Win32_Processor",
		&processors,
	)

	if err == nil && len(processors) > 0 {
		cpu := processors[0]

		info.CPU.Model = strings.TrimSpace(cpu.Name)
		info.CPU.Vendor = strings.TrimSpace(cpu.Manufacturer)
		info.CPU.Cores = int64(cpu.NumberOfCores)
		info.CPU.LogicalProcessors = int64(cpu.NumberOfLogicalProcessors)
		info.CPU.Architecture = runtime.GOARCH

		if cpu.MaxClockSpeed > 0 {
			info.CPU.Frequency = fmt.Sprintf(
				"%d MHz",
				cpu.MaxClockSpeed,
			)
		}
	}

	// --------------------------------------------------------
	// Memory
	// --------------------------------------------------------

	var memories []struct {
		TotalVisibleMemorySize uint64
		FreePhysicalMemory     uint64
	}

	err = wmi.Query(
		"SELECT TotalVisibleMemorySize, FreePhysicalMemory FROM Win32_OperatingSystem",
		&memories,
	)

	if err == nil && len(memories) > 0 {
		total := memories[0].TotalVisibleMemorySize * 1024
		free := memories[0].FreePhysicalMemory * 1024

		var used uint64

		if total >= free {
			used = total - free
		}

		info.Memory.Total = windowsFormatBytes(total)
		info.Memory.Free = windowsFormatBytes(free)
		info.Memory.Available = windowsFormatBytes(free)
		info.Memory.Used = windowsFormatBytes(used)

		if total > 0 {
			info.Memory.Usage = fmt.Sprintf(
				"%.2f%%",
				float64(used)/float64(total)*100,
			)
		}
	}

	// --------------------------------------------------------
	// Network
	// --------------------------------------------------------

	info.Network = windowsNetworkInfo()

	// --------------------------------------------------------
	// Disks
	// --------------------------------------------------------

	info.Disks = windowsDiskInfo()

	// --------------------------------------------------------
	// Computer platform
	// --------------------------------------------------------

	var computers []struct {
		Manufacturer string
		Model        string
	}

	err = wmi.Query(
		"SELECT Manufacturer, Model FROM Win32_ComputerSystem",
		&computers,
	)

	if err == nil && len(computers) > 0 {
		info.Platform.Manufacturer =
			strings.TrimSpace(computers[0].Manufacturer)

		info.Platform.Model =
			strings.TrimSpace(computers[0].Model)
	}

	// --------------------------------------------------------
	// BIOS
	// --------------------------------------------------------

	var bios []struct {
		SMBIOSBIOSVersion string
	}

	err = wmi.Query(
		"SELECT SMBIOSBIOSVersion FROM Win32_BIOS",
		&bios,
	)

	if err == nil && len(bios) > 0 {
		info.Platform.BIOSVersion =
			strings.TrimSpace(bios[0].SMBIOSBIOSVersion)
	}

	return info, nil
}

// ============================================================
// WINDOWS NETWORK
// ============================================================

func windowsNetworkInfo() []NetworkInfo {
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
			var ip string

			switch value := address.(type) {
			case *net.IPNet:
				ip = value.IP.String()

			case *net.IPAddr:
				ip = value.IP.String()
			}

			if ip == "" {
				continue
			}

			// Ignore APIPA addresses.
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

// ============================================================
// WINDOWS DISKS
// ============================================================

func windowsDiskInfo() []DiskInfo {
	var disks []struct {
		DeviceID   string
		FileSystem string
		Size       uint64
		FreeSpace  uint64
	}

	err := wmi.Query(
		"SELECT DeviceID, FileSystem, Size, FreeSpace FROM Win32_LogicalDisk WHERE DriveType = 3",
		&disks,
	)

	if err != nil {
		return nil
	}

	var result []DiskInfo

	for _, disk := range disks {
		if disk.Size == 0 {
			continue
		}

		var used uint64

		if disk.Size >= disk.FreeSpace {
			used = disk.Size - disk.FreeSpace
		}

		usage := float64(used) /
			float64(disk.Size) *
			100

		result = append(result, DiskInfo{
			Device:     disk.DeviceID,
			MountPoint: disk.DeviceID,
			FileSystem: disk.FileSystem,
			Total:      windowsFormatBytes(disk.Size),
			Used:       windowsFormatBytes(used),
			Free:       windowsFormatBytes(disk.FreeSpace),
			Usage:      fmt.Sprintf("%.2f%%", usage),
		})
	}

	return result
}

// ============================================================
// WINDOWS REGISTRY
// ============================================================

func windowsRegistryValue(
	root registry.Key,
	path string,
	name string,
) string {
	key, err := registry.OpenKey(
		root,
		path,
		registry.QUERY_VALUE,
	)

	if err != nil {
		return ""
	}

	defer key.Close()

	value, _, err := key.GetStringValue(name)

	if err != nil {
		return ""
	}

	return strings.TrimSpace(value)
}

func windowsRegistryDWORD(
	root registry.Key,
	path string,
	name string,
) string {
	key, err := registry.OpenKey(
		root,
		path,
		registry.QUERY_VALUE,
	)

	if err != nil {
		return ""
	}

	defer key.Close()

	value, _, err := key.GetIntegerValue(name)

	if err != nil {
		return ""
	}

	return strconv.FormatUint(value, 10)
}

// ============================================================
// WINDOWS WMI DATE
// ============================================================

func windowsWMIDate(value string) string {
	value = strings.TrimSpace(value)

	if len(value) < 14 {
		return value
	}

	year := value[0:4]
	month := value[4:6]
	day := value[6:8]
	hour := value[8:10]
	minute := value[10:12]
	second := value[12:14]

	return fmt.Sprintf(
		"%s-%s-%s %s:%s:%s",
		year,
		month,
		day,
		hour,
		minute,
		second,
	)
}

// ============================================================
// WINDOWS UPTIME
// ============================================================

func windowsUptimeFromBootTime(value string) string {
	if len(value) < 14 {
		return ""
	}

	year, err1 := strconv.Atoi(value[0:4])
	month, err2 := strconv.Atoi(value[4:6])
	day, err3 := strconv.Atoi(value[6:8])
	hour, err4 := strconv.Atoi(value[8:10])
	minute, err5 := strconv.Atoi(value[10:12])
	second, err6 := strconv.Atoi(value[12:14])

	if err1 != nil ||
		err2 != nil ||
		err3 != nil ||
		err4 != nil ||
		err5 != nil ||
		err6 != nil {
		return ""
	}

	bootTime := time.Date(
		year,
		time.Month(month),
		day,
		hour,
		minute,
		second,
		0,
		time.Local,
	)

	duration := time.Since(bootTime)

	if duration < 0 {
		return ""
	}

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

// ============================================================
// WINDOWS BYTE FORMAT
// ============================================================

func windowsFormatBytes(bytes uint64) string {
	const (
		KB = uint64(1024)
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf(
			"%.2f TB",
			float64(bytes)/float64(TB),
		)

	case bytes >= GB:
		return fmt.Sprintf(
			"%.2f GB",
			float64(bytes)/float64(GB),
		)

	case bytes >= MB:
		return fmt.Sprintf(
			"%.2f MB",
			float64(bytes)/float64(MB),
		)

	case bytes >= KB:
		return fmt.Sprintf(
			"%.2f KB",
			float64(bytes)/float64(KB),
		)

	default:
		return fmt.Sprintf(
			"%d B",
			bytes,
		)
	}
}

// ============================================================
// WINDOWS OUTPUT
// ============================================================

func printWindowsSystemInfo(info Systeminfo) {
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
}
