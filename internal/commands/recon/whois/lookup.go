package whois

import (
	"fmt"
	"io"
	"net"
	"path/filepath"
	"strings"
	"time"

	"zebra/internal/commands/output"
)

func WHOISLookup(args []string) bool {

	// =========================================================
	// ARGUMENT VALIDATION
	// =========================================================

	if len(args) == 0 {
		fmt.Println("usage: whois <domain|ip> [-server <server>] [-timeout <duration>] [-no-follow] [-o <file>]")
		return false
	}

	target := strings.TrimSpace(args[0])

	if target == "" {
		fmt.Println("target cannot be empty")
		return false
	}

	// =========================================================
	// DOMAIN / IP DETECTION
	// =========================================================

	ip := net.ParseIP(target)

	targetType := "domain"

	if ip != nil {
		targetType = "ip"
	} else {

		// Remove a trailing dot from a fully-qualified domain.
		target = strings.TrimSuffix(target, ".")

		// Basic domain validation.
		if len(target) > 253 {
			fmt.Println("invalid domain: domain is too long")
			return false
		}

		if strings.Contains(target, "..") {
			fmt.Println("invalid domain: consecutive dots are not allowed")
			return false
		}

		labels := strings.Split(target, ".")

		if len(labels) < 2 {
			fmt.Println("invalid domain:", target)
			return false
		}

		for _, label := range labels {

			if label == "" {
				fmt.Println("invalid domain:", target)
				return false
			}

			if len(label) > 63 {
				fmt.Println("invalid domain label:", label)
				return false
			}

			if strings.HasPrefix(label, "-") ||
				strings.HasSuffix(label, "-") {

				fmt.Println("invalid domain label:", label)
				return false
			}

			for _, r := range label {

				if (r >= 'a' && r <= 'z') ||
					(r >= 'A' && r <= 'Z') ||
					(r >= '0' && r <= '9') ||
					r == '-' {

					continue
				}

				fmt.Println("invalid character in domain:", string(r))
				return false
			}
		}
	}

	// =========================================================
	// OPTIONS
	// =========================================================

	var server string

	timeout := 10 * time.Second

	var outputFile string

	followReferral := true

	for i := 1; i < len(args); i++ {

		switch args[i] {

		case "-server":

			if i+1 >= len(args) {
				fmt.Println("missing value for -server")
				return false
			}

			server = strings.TrimSpace(args[i+1])

			if server == "" {
				fmt.Println("WHOIS server cannot be empty")
				return false
			}

			i++

		case "-timeout":

			if i+1 >= len(args) {
				fmt.Println("missing value for -timeout")
				return false
			}

			parsed, err := time.ParseDuration(args[i+1])

			if err != nil {
				fmt.Println("invalid timeout:", args[i+1])
				return false
			}

			if parsed <= 0 {
				fmt.Println("timeout must be greater than zero")
				return false
			}

			timeout = parsed

			i++

		case "-no-follow":

			followReferral = false

		case "-o":

			if i+1 >= len(args) {
				fmt.Println("missing filename for -o")
				return false
			}

			outputFile = strings.TrimSpace(args[i+1])

			if outputFile == "" {
				fmt.Println("output filename cannot be empty")
				return false
			}

			i++

		default:

			fmt.Println("unknown option:", args[i])
			return false
		}
	}

	// =========================================================
	// DEFAULT WHOIS SERVER
	// =========================================================

	if server == "" {
		server = "whois.iana.org"
	}

	// =========================================================
	// REFERRAL SETTINGS
	// =========================================================

	maxReferrals := 5

	referralCount := 0

	visitedServers := make(map[string]bool)

	currentServer := server

	var response string

	// =========================================================
	// WHOIS QUERY / REFERRAL LOOP
	// =========================================================

	for {

		normalizedServer := strings.ToLower(
			strings.TrimSpace(currentServer),
		)

		// Prevent referral loops.
		if visitedServers[normalizedServer] {
			fmt.Println("WHOIS referral loop detected:", currentServer)
			return false
		}

		visitedServers[normalizedServer] = true

		// ---------------------------------------------------------
		// TCP ADDRESS
		// ---------------------------------------------------------

		address := net.JoinHostPort(currentServer, "43")

		// ---------------------------------------------------------
		// CONNECT
		// ---------------------------------------------------------

		conn, err := net.DialTimeout(
			"tcp",
			address,
			timeout,
		)

		if err != nil {
			fmt.Println("WHOIS connection failed:", err)
			return false
		}

		// ---------------------------------------------------------
		// SEND QUERY
		// ---------------------------------------------------------

		_, err = conn.Write(
			[]byte(target + "\r\n"),
		)

		if err != nil {
			conn.Close()

			fmt.Println("failed to send WHOIS query:", err)
			return false
		}

		// ---------------------------------------------------------
		// READ RESPONSE
		// ---------------------------------------------------------

		data, err := io.ReadAll(conn)

		conn.Close()

		if err != nil {
			fmt.Println("failed to read WHOIS response:", err)
			return false
		}

		response = string(data)

		// ---------------------------------------------------------
		// DO NOT FOLLOW?
		// ---------------------------------------------------------

		if !followReferral {
			break
		}

		// ---------------------------------------------------------
		// FIND REFERRAL
		// ---------------------------------------------------------

		referralServer := ""

		lines := strings.Split(response, "\n")

		for _, line := range lines {

			line = strings.TrimSpace(line)

			if line == "" {
				continue
			}

			colonIndex := strings.Index(line, ":")

			if colonIndex == -1 {
				continue
			}

			key := strings.TrimSpace(
				strings.ToLower(line[:colonIndex]),
			)

			if key != "refer" {
				continue
			}

			referralServer = strings.TrimSpace(
				line[colonIndex+1:],
			)

			break
		}

		// No referral means this is the final response.
		if referralServer == "" {
			break
		}

		// ---------------------------------------------------------
		// REFERRAL LIMIT
		// ---------------------------------------------------------

		if referralCount >= maxReferrals {
			fmt.Println("maximum WHOIS referral limit reached")
			break
		}

		referralCount++

		fmt.Println(
			"Following referral:",
			referralServer,
		)

		currentServer = referralServer
	}

	// =========================================================
	// TERMINAL OUTPUT
	// =========================================================

	fmt.Print(response)

	// =========================================================
	// FILE OUTPUT
	// =========================================================

	if outputFile != "" {

		extension := strings.ToLower(
			filepath.Ext(outputFile),
		)

		switch extension {

		case ".json":

			result := map[string]any{
				"target":             target,
				"target_type":        targetType,
				"server":             currentServer,
				"referrals_followed": referralCount,
				"response":           response,
			}

			err := output.SaveJSON(
				outputFile,
				result,
			)

			if err != nil {
				fmt.Println("failed to save JSON:", err)
				return false
			}

		default:

			err := output.SaveText(
				outputFile,
				response,
			)

			if err != nil {
				fmt.Println("failed to save WHOIS result:", err)
				return false
			}
		}

		fmt.Println("WHOIS result saved to:", outputFile)
	}

	return false
}
