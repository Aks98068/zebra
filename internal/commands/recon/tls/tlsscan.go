package tls

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type CertificateInfo struct {
	Subject            string    `json:"subject"`
	Issuer             string    `json:"issuer"`
	SerialNumber       string    `json:"serial_number"`
	NotBefore          time.Time `json:"not_before"`
	NotAfter           time.Time `json:"not_after"`
	DNSNames           []string  `json:"dns_names"`
	IPAddresses        []string  `json:"ip_addresses"`
	SignatureAlgorithm string    `json:"signature_algorithm"`
	PublicKeyAlgorithm string    `json:"public_key_algorithm"`
}

type TLSResult struct {
	Target           string           `json:"target"`
	Port             int              `json:"port"`
	ScannedAt        time.Time        `json:"scanned_at"`
	TLSVersion       string           `json:"tls_version"`
	CipherSuite      string           `json:"cipher_suite"`
	ServerName       string           `json:"server_name"`
	Certificate      *CertificateInfo `json:"certificate"`
	HostnameValid    bool             `json:"hostname_valid"`
	CertificateValid bool             `json:"certificate_valid"`
	Findings         []string         `json:"findings"`
}

func TLSRecon(args []string) bool {
	if len(args) == 0 {
		fmt.Println("tls <host> [port]")
		return false
	}

	host := args[0]
	port := 443

	if len(args) >= 2 {
		parsedPort, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println("Error: invalid port:", args[1])
			return false
		}

		port = parsedPort
	}

	if err := ValidateTLSINPUT(host, port); err != nil {
		fmt.Println("Error:", err)
		return false
	}

	conn, err := ConnectTLS(host, port)
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}
	defer conn.Close()

	certificate, err := InspectCertificate(conn)
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}

	state := conn.ConnectionState()

	result := TLSResult{
		Target:           host,
		Port:             port,
		ScannedAt:        time.Now().UTC(),
		TLSVersion:       tlsVersionName(state.Version),
		CipherSuite:      tls.CipherSuiteName(state.CipherSuite),
		ServerName:       host,
		Certificate:      certificate,
		HostnameValid:    true,
		CertificateValid: true,
		Findings:         []string{},
	}

	if certificate != nil {
		now := time.Now()

		if now.Before(certificate.NotBefore) {
			result.CertificateValid = false
			result.Findings = append(
				result.Findings,
				"certificate is not yet valid",
			)
		}

		if now.After(certificate.NotAfter) {
			result.CertificateValid = false
			result.Findings = append(
				result.Findings,
				"certificate has expired",
			)
		}
	}

	if err := SaveTLSReport(result); err != nil {
		fmt.Println("Error saving TLS report:", err)
		return false
	}

	PrintTLSResult(result)

	return true
}

func ValidateTLSINPUT(host string, port int) error {
	host = strings.TrimSpace(host)

	if host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	if strings.IndexFunc(host, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsControl(r)
	}) != -1 {
		return fmt.Errorf("host contains invalid whitespace or control characters")
	}

	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid port: %d", port)
	}

	return nil
}

func ConnectTLS(host string, port int) (*tls.Conn, error) {
	address := net.JoinHostPort(host, strconv.Itoa(port))

	config := &tls.Config{
		ServerName: host,
		MinVersion: tls.VersionTLS12,
	}

	dialer := &net.Dialer{
		Timeout: 5 * time.Second,
	}

	conn, err := tls.DialWithDialer(
		dialer,
		"tcp",
		address,
		config,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"TLS connection to %s failed: %w",
			address,
			err,
		)
	}

	return conn, nil
}

