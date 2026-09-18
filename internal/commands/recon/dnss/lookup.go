package dnss

import (
	"fmt"
	"net"
	"path/filepath"
	"strings"

	"zebra/internal/commands/output"
)

func DNSLookup(args []string) bool {

	if len(args) == 0 {
		fmt.Println("usage: dns <domain> [-o <file>]")
		fmt.Println("       dns reverse <ip> [-o <file>]")
		return false
	}

	operation := args[0]

	switch operation {

	case "reverse":

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

		if len(args) >= 4 && args[2] == "-o" {

			filename := args[3]

			if strings.EqualFold(filepath.Ext(filename), ".json") {

				data := map[string]any{
					"type":      "reverse_dns",
					"target":    ip,
					"hostnames": names,
				}

				err := output.SaveJSON(filename, data)

				if err != nil {
					fmt.Println("failed to save JSON:", err)
					return false
				}

			} else {

				err := output.SaveText(filename, content.String())

				if err != nil {
					fmt.Println("failed to save file:", err)
					return false
				}
			}

			fmt.Println("Output saved to:", filename)
		}

	default:

		domain := args[0]

		addresses, err := net.LookupHost(domain)

		if err != nil {
			fmt.Println("DNS lookup failed:", err)
			return false
		}

		fmt.Println("DNS results for:", domain)

		var content strings.Builder

		content.WriteString("DNS results for: ")
		content.WriteString(domain)
		content.WriteString("\n\n")

		for _, address := range addresses {
			fmt.Println(address)
			content.WriteString(address)
			content.WriteString("\n")
		}

		if len(args) >= 3 && args[1] == "-o" {

			filename := args[2]

			if strings.EqualFold(filepath.Ext(filename), ".json") {

				data := map[string]any{
					"type":      "dns_lookup",
					"target":    domain,
					"addresses": addresses,
				}

				err := output.SaveJSON(filename, data)

				if err != nil {
					fmt.Println("failed to save JSON:", err)
					return false
				}

			} else {

				err := output.SaveText(filename, content.String())

				if err != nil {
					fmt.Println("failed to save file:", err)
					return false
				}
			}

			fmt.Println("Output saved to:", filename)
		}
	}

	return false
}


func DNSRecordLookup(domain string, recordType string) bool {

	switch recordType {

	case "A":
	Alookup(){
		addresses, err:= net.LookupHost(domain)
	if err != nil{
		fmt.Println("A lookup failed: ", err)
		return false
	}

	for _, address := range addresses{
		fmt.Println(address)
	}
	}

	Alookup()

	case "MX":
		func MXLook(){
			records, err:= net.LookupMX(domain)
		if err != nil{
			fmt.Println("MX  lookup failed: ", err)
			return false
		}
		for _, record:= range records{
			fmt.Println(record)
		}
		}
		MXLook()

	case "NS":
		func NSlookup(){
			records, err := net.LookupNS(domain)
			if err != nil{
				fmt.Println("NS lookup failed",err)
				return false
			}
			for _, record:= range records{
				fmt.Println(record)
			}
		}
		NSlookup()

	case "TXT":
		fmt.Println("Looking up TXT records for:", domain)

	case "CNAME":
		fmt.Println("Looking up CNAME records for:", domain)

	case "ALL":



	default:
		fmt.Println("Unsupported DNS record type:", recordType)
	}

	return false
}