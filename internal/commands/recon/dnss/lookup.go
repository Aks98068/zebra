package dnss

import (
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"strings"

	"zebra/internal/commands/output"
)

func DNSLookup(args []string) bool {

	if len(args) == 0 {
		fmt.Println("usage: dns <domain> [-type <record>] [-o <file>]")
		fmt.Println("       dns reverse <ip> [-o <file>]")
		return false
	}

	// =========================================================
	// Parse arguments
	// =========================================================

	operation := args[0]

	var recordType string
	var outputFile string

	for i := 1; i < len(args); i++ {

		switch args[i] {

		case "-type":

			if i+1 >= len(args) {
				fmt.Println("missing value for -type")
				return false
			}

			recordType = strings.ToUpper(args[i+1])
			i++

		case "-o":

			if i+1 >= len(args) {
				fmt.Println("missing filename for -o")
				return false
			}

			outputFile = args[i+1]
			i++

		default:

			fmt.Println("unknown option:", args[i])
			return false
		}
	}

	// =========================================================
	// Reverse DNS
	// =========================================================

	if operation == "reverse" {

		if len(args) < 2 {
			fmt.Println("provide an IP address")
			return false
		}

		ip := args[1]

		names, err := net.LookupAddr(ip)

		if err != nil {
			fmt.Println("Reverse DNS failed for:", ip, err)
			return false
		}

		fmt.Println("Reverse DNS results for:", ip)

		var content strings.Builder

		content.WriteString("Reverse DNS results for: ")
		content.WriteString(ip)
		content.WriteString("\n\n")

		for _, name := range names {
			fmt.Println(name)
			content.WriteString(name)
			content.WriteString("\n")
		}

		// Save only when -o is provided.
		if outputFile != "" {

			if strings.EqualFold(filepath.Ext(outputFile), ".json") {

				data := map[string]any{
					"type":      "reverse_dns",
					"target":    ip,
					"hostnames": names,
				}

				if err := output.SaveJSON(outputFile, data); err != nil {
					fmt.Println("failed to save JSON:", err)
					return false
				}

			} else {

				if err := output.SaveText(outputFile, content.String()); err != nil {
					fmt.Println("failed to save file:", err)
					return false
				}
			}

			fmt.Println("Output saved to:", outputFile)
		}

		return false
	}

	// =========================================================
	// Normal DNS lookup
	// =========================================================

	domain := operation

	// Default = A record
	if recordType == "" {
		recordType = "A"
	}

	// =========================================================
	// Perform lookup
	// =========================================================

	switch recordType {

	case "A":

		return ALookup(domain, outputFile)

	case "AAAA":

		return AAAALookup(domain, outputFile)

	case "MX":

		return MXLookup(domain, outputFile)

	case "NS":

		return NSLookup(domain, outputFile)

	case "TXT":

		return TXTLookup(domain, outputFile)

	case "CNAME":

		return CNAMELookup(domain, outputFile)

	case "ALL":

		return ALLLookup(domain, outputFile)

	default:

		fmt.Println("Unsupported DNS record type:", recordType)
		return false
	}
}

// =============================================================
// A RECORD
// =============================================================

func ALookup(domain string, outputFile string) bool {

	addresses, err := net.LookupHost(domain)

	if err != nil {
		fmt.Println("A lookup failed:", err)
		return false
	}

	var ipv4 []string
	var content strings.Builder

	content.WriteString("A records for: ")
	content.WriteString(domain)
	content.WriteString("\n\n")

	fmt.Println("A records for:", domain)

	for _, address := range addresses {

		ip := net.ParseIP(address)

		if ip != nil && ip.To4() != nil {

			ipv4 = append(ipv4, address)

			fmt.Println(address)

			content.WriteString(address)
			content.WriteString("\n")
		}
	}

	if outputFile != "" {

		data := map[string]any{
			"type":      "A",
			"target":    domain,
			"addresses": ipv4,
		}

		if err := saveDNSResult(outputFile, data, content.String()); err != nil {
			fmt.Println("failed to save result:", err)
			return false
		}

		fmt.Println("Output saved to:", outputFile)
	}

	return false
}

