package networkscan

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func ResolvePortSpecification(config ScanConfig) ([]int, error) {
	if config.AllPorts ||
		config.PortSpecification == "1-65535" {
		return generatePortRange(1, 65535), nil
	}

	if strings.TrimSpace(config.PortSpecification) == "" {
		return defaultPorts(), nil
	}

	ports, err := ParsePortSpecification(
		config.PortSpecification,
	)
	if err != nil {
		return nil, err
	}

	return ports, nil
}

func ParsePortSpecification(spec string) ([]int, error) {
	spec = strings.TrimSpace(spec)

	if spec == "" {
		return nil, fmt.Errorf(
			"port specification cannot be empty",
		)
	}

	portMap := make(map[int]struct{})

	parts := strings.Split(spec, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)

		if part == "" {
			return nil, fmt.Errorf(
				"invalid port specification: %q",
				spec,
			)
		}

		if strings.Contains(part, "-") {
			if err := parsePortRange(part, portMap); err != nil {
				return nil, err
			}

			continue
		}

		port, err := parsePort(part)
		if err != nil {
			return nil, err
		}

		portMap[port] = struct{}{}
	}

	ports := make([]int, 0, len(portMap))

	for port := range portMap {
		ports = append(ports, port)
	}

	sort.Ints(ports)

	if len(ports) == 0 {
		return nil, fmt.Errorf(
			"no valid ports specified",
		)
	}

	return ports, nil
}

func parsePort(value string) (int, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return 0, fmt.Errorf(
			"port cannot be empty",
		)
	}

	port, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf(
			"invalid port %q",
			value,
		)
	}

	if port < 1 || port > 65535 {
		return 0, fmt.Errorf(
			"port %d is out of range",
			port,
		)
	}

	return port, nil
}

func parsePortRange(value string, portMap map[int]struct{}) error {
	parts := strings.Split(value, "-")

	if len(parts) != 2 {
		return fmt.Errorf(
			"invalid port range %q",
			value,
		)
	}

	start, err := parsePort(parts[0])
	if err != nil {
		return err
	}

	end, err := parsePort(parts[1])
	if err != nil {
		return err
	}

	if start > end {
		return fmt.Errorf(
			"invalid port range %q: start is greater than end",
			value,
		)
	}

	for port := start; port <= end; port++ {
		portMap[port] = struct{}{}
	}

	return nil
}

func generatePortRange(start, end int) []int {
	ports := make([]int, 0, end-start+1)

	for port := start; port <= end; port++ {
		ports = append(ports, port)
	}

	return ports
}

func defaultPorts() []int {
	return []int{
		21,
		22,
		23,
		25,
		53,
		80,
		110,
		111,
		135,
		139,
		143,
		443,
		445,
		993,
		995,
		1433,
		1521,
		3306,
		3389,
		5432,
		5900,
		6379,
		8080,
		8443,
		9200,
		27017,
	}
}
