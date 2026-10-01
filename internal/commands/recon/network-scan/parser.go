package networkscan

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

func ParseScanArguments(args []string) (ScanConfig, error) {
	config := ScanConfig{
		TCP:     true,
		Timeout: 3 * time.Second,
		Timing:  3,
		Workers: 100,
	}

	if len(args) == 0 {
		return ScanConfig{}, fmt.Errorf("no target specified")
	}

	if strings.HasPrefix(args[0], "-") {
		return ScanConfig{}, fmt.Errorf(
			"target must be specified before options",
		)
	}

	config.Target = strings.TrimSpace(args[0])

	if config.Target == "" {
		return ScanConfig{}, fmt.Errorf("target cannot be empty")
	}

	i := 1
	tcpExplicit := false

	for i < len(args) {
		arg := strings.TrimSpace(args[i])

		if arg == "" {
			return ScanConfig{}, fmt.Errorf(
				"empty argument at position %d",
				i+1,
			)
		}

		switch {
		case arg == "-sT":
			config.TCP = true
			tcpExplicit = true

		case arg == "-sU":
			config.UDP = true

			if !tcpExplicit {
				config.TCP = false
			}

		case arg == "-sV":
			config.ServiceDetection = true
			config.VersionDetection = true

		case arg == "-O":
			config.OSDetection = true

		case arg == "-v":
			config.Verbose = true

		case arg == "--open":
			config.OpenOnly = true

		case arg == "--reason":
			config.ShowReason = true

		case arg == "-p-", arg == "--all-ports":
			config.PortSpecification = "1-65535"
			config.AllPorts = true

		case arg == "-p":
			if i+1 >= len(args) {
				return ScanConfig{}, fmt.Errorf(
					"missing port specification after -p",
				)
			}

			i++

			portSpec := strings.TrimSpace(args[i])

			if portSpec == "" {
				return ScanConfig{}, fmt.Errorf(
					"port specification cannot be empty",
				)
			}

			config.PortSpecification = portSpec
			config.AllPorts = false

		case arg == "--timeout":
			if i+1 >= len(args) {
				return ScanConfig{}, fmt.Errorf(
					"missing timeout value after --timeout",
				)
			}

			i++

			value := strings.TrimSpace(args[i])

			timeout, err := time.ParseDuration(value)
			if err != nil {
				return ScanConfig{}, fmt.Errorf(
					"invalid timeout value %q: %w",
					value,
					err,
				)
			}

			config.Timeout = timeout

		case strings.HasPrefix(arg, "-T"):
			value := strings.TrimPrefix(arg, "-T")

			if value == "" {
				return ScanConfig{}, fmt.Errorf(
					"option -T requires a value from 1 to 5",
				)
			}

			timing, err := strconv.Atoi(value)
			if err != nil || timing < 1 || timing > 5 {
				return ScanConfig{}, fmt.Errorf(
					"invalid timing value %q: must be between 1 and 5",
					value,
				)
			}

			config.Timing = timing

		case arg == "-o":
			if i+1 >= len(args) {
				return ScanConfig{}, fmt.Errorf(
					"option -o requires a file path",
				)
			}

			i++

			config.OutputFile = strings.TrimSpace(args[i])

			if config.OutputFile == "" {
				return ScanConfig{}, fmt.Errorf(
					"output file path cannot be empty",
				)
			}

			config.OutputType = "text"

		case arg == "-oJ":
			if i+1 >= len(args) {
				return ScanConfig{}, fmt.Errorf(
					"option -oJ requires a file path",
				)
			}

			i++

			config.OutputFile = strings.TrimSpace(args[i])

			if config.OutputFile == "" {
				return ScanConfig{}, fmt.Errorf(
					"output file path cannot be empty",
				)
			}

			config.OutputType = "json"

		case arg == "-oX":
			if i+1 >= len(args) {
				return ScanConfig{}, fmt.Errorf(
					"option -oX requires a file path",
				)
			}

			i++

			config.OutputFile = strings.TrimSpace(args[i])

			if config.OutputFile == "" {
				return ScanConfig{}, fmt.Errorf(
					"output file path cannot be empty",
				)
			}

			config.OutputType = "xml"

		case arg == "-oH":
			if i+1 >= len(args) {
				return ScanConfig{}, fmt.Errorf(
					"option -oH requires a file path",
				)
			}

			i++

			config.OutputFile = strings.TrimSpace(args[i])

			if config.OutputFile == "" {
				return ScanConfig{}, fmt.Errorf(
					"output file path cannot be empty",
				)
			}

			config.OutputType = "html"

		case arg == "-h", arg == "--help":
			return ScanConfig{}, fmt.Errorf("help requested")

		default:
			return ScanConfig{}, fmt.Errorf(
				"unknown option: %q",
				arg,
			)
		}

		i++
	}

	if err := ValidateScanConfig(config); err != nil {
		return ScanConfig{}, err
	}

	return config, nil
}

