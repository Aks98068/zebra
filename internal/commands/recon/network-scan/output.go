package networkscan

import (
	"fmt"
	"sort"
	"time"
)

func ClassifyPortResults(
	tcpResults []TCPResult,
	udpResults []UDPResult,
) []PortResult {
	results := make([]PortResult, 0,
		len(tcpResults)+len(udpResults),
	)

	for _, result := range tcpResults {
		results = append(results, PortResult{
			Port:     result.Port,
			Protocol: "tcp",
			State:    result.State,
			Reason:   result.Reason,
		})
	}

	for _, result := range udpResults {
		results = append(results, PortResult{
			Port:     result.Port,
			Protocol: "udp",
			State:    result.State,
			Reason:   result.Reason,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Port == results[j].Port {
			return results[i].Protocol < results[j].Protocol
		}

		return results[i].Port < results[j].Port
	})

	return results
}

func BuildScanResult(
	config ScanConfig,
	tcpResults []TCPResult,
	udpResults []UDPResult,
	startedAt time.Time,

) ScanResult {
	ports := ClassifyPortResults(
		tcpResults,
		udpResults,
	)

	openPorts := 0

	for _, port := range ports {
		if port.State == "open" {
			openPorts++
		}
	}

	filterAnalysis := AnalyzeFiltering(ports)

	return ScanResult{
		Target:         config.Target,
		StartedAt:      startedAt,
		Duration:       time.Since(startedAt),
		Ports:          ports,
		TotalPorts:     len(ports),
		OpenPorts:      openPorts,
		FilterAnalysis: filterAnalysis,
	}
}

func PrintPortResults(
	results []PortResult,
	showReason bool,
	openOnly bool,
) {
	for _, result := range results {
		if openOnly && result.State != "open" {
			continue
		}

		if showReason {
			fmt.Printf(
				"%5d/%-3s %-14s %s\n",
				result.Port,
				result.Protocol,
				result.State,
				result.Reason,
			)
		} else {
			fmt.Printf(
				"%5d/%-3s %-14s\n",
				result.Port,
				result.Protocol,
				result.State,
			)
		}
	}
}
