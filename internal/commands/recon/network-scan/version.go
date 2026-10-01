package networkscan

import (
	"regexp"
	"strconv"
	"strings"
)

type VersionInfo struct {
	Product    string
	Version    string
	Confidence int
}

func DetectVersions(
	config ScanConfig,
	results []PortResult,
) []PortResult {
	if !config.VersionDetection {
		return results
	}

	for i := range results {
		if results[i].State != "open" {
			continue
		}

		info := DetectServiceVersion(results[i])

		results[i].Product = info.Product
		results[i].Version = info.Version
		results[i].Confidence = info.Confidence
	}

	return results
}

func DetectServiceVersion(result PortResult) VersionInfo {
	banner := strings.TrimSpace(result.Banner)

	if banner == "" {
		return VersionInfo{
			Product:    result.Service,
			Confidence: 30,
		}
	}

	switch strings.ToLower(result.Service) {
	case "ssh":
		return detectSSHVersion(banner)

	case "ftp":
		return detectFTPVersion(banner)

	case "smtp":
		return detectSMTPVersion(banner)

	case "http", "https", "http-proxy":
		return detectHTTPVersion(banner)

	case "mysql":
		return detectMySQLVersion(banner)

	case "redis":
		return detectRedisVersion(banner)

	case "postgresql":
		return detectPostgreSQLVersion(banner)

	default:
		return detectGenericVersion(
			result.Service,
			banner,
		)
	}
}

func detectSSHVersion(banner string) VersionInfo {
	re := regexp.MustCompile(
		`(?i)^SSH-[0-9.]+-([^\s]+)`,
	)

	match := re.FindStringSubmatch(
		strings.TrimSpace(banner),
	)

	if len(match) < 2 {
		return VersionInfo{
			Product:    "SSH",
			Confidence: 60,
		}
	}

	product, version := splitProductVersion(match[1])

	return VersionInfo{
		Product:    product,
		Version:    version,
		Confidence: 95,
	}
}

func detectFTPVersion(banner string) VersionInfo {
	value := strings.TrimSpace(banner)

	product, version := extractProductVersion(value)

	return VersionInfo{
		Product:    product,
		Version:    version,
		Confidence: versionConfidence(version, 85),
	}
}

func detectSMTPVersion(banner string) VersionInfo {
	value := strings.TrimSpace(banner)

	product, version := extractProductVersion(value)

	return VersionInfo{
		Product:    product,
		Version:    version,
		Confidence: versionConfidence(version, 80),
	}
}

func detectHTTPVersion(banner string) VersionInfo {
	server := extractHeader(
		banner,
		"Server:",
	)

	if server == "" {
		return VersionInfo{
			Product:    "HTTP",
			Confidence: 60,
		}
	}

	product, version := extractProductVersion(server)

	if product == "" {
		product = "HTTP"
	}

	return VersionInfo{
		Product:    product,
		Version:    version,
		Confidence: versionConfidence(version, 90),
	}
}

func detectMySQLVersion(banner string) VersionInfo {
	product, version := extractProductVersion(
		banner,
	)

	if product == "" {
		product = "MySQL"
	}

	return VersionInfo{
		Product:    product,
		Version:    version,
		Confidence: versionConfidence(version, 85),
	}
}

func detectRedisVersion(banner string) VersionInfo {
	version := extractVersion(
		banner,
	)

	return VersionInfo{
		Product:    "Redis",
		Version:    version,
		Confidence: versionConfidence(version, 85),
	}
}

func detectPostgreSQLVersion(banner string) VersionInfo {
	product, version := extractProductVersion(
		banner,
	)

	if product == "" {
		product = "PostgreSQL"
	}

	return VersionInfo{
		Product:    product,
		Version:    version,
		Confidence: versionConfidence(version, 85),
	}
}

func detectGenericVersion(
	service string,
	banner string,
) VersionInfo {
	product, version := extractProductVersion(
		banner,
	)

	if product == "" {
		product = service
	}

	return VersionInfo{
		Product:    product,
		Version:    version,
		Confidence: versionConfidence(version, 50),
	}
}

func extractHeader(
	banner string,
	header string,
) string {
	lines := strings.Split(banner, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(
			strings.ToLower(line),
			strings.ToLower(header),
		) {
			return strings.TrimSpace(
				line[len(header):],
			)
		}
	}

	return ""
}

func extractProductVersion(
	value string,
) (string, string) {
	value = strings.TrimSpace(value)

	if value == "" {
		return "", ""
	}

	re := regexp.MustCompile(
		`(?i)([A-Za-z][A-Za-z0-9._-]*)[/ ]([0-9]+(?:\.[0-9]+){0,3})`,
	)

	match := re.FindStringSubmatch(value)

	if len(match) >= 3 {
		return match[1], match[2]
	}

	return value, ""
}

func splitProductVersion(
	value string,
) (string, string) {
	value = strings.TrimSpace(value)

	parts := strings.SplitN(
		value,
		"_",
		2,
	)

	if len(parts) == 2 {
		return parts[0], parts[1]
	}

	parts = strings.SplitN(
		value,
		"-",
		2,
	)

	if len(parts) == 2 {
		if isVersion(parts[1]) {
			return parts[0], parts[1]
		}
	}

	return value, ""
}

func extractVersion(value string) string {
	re := regexp.MustCompile(
		`[0-9]+(?:\.[0-9]+){1,3}`,
	)

	return re.FindString(value)
}

func isVersion(value string) bool {
	_, err := strconv.Atoi(
		strings.Split(value, ".")[0],
	)

	return err == nil
}

func versionConfidence(
	version string,
	base int,
) int {
	if version == "" {
		return base - 30
	}

	if base > 100 {
		return 100
	}

	return base
}