// =============================================================
// AAAA RECORD
// =============================================================

func AAAALookup(domain string, outputFile string) bool {

	addresses, err := net.LookupHost(domain)

	if err != nil {
		fmt.Println("AAAA lookup failed:", err)
		return false
	}

	var ipv6 []string
	var content strings.Builder

	content.WriteString("AAAA records for: ")
	content.WriteString(domain)
	content.WriteString("\n\n")

	fmt.Println("AAAA records for:", domain)

	for _, address := range addresses {

		ip := net.ParseIP(address)

		if ip != nil && ip.To4() == nil {

			ipv6 = append(ipv6, address)

			fmt.Println(address)

			content.WriteString(address)
			content.WriteString("\n")
		}
	}

	if outputFile != "" {

		data := map[string]any{
			"type":      "AAAA",
			"target":    domain,
			"addresses": ipv6,
		}

		if err := saveDNSResult(outputFile, data, content.String()); err != nil {
			fmt.Println("failed to save result:", err)
			return false
		}

		fmt.Println("Output saved to:", outputFile)
	}

	return false
}

// =============================================================
// MX RECORD
// =============================================================

func MXLookup(domain string, outputFile string) bool {

	records, err := net.LookupMX(domain)

	if err != nil {
		fmt.Println("MX lookup failed:", err)
		return false
	}

	fmt.Println("MX records for:", domain)

	var content strings.Builder

	content.WriteString("MX records for: ")
	content.WriteString(domain)
	content.WriteString("\n\n")

	var results []map[string]any

	for _, record := range records {

		fmt.Println("Host:", record.Host)
		fmt.Println("Priority:", record.Pref)

		content.WriteString("Host: ")
		content.WriteString(record.Host)
		content.WriteString("\n")

		content.WriteString("Priority: ")
		content.WriteString(strconv.Itoa(int(record.Pref)))
		content.WriteString("\n\n")

		results = append(results, map[string]any{
			"host":     record.Host,
			"priority": record.Pref,
		})
	}

	if outputFile != "" {

		data := map[string]any{
			"type":      "MX",
			"target":    domain,
			"records":   results,
		}

		if err := saveDNSResult(outputFile, data, content.String()); err != nil {
			fmt.Println("failed to save result:", err)
			return false
		}

		fmt.Println("Output saved to:", outputFile)
	}

	return false
}

// =============================================================
// NS RECORD
// =============================================================

func NSLookup(domain string, outputFile string) bool {

	records, err := net.LookupNS(domain)

	if err != nil {
		fmt.Println("NS lookup failed:", err)
		return false
	}

	fmt.Println("NS records for:", domain)

	var content strings.Builder
	var nameservers []string

	content.WriteString("NS records for: ")
	content.WriteString(domain)
	content.WriteString("\n\n")

	for _, record := range records {

		fmt.Println(record.Host)

		nameservers = append(nameservers, record.Host)

		content.WriteString(record.Host)
		content.WriteString("\n")
	}

	if outputFile != "" {

		data := map[string]any{
			"type":        "NS",
			"target":      domain,
			"nameservers": nameservers,
		}

		if err := saveDNSResult(outputFile, data, content.String()); err != nil {
			fmt.Println("failed to save result:", err)
			return false
		}

		fmt.Println("Output saved to:", outputFile)
	}

	return false
}

// =============================================================
// TXT RECORD
// =============================================================

func TXTLookup(domain string, outputFile string) bool {

	records, err := net.LookupTXT(domain)

	if err != nil {
		fmt.Println("TXT lookup failed:", err)
		return false
	}

	fmt.Println("TXT records for:", domain)

	var content strings.Builder

	content.WriteString("TXT records for: ")
	content.WriteString(domain)
	content.WriteString("\n\n")

	for _, record := range records {

		fmt.Println(record)

		content.WriteString(record)
		content.WriteString("\n")
	}

	if outputFile != "" {

		data := map[string]any{
			"type":    "TXT",
			"target":  domain,
			"records": records,
		}

		if err := saveDNSResult(outputFile, data, content.String()); err != nil {
			fmt.Println("failed to save result:", err)
			return false
		}

		fmt.Println("Output saved to:", outputFile)
	}

	return false
}