func ValidateScanConfig(config ScanConfig) error {
	if strings.TrimSpace(config.Target) == "" {
		return fmt.Errorf("target cannot be empty")
	}

	if !config.TCP && !config.UDP {
		return fmt.Errorf(
			"at least one scan type must be enabled",
		)
	}

	if config.Timeout <= 0 {
		return fmt.Errorf(
			"timeout must be greater than zero",
		)
	}

	if config.Workers <= 0 {
		return fmt.Errorf(
			"worker count must be greater than zero",
		)
	}

	if config.Timing < 1 || config.Timing > 5 {
		return fmt.Errorf(
			"timing must be between T1 and T5",
		)
	}

	if config.AllPorts &&
		config.PortSpecification != "1-65535" {
		return fmt.Errorf(
			"invalid all-port configuration",
		)
	}

	if config.PortSpecification != "" &&
		strings.TrimSpace(config.PortSpecification) == "" {
		return fmt.Errorf(
			"port specification cannot be empty",
		)
	}

	if config.OutputFile != "" &&
		config.OutputType == "" {
		return fmt.Errorf(
			"output type is missing",
		)
	}

	switch config.OutputType {
	case "", "text", "json", "xml", "html":
	default:
		return fmt.Errorf(
			"unsupported output type: %s",
			config.OutputType,
		)
	}

	if config.VersionDetection &&
		!config.ServiceDetection {
		return fmt.Errorf(
			"version detection requires service detection",
		)
	}

	return nil
}

func ValidateTarget(target string) (string, error) {
	target = strings.TrimSpace(target)

	if target == "" {
		return "", fmt.Errorf("target cannot be empty")
	}

	if strings.ContainsAny(target, " \t\r\n") {
		return "", fmt.Errorf("target cannot contain whitespace")
	}

	if ip := net.ParseIP(target); ip != nil {
		return ip.String(), nil
	}

	if err := validateHostname(target); err != nil {
		return "", err
	}

	return strings.ToLower(strings.TrimSuffix(target, ".")), nil
}

func validateHostname(host string) error {
	if len(host) > 253 {
		return fmt.Errorf("hostname is too long")
	}

	if strings.HasPrefix(host, ".") ||
		strings.HasSuffix(host, ".") {
		return fmt.Errorf("invalid hostname: %q", host)
	}

	labels := strings.Split(host, ".")

	for _, label := range labels {
		if label == "" {
			return fmt.Errorf("invalid hostname: %q", host)
		}

		if len(label) > 63 {
			return fmt.Errorf(
				"hostname label is too long: %q",
				label,
			)
		}

		if label[0] == '-' ||
			label[len(label)-1] == '-' {
			return fmt.Errorf(
				"hostname label cannot start or end with '-': %q",
				label,
			)
		}

		for _, r := range label {
			if !(r >= 'a' && r <= 'z' ||
				r >= 'A' && r <= 'Z' ||
				r >= '0' && r <= '9' ||
				r == '-') {
				return fmt.Errorf(
					"invalid character in hostname: %q",
					host,
				)
			}
		}
	}

	return nil
}
