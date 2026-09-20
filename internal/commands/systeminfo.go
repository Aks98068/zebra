package commands

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"time"
	"unsafe"

	"github.com/yusufpapurcu/wmi"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

type Systeminfo struct {
	Hostname     string        `json:"hostname"`
	OS           string        `json:"os"`
	OsVersion    string        `json:"os_version"`
	Architecture string        `json:"architecture"`
	Kernel       string        `json:"kernel"`
	Uptime       string        `json:"uptime"`
	CPU          CPUInfo       `json:"cpu"`
	Memory       MemoryInfo    `json:"memory"`
	Network      []NetworkInfo `json:"network"`
	Disks        []DiskInfo    `json:"disks"`
	Platform     Platforminfo  `json:"platform"`
}

type CPUInfo struct {
	Cores        int64  `json:"cores"`
	Architecture string `json:"architecture"`
	Model        string `json:"model"`
	Vendor       string `json:"vendor"`
	Frequency    string `json:"frequency"`
}

type MemoryInfo struct {
	Total     int64 `json:"total_bytes"`
	Available int64 `json:"available_bytes"`
	Used      int64 `json:"used_bytes"`
}

type NetworkInfo struct {
	Name      string `json:"name"`
	MAC       string `json:"mac"`
	Status    string `json:"status"`
	IPAddress string `json:"ip_address"`
}

type DiskInfo struct {
	Path      string `json:"path"`
	Total     int64  `json:"total_bytes"`
	Free      int64  `json:"free_bytes"`
	Available int64  `json:"available_bytes"`
}

type Platforminfo struct {
	Manufacturer string `json:"manufacturer,omitempty"`
	Model        string `json:"model,omitempty"`
	BIOS         string `json:"bios,omitempty"`
	OSBuild      string `json:"os_build,omitempty"`
	InstallDate  string `json:"install_date,omitempty"`
}

func SystemInfo(args []string, ctx *Context) bool {
	switch runtime.GOOS {
	case "windows":
		info, err := Windowsinfo()
		if err != nil {
			fmt.Printf("failed to collect Windows system information: %v\n", err)
			return false
		}

		fmt.Printf("%+v\n", info)

	case "linux":
		linuxinfo()

	case "darwin":
		Darwininfo()

	default:
		fmt.Println("invalid os detected")
	}

	return false
}