func InspectCertificate(conn *tls.Conn) (*CertificateInfo, error) {
	if conn == nil {
		return nil, fmt.Errorf("TLS connection is nil")
	}

	state := conn.ConnectionState()

	if len(state.PeerCertificates) == 0 {
		return nil, fmt.Errorf("server did not provide a certificate")
	}

	cert := state.PeerCertificates[0]

	ipAddresses := make([]string, 0, len(cert.IPAddresses))

	for _, ip := range cert.IPAddresses {
		ipAddresses = append(ipAddresses, ip.String())
	}

	dnsNames := append([]string(nil), cert.DNSNames...)

	info := &CertificateInfo{
		Subject:            cert.Subject.String(),
		Issuer:             cert.Issuer.String(),
		SerialNumber:       cert.SerialNumber.String(),
		NotBefore:          cert.NotBefore,
		NotAfter:           cert.NotAfter,
		DNSNames:           dnsNames,
		IPAddresses:        ipAddresses,
		SignatureAlgorithm: cert.SignatureAlgorithm.String(),
		PublicKeyAlgorithm: cert.PublicKeyAlgorithm.String(),
	}

	return info, nil
}

func SaveTLSReport(result TLSResult) error {
	reportDir := filepath.Join("reports", "tls")

	if err := os.MkdirAll(reportDir, 0750); err != nil {
		return fmt.Errorf("create TLS report directory: %w", err)
	}

	targetName := safeReportName(result.Target)

	timestamp := result.ScannedAt.UTC().Format("20060102_150405")

	filename := fmt.Sprintf(
		"%s_%d_%s.json",
		targetName,
		result.Port,
		timestamp,
	)

	reportPath := filepath.Join(reportDir, filename)

	data, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		return fmt.Errorf("encode TLS report: %w", err)
	}

	data = append(data, '\n')

	file, err := os.OpenFile(
		reportPath,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0600,
	)
	if err != nil {
		return fmt.Errorf("create TLS report: %w", err)
	}

	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("write TLS report: %w", err)
	}

	fmt.Println("TLS report saved:", reportPath)

	return nil
}

func safeReportName(value string) string {
	value = strings.TrimSpace(value)

	if value == "" {
		return "unknown"
	}

	var builder strings.Builder

	for _, r := range value {
		switch {
		case unicode.IsLetter(r):
			builder.WriteRune(r)

		case unicode.IsDigit(r):
			builder.WriteRune(r)

		case r == '.' || r == '-' || r == '_':
			builder.WriteRune(r)

		default:
			builder.WriteRune('_')
		}
	}

	name := builder.String()

	if name == "" {
		return "unknown"
	}

	return name
}

func tlsVersionName(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"

	case tls.VersionTLS11:
		return "TLS 1.1"

	case tls.VersionTLS12:
		return "TLS 1.2"

	case tls.VersionTLS13:
		return "TLS 1.3"

	default:
		return fmt.Sprintf("Unknown (0x%04x)", version)
	}
}

func PrintTLSResult(result TLSResult) {
	fmt.Println()
	fmt.Println("========== TLS RECON ==========")
	fmt.Println("Target:       ", result.Target)
	fmt.Println("Port:         ", result.Port)
	fmt.Println("TLS Version:  ", result.TLSVersion)
	fmt.Println("Cipher Suite: ", result.CipherSuite)
	fmt.Println("Server Name:  ", result.ServerName)

	fmt.Println()

	fmt.Println("Certificate:")

	if result.Certificate == nil {
		fmt.Println("  No certificate information")
	} else {
		fmt.Println("  Subject:     ", result.Certificate.Subject)
		fmt.Println("  Issuer:      ", result.Certificate.Issuer)
		fmt.Println("  Serial:      ", result.Certificate.SerialNumber)
		fmt.Println("  Valid From:  ", result.Certificate.NotBefore)
		fmt.Println("  Valid Until: ", result.Certificate.NotAfter)
		fmt.Println("  DNS Names:   ", strings.Join(result.Certificate.DNSNames, ", "))
		fmt.Println("  Signature:   ", result.Certificate.SignatureAlgorithm)
		fmt.Println("  Public Key:  ", result.Certificate.PublicKeyAlgorithm)
	}

	fmt.Println()

	if len(result.Findings) == 0 {
		fmt.Println("Findings:      none")
	} else {
		fmt.Println("Findings:")

		for _, finding := range result.Findings {
			fmt.Println("  -", finding)
		}
	}

	fmt.Println("===============================")
	fmt.Println()
}