// =============================================================
// CNAME RECORD
// =============================================================

func CNAMELookup(domain string, outputFile string) bool {

	record, err := net.LookupCNAME(domain)

	if err != nil {
		fmt.Println("CNAME lookup failed:", err)
		return false
	}

	fmt.Println("CNAME record for:", domain)
	fmt.Println(record)

	content := "CNAME record for: " + domain + "\n\n" + record + "\n"

	if outputFile != "" {

		data := map[string]any{
			"type":   "CNAME",
			"target": domain,
			"record": record,
		}

		if err := saveDNSResult(outputFile, data, content); err != nil {
			fmt.Println("failed to save result:", err)
			return false
		}

		fmt.Println("Output saved to:", outputFile)
	}

	return false
}

// =============================================================
// ALL RECORDS
// =============================================================

func ALLLookup(domain string, outputFile string) bool {

	fmt.Println("DNS records for:", domain)
	fmt.Println()

	// For now, terminal output is produced by each lookup.
	ALookup(domain, "")
	AAAALookup(domain, "")
	MXLookup(domain, "")
	NSLookup(domain, "")
	TXTLookup(domain, "")
	CNAMELookup(domain, "")

	// If -o is supplied, create a combined result.
	if outputFile != "" {

		var content strings.Builder

		content.WriteString("DNS records for: ")
		content.WriteString(domain)
		content.WriteString("\n\n")

		content.WriteString("=== A ===\n")

		addresses, _ := net.LookupHost(domain)

		for _, address := range addresses {
			ip := net.ParseIP(address)

			if ip != nil && ip.To4() != nil {
				content.WriteString(address)
				content.WriteString("\n")
			}
		}

		content.WriteString("\n=== AAAA ===\n")

		for _, address := range addresses {
			ip := net.ParseIP(address)

			if ip != nil && ip.To4() == nil {
				content.WriteString(address)
				content.WriteString("\n")
			}
		}

		content.WriteString("\n=== MX ===\n")

		mxRecords, _ := net.LookupMX(domain)

		for _, record := range mxRecords {
			content.WriteString("Host: ")
			content.WriteString(record.Host)
			content.WriteString("\n")

			content.WriteString("Priority: ")
			content.WriteString(strconv.Itoa(int(record.Pref)))
			content.WriteString("\n")
		}

		content.WriteString("\n=== NS ===\n")

		nsRecords, _ := net.LookupNS(domain)

		for _, record := range nsRecords {
			content.WriteString(record.Host)
			content.WriteString("\n")
		}

		content.WriteString("\n=== TXT ===\n")

		txtRecords, _ := net.LookupTXT(domain)

		for _, record := range txtRecords {
			content.WriteString(record)
			content.WriteString("\n")
		}

		content.WriteString("\n=== CNAME ===\n")

		cname, _ := net.LookupCNAME(domain)
		content.WriteString(cname)
		content.WriteString("\n")

		if strings.EqualFold(filepath.Ext(outputFile), ".json") {

			data := map[string]any{
				"type":   "ALL",
				"target": domain,
			}

			if err := output.SaveJSON(outputFile, data); err != nil {
				fmt.Println("failed to save JSON:", err)
				return false
			}

		} else {

			if err := output.SaveText(outputFile, content.String()); err != nil {
				fmt.Println("failed to save file:", err)
				return false
			}
		}

		fmt.Println("Output saved to:", outputFile)
	}

	return false
}

// =============================================================
// SAVE DNS RESULT
// =============================================================

func saveDNSResult(filename string, data any, text string) error {

	if strings.EqualFold(filepath.Ext(filename), ".json") {
		return output.SaveJSON(filename, data)
	}

	return output.SaveText(filename, text)
}