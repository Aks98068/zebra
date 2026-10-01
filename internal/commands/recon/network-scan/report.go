package networkscan

import (
	"fmt"
	"strings"
	"time"
)

func PrintScanSummary(
	config ScanConfig,
	result ScanResult,
) {
	fmt.Println()
	fmt.Println("ZEBRA Network Scan")
	fmt.Println("==================")
	fmt.Println()

	fmt.Printf("Target:        %s\n", result.Target)
	fmt.Printf("Duration:      %s\n", result.Duration.Round(time.Millisecond))
	fmt.Printf("Ports scanned: %d\n", result.TotalPorts)
	fmt.Printf("Open:          %d\n", result.FilterAnalysis.Open)
	fmt.Printf("Closed:        %d\n", result.FilterAnalysis.Closed)
	fmt.Printf("Filtered:      %d\n", result.FilterAnalysis.Filtered)
	fmt.Printf("Open|Filtered: %d\n", result.FilterAnalysis.OpenFiltered)

	fmt.Println()
	fmt.Println("Scan Types")

	if config.TCP {
		fmt.Println("  TCP: enabled")
	}

	if config.UDP {
		fmt.Println("  UDP: enabled")
	}

	if config.ServiceDetection {
		fmt.Println("  Service Detection: enabled")
	}

	if config.VersionDetection {
		fmt.Println("  Version Detection: enabled")
	}

	if config.OSDetection {
		fmt.Println("  OS Detection: enabled")
	}

	fmt.Println()

	if config.OSDetection {
		printOSSummary(result.OS)
	}

	if config.ServiceDetection ||
		config.VersionDetection {
		printServiceSummary(result.Ports)
	}

	if len(result.FilterAnalysis.Observations) > 0 {
		fmt.Println()
		fmt.Println("Observations")
		fmt.Println("------------")

		for _, observation := range result.FilterAnalysis.Observations {
			fmt.Printf("  - %s\n", observation)
		}
	}
}

func printOSSummary(info OSInfo) {
	fmt.Println("OS Detection")
	fmt.Println("------------")

	if info.Name == "" {
		fmt.Println("  OS: Unknown")
		return
	}

	fmt.Printf("  OS:         %s\n", info.Name)
	fmt.Printf("  Family:     %s\n", info.Family)
	fmt.Printf("  Version:    %s\n", info.Version)
	fmt.Printf("  Confidence: %d%%\n", info.Confidence)

	if len(info.Evidence) > 0 {
		fmt.Println("  Evidence:")

		for _, evidence := range info.Evidence {
			fmt.Printf("    - %s\n", evidence)
		}
	}
}

func printServiceSummary(results []PortResult) {
	fmt.Println("Service Detection")
	fmt.Println("------------------")

	found := false

	for _, result := range results {
		if result.State != "open" {
			continue
		}

		if result.Service == "" &&
			result.Product == "" &&
			result.Version == "" {
			continue
		}

		found = true

		service := result.Service
		product := result.Product
		version := result.Version

		fmt.Printf(
			"  %d/%s  %-12s",
			result.Port,
			result.Protocol,
			service,
		)

		if product != "" {
			fmt.Printf(" %-18s", product)
		}

		if version != "" {
			fmt.Printf(" %s", version)
		}

		if result.Confidence > 0 {
			fmt.Printf(
				" [%d%%]",
				result.Confidence,
			)
		}

		fmt.Println()
	}

	if !found {
		fmt.Println("  No service information detected")
	}
}

func formatScanTypes(config ScanConfig) string {
	var types []string

	if config.TCP {
		types = append(types, "TCP")
	}

	if config.UDP {
		types = append(types, "UDP")
	}

	return strings.Join(types, ", ")
}
