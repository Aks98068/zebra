package networkscan

import (
	"fmt"
	"net"
	"strings"
	"time"
)

type OSInfo struct {
	Name       string
	Family     string
	Version    string
	Confidence int
	Evidence   []string
}

func DetectOS(
	config ScanConfig,
	results []PortResult,
) OSInfo {
	if !config.OSDetection {
		return OSInfo{}
	}

	info := OSInfo{
		Name:       "Unknown",
		Family:     "Unknown",
		Confidence: 0,
		Evidence:   []string{},
	}

	ttl, err := probeTTL(config.Target, config.Timeout)

	if err == nil {
		info.Evidence = append(
			info.Evidence,
			fmt.Sprintf("observed TTL=%d", ttl),
		)

		family, confidence := classifyTTL(ttl)

		if family != "" {
			info.Family = family
			info.Confidence = confidence
		}
	}

	tcpEvidence := analyzeTCPResults(results)

	info.Evidence = append(
		info.Evidence,
		tcpEvidence...,
	)

	info.Name = inferOSName(
		info.Family,
		results,
	)

	if info.Name != "Unknown" && info.Confidence < 50 {
		info.Confidence = 50
	}

	if info.Confidence > 100 {
		info.Confidence = 100
	}

	return info
}

func probeTTL(
	target string,
	timeout time.Duration,
) (int, error) {
	address := net.JoinHostPort(
		target,
		"80",
	)

	conn, err := net.DialTimeout(
		"tcp",
		address,
		timeout,
	)

	if err != nil {
		return 0, err
	}

	defer conn.Close()

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		_ = tcpConn.SetKeepAlive(false)
	}

	return 0, fmt.Errorf(
		"TTL unavailable through standard Go TCP API",
	)
}

func classifyTTL(ttl int) (string, int) {
	switch {
	case ttl <= 64:
		return "Unix/Linux", 55

	case ttl <= 128:
		return "Windows", 55

	case ttl <= 255:
		return "Network device/Unix", 40

	default:
		return "", 0
	}
}

func analyzeTCPResults(
	results []PortResult,
) []string {
	evidence := make([]string, 0)

	openTCP := 0
	closedTCP := 0
	filteredTCP := 0

	for _, result := range results {
		if result.Protocol != "tcp" {
			continue
		}

		switch strings.ToLower(result.State) {
		case "open":
			openTCP++

		case "closed":
			closedTCP++

		case "filtered":
			filteredTCP++
		}
	}

	if openTCP > 0 {
		evidence = append(
			evidence,
			fmt.Sprintf(
				"TCP open ports=%d",
				openTCP,
			),
		)
	}

	if closedTCP > 0 {
		evidence = append(
			evidence,
			fmt.Sprintf(
				"TCP closed ports=%d",
				closedTCP,
			),
		)
	}

	if filteredTCP > 0 {
		evidence = append(
			evidence,
			fmt.Sprintf(
				"TCP filtered ports=%d",
				filteredTCP,
			),
		)
	}

	return evidence
}

func inferOSName(
	family string,
	results []PortResult,
) string {
	switch family {
	case "Windows":
		return "Windows"

	case "Unix/Linux":
		return "Linux/Unix"

	case "Network device/Unix":
		if hasService(results, "ssh") {
			return "Unix/Linux or network device"
		}

		return "Network device/Unix"

	default:
		return "Unknown"
	}
}

func hasService(
	results []PortResult,
	service string,
) bool {
	for _, result := range results {
		if strings.EqualFold(
			result.Service,
			service,
		) {
			return true
		}
	}

	return false
}
