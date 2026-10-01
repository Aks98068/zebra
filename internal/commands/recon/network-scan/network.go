package networkscan

import (
	"fmt"
	"time"
)

func NetworkScan(args []string) bool {
	start := time.Now()

	fmt.Println()
	fmt.Println("Starting network reconnaissance...")
	fmt.Println()

	config, err := ParseScanArguments(args)
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}

	target, err := ValidateTarget(config.Target)
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}

	config.Target = target

	ports, err := ResolvePortSpecification(config)
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}

	config.Ports = ports

	var tcpResults []TCPResult
	var udpResults []UDPResult

	if config.TCP {
		fmt.Println()
		fmt.Println("Starting TCP scan...")

		tcpResults, err = RunTCPScan(config)
		if err != nil {
			fmt.Println("Error:", err)
			return false
		}
	}

	if config.UDP {
		fmt.Println()
		fmt.Println("Starting UDP scan...")

		udpResults, err = RunUDPScan(config)
		if err != nil {
			fmt.Println("Error:", err)
			return false
		}
	}

	scanResult := BuildScanResult(
		config,
		tcpResults,
		udpResults,
		start,
	)

	scanResult.Ports = DetectServices(
		config,
		scanResult.Ports,
	)

	scanResult.Ports = DetectVersions(
		config,
		scanResult.Ports,
	)

	scanResult.OS = DetectOS(
		config,
		scanResult.Ports,
	)

	scanResult.Duration = time.Since(start)

	fmt.Println()
	fmt.Println("Scan Results")
	fmt.Println("============")

	PrintPortResults(
		scanResult.Ports,
		config.ShowReason,
		config.OpenOnly,
	)

	PrintScanSummary(
		config,
		scanResult,
	)

	if err := WriteScanReport(
		config,
		scanResult,
	); err != nil {
		fmt.Println("Error:", err)
		return false
	}

	return true
}