func Windowsinfo() (Systeminfo, error) {

	// Get the computer name from Windows.
	hostname, err := os.Hostname()
	if err != nil {
		return Systeminfo{}, fmt.Errorf("get hostname: %w", err)
	}

	osname := "Windows"
	architecture := runtime.GOARCH

	// Windows keeps most of its basic version information in this registry key.
	key, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows NT\CurrentVersion`,
		registry.QUERY_VALUE,
	)
	if err != nil {
		return Systeminfo{}, fmt.Errorf(
			"open Windows version registry key: %w",
			err,
		)
	}
	defer key.Close()

	// This gives us values such as 24H2, depending on the Windows release.
	displayVersion, _, err := key.GetStringValue("DisplayVersion")
	if err != nil {
		return Systeminfo{}, fmt.Errorf(
			"read Windows display version: %w",
			err,
		)
	}

	// Read the Windows version numbers and build number.
	major, _, err := key.GetIntegerValue("CurrentMajorVersionNumber")
	if err != nil {
		return Systeminfo{}, fmt.Errorf(
			"read Windows major version: %w",
			err,
		)
	}

	minor, _, err := key.GetIntegerValue("CurrentMinorVersionNumber")
	if err != nil {
		return Systeminfo{}, fmt.Errorf(
			"read Windows minor version: %w",
			err,
		)
	}

	build, _, err := key.GetStringValue("CurrentBuildNumber")
	if err != nil {
		return Systeminfo{}, fmt.Errorf(
			"read Windows build number: %w",
			err,
		)
	}

	// Combine the numbers into the familiar Windows kernel version format.
	kernel := strconv.FormatUint(major, 10) +
		"." +
		strconv.FormatUint(minor, 10) +
		"." +
		build

	// kernel32.dll contains the Windows APIs we need for uptime and memory.
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")

	// GetTickCount64 returns the number of milliseconds since Windows started.
	getTickCount64 := kernel32.NewProc("GetTickCount64")

	result, _, _ := getTickCount64.Call()

	milliseconds := uint64(result)
	duration := time.Duration(milliseconds) * time.Millisecond
	uptime := duration.String()

	// Query WMI for the processor details.
	var processors []struct {
		Name                      string
		Manufacturer              string
		MaxClockSpeed             uint32
		NumberOfCores             uint32
		NumberOfLogicalProcessors uint32
	}

	err = wmi.Query(
		"SELECT Name, Manufacturer, MaxClockSpeed, NumberOfCores, NumberOfLogicalProcessors FROM Win32_Processor",
		&processors,
	)
	if err != nil {
		return Systeminfo{}, fmt.Errorf(
			"query CPU information: %w",
			err,
		)
	}

	if len(processors) == 0 {
		return Systeminfo{}, fmt.Errorf("no CPU information found")
	}

	processor := processors[0]

	cpuInfo := CPUInfo{
		Cores:        int64(processor.NumberOfCores),
		Architecture: architecture,
		Model:        processor.Name,
		Vendor:       processor.Manufacturer,
		Frequency:    fmt.Sprintf("%d MHz", processor.MaxClockSpeed),
	}

	// MEMORYSTATUSEX is the structure expected by GlobalMemoryStatusEx.
	// We define it here because the installed x/sys/windows version does
	// not expose the structure and function under the names we originally used.
	type memoryStatusEx struct {
		DwLength                uint32
		DwMemoryLoad            uint32
		UllTotalPhys            uint64
		UllAvailPhys            uint64
		UllTotalPageFile        uint64
		UllAvailPageFile        uint64
		UllTotalVirtual         uint64
		UllAvailVirtual         uint64
		UllAvailExtendedVirtual uint64
	}

	// Windows requires the size of the structure before calling the API.
	memoryStatus := memoryStatusEx{
		DwLength: uint32(unsafe.Sizeof(memoryStatusEx{})),
	}

	globalMemoryStatusEx := kernel32.NewProc("GlobalMemoryStatusEx")

	result, _, callErr := globalMemoryStatusEx.Call(
		uintptr(unsafe.Pointer(&memoryStatus)),
	)

	if result == 0 {
		return Systeminfo{}, fmt.Errorf(
			"get memory information: %w",
			callErr,
		)
	}

	totalMemory := int64(memoryStatus.UllTotalPhys)
	availableMemory := int64(memoryStatus.UllAvailPhys)
	usedMemory := totalMemory - availableMemory

	memoryInfo := MemoryInfo{
		Total:     totalMemory,
		Available: availableMemory,
		Used:      usedMemory,
	}

	// Get all network adapters visible to Windows.
	interfaces, err := net.Interfaces()
	if err != nil {
		return Systeminfo{}, fmt.Errorf(
			"get network interfaces: %w",
			err,
		)
	}

	var networkInfo []NetworkInfo

	for _, iface := range interfaces {

		// Loopback is normally not useful when displaying physical
		// and virtual network adapters, so leave it out.
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		name := iface.Name
		mac := iface.HardwareAddr.String()

		status := "down"
		if iface.Flags&net.FlagUp != 0 {
			status = "up"
		}

		addresses, err := iface.Addrs()
		if err != nil {
			return Systeminfo{}, fmt.Errorf(
				"get addresses for interface %s: %w",
				iface.Name,
				err,
			)
		}

		// An adapter may exist without currently having an IP address.
		if len(addresses) == 0 {
			networkInfo = append(networkInfo, NetworkInfo{
				Name:      name,
				MAC:       mac,
				Status:    status,
				IPAddress: "",
			})

			continue
		}

		// Keep one NetworkInfo entry for each IP address.
		// This also handles adapters that have both IPv4 and IPv6.
		for _, addr := range addresses {

			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				continue
			}

			networkInfo = append(networkInfo, NetworkInfo{
				Name:      name,
				MAC:       mac,
				Status:    status,
				IPAddress: ip.String(),
			})
		}
	}

	// Return everything collected so far.
	// Disk and platform information will be added later.
	return Systeminfo{
		Hostname:     hostname,
		OS:           osname,
		OsVersion:    displayVersion,
		Architecture: architecture,
		Kernel:       kernel,
		Uptime:       uptime,
		CPU:          cpuInfo,
		Memory:       memoryInfo,
		Network:      networkInfo,
	}, nil
}

func linuxinfo() {

}

func Darwininfo() {

}
