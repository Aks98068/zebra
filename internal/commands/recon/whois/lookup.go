package whois

import (
	"fmt"
	"io"
	"net"
	"time"
)

func WHOISLookup(args []string) bool {

	if len(args) == 0 {
		fmt.Println("usage: whois <domain> [-server <server>] [-timeout <duration>] [-o <file>]")
		return false
	}

	domain := args[0]

	var server string
	timeout := 10 * time.Second
	var outputFile string

	for i := 1; i < len(args); i++ {

		switch args[i] {

		case "-server":

			if i+1 >= len(args) {
				fmt.Println("missing value for -server")
				return false
			}

			server = args[i+1]
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

			timeout = parsed
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

	// Default WHOIS server
	if server == "" {
		server = "whois.iana.org"
	}

	// WHOIS uses TCP port 43
	address := server + ":43"

	conn, err := net.DialTimeout("tcp", address, timeout)

	if err != nil {
		fmt.Println("connection failed:", err)
		return false
	}

	defer conn.Close()

	// Send domain to WHOIS server
	_, err = conn.Write([]byte(domain + "\r\n"))

	if err != nil {
		fmt.Println("failed to send query:", err)
		return false
	}

	// Read complete response
	data, err := io.ReadAll(conn)

	if err != nil {
		fmt.Println("failed to read response:", err)
		return false
	}

	response := string(data)

	// Display response
	fmt.Print(response)

	// We will implement -o after the basic lookup works.
	_ = outputFile

	return false
}