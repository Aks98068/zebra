package networkscan

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

func DetectServices(config ScanConfig, results []PortResult) []PortResult {
	if !config.ServiceDetection {
		return results
	}

	for i := range results {
		if results[i].State != "open" {
			continue
		}

		switch results[i].Protocol {
		case "tcp":
			service, banner := detectTCPService(
				config.Target,
				results[i].Port,
				config.Timeout,
			)

			results[i].Service = service
			results[i].Banner = banner

		case "udp":
			results[i].Service = detectUDPService(
				results[i].Port,
			)
		}

		if results[i].Service == "" {
			results[i].Service = serviceByPort(
				results[i].Port,
				results[i].Protocol,
			)
		}
	}

	return results
}

func detectTCPService(
	target string,
	port int,
	timeout time.Duration,
) (string, string) {
	switch port {
	case 80, 8080, 8000, 8008, 8888:
		return probeHTTP(target, port, timeout)

	case 443, 8443:
		return probeHTTPS(target, port, timeout)

	case 22:
		return probeSSH(target, port, timeout)

	case 21:
		return probeFTP(target, port, timeout)

	case 25, 465, 587:
		return probeSMTP(target, port, timeout)

	default:
		return probeTCPBanner(target, port, timeout)
	}
}

func detectUDPService(port int) string {
	switch port {
	case 53:
		return "dns"
	case 67, 68:
		return "dhcp"
	case 69:
		return "tftp"
	case 123:
		return "ntp"
	case 161, 162:
		return "snmp"
	case 500:
		return "ike"
	case 514:
		return "syslog"
	case 1900:
		return "ssdp"
	default:
		return ""
	}
}

func probeHTTP(
	target string,
	port int,
	timeout time.Duration,
) (string, string) {
	return "http", ""
}

func probeHTTPS(
	target string,
	port int,
	timeout time.Duration,
) (string, string) {
	return "https", ""
}

func probeSSH(
	target string,
	port int,
	timeout time.Duration,
) (string, string) {
	return "ssh", ""
}

func probeFTP(
	target string,
	port int,
	timeout time.Duration,
) (string, string) {
	return "ftp", ""
}

func probeSMTP(
	target string,
	port int,
	timeout time.Duration,
) (string, string) {
	return "smtp", ""
}

func probeTCPBanner(
	target string,
	port int,
	timeout time.Duration,
) (string, string) {
	address := net.JoinHostPort(
		target,
		strconv.Itoa(port),
	)

	conn, err := net.DialTimeout(
		"tcp",
		address,
		timeout,
	)
	if err != nil {
		return "", ""
	}

	defer conn.Close()

	_ = conn.SetReadDeadline(
		time.Now().Add(timeout),
	)

	reader := bufio.NewReader(conn)

	data := make([]byte, 1024)

	n, err := reader.Read(data)
	if err != nil || n == 0 {
		return "", ""
	}

	banner := cleanBanner(string(data[:n]))

	return identifyBannerService(banner), banner
}

func identifyBannerService(banner string) string {
	value := strings.ToLower(banner)

	switch {
	case strings.Contains(value, "ssh-"):
		return "ssh"

	case strings.Contains(value, "ftp"):
		return "ftp"

	case strings.Contains(value, "smtp"):
		return "smtp"

	case strings.Contains(value, "mysql"):
		return "mysql"

	case strings.Contains(value, "redis"):
		return "redis"

	case strings.Contains(value, "postgres"):
		return "postgresql"

	default:
		return ""
	}
}

func serviceByPort(port int, protocol string) string {
	if protocol == "udp" {
		return detectUDPService(port)
	}

	switch port {
	case 21:
		return "ftp"
	case 22:
		return "ssh"
	case 23:
		return "telnet"
	case 25:
		return "smtp"
	case 53:
		return "dns"
	case 80:
		return "http"
	case 110:
		return "pop3"
	case 143:
		return "imap"
	case 443:
		return "https"
	case 445:
		return "smb"
	case 587:
		return "smtp"
	case 993:
		return "imaps"
	case 995:
		return "pop3s"
	case 1433:
		return "mssql"
	case 1521:
		return "oracle"
	case 3306:
		return "mysql"
	case 3389:
		return "rdp"
	case 5432:
		return "postgresql"
	case 5900:
		return "vnc"
	case 6379:
		return "redis"
	case 8080:
		return "http-proxy"
	case 8443:
		return "https"
	case 9200:
		return "elasticsearch"
	case 27017:
		return "mongodb"
	default:
		return ""
	}
}

func cleanBanner(banner string) string {
	banner = strings.TrimSpace(banner)

	banner = strings.ReplaceAll(
		banner,
		"\r",
		"",
	)

	banner = strings.ReplaceAll(
		banner,
		"\n",
		" ",
	)

	if len(banner) > 512 {
		banner = banner[:512]
	}

	return strings.TrimSpace(banner)
}

func formatServiceResult(result PortResult) string {
	if result.Service == "" {
		return fmt.Sprintf(
			"%d/%s",
			result.Port,
			result.Protocol,
		)
	}

	return fmt.Sprintf(
		"%d/%s %s",
		result.Port,
		result.Protocol,
		result.Service,
	)
}
