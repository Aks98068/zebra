package networkscan

import (
	"context"
	"fmt"
	"net"
	"sort"
	"sync"
	"time"
)

type UDPResult struct {
	Port   int
	State  string
	Reason string
}

func RunUDPScan(config ScanConfig) ([]UDPResult, error) {
	if !config.UDP {
		return nil, nil
	}

	if len(config.Ports) == 0 {
		return nil, fmt.Errorf("no ports available for UDP scan")
	}

	workers := config.Workers

	if workers <= 0 {
		workers = 100
	}

	if workers > len(config.Ports) {
		workers = len(config.Ports)
	}

	jobs := make(chan int)
	results := make(chan UDPResult)

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for port := range jobs {
				results <- scanUDPPort(
					config.Target,
					port,
					config.Timeout,
				)
			}
		}()
	}

	go func() {
		defer close(jobs)

		for _, port := range config.Ports {
			jobs <- port
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	scanResults := make([]UDPResult, 0, len(config.Ports))

	for result := range results {
		scanResults = append(scanResults, result)
	}

	sort.Slice(scanResults, func(i, j int) bool {
		return scanResults[i].Port < scanResults[j].Port
	})

	return scanResults, nil
}

func scanUDPPort(
	target string,
	port int,
	timeout time.Duration,
) UDPResult {
	address := net.JoinHostPort(
		target,
		fmt.Sprintf("%d", port),
	)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		timeout,
	)
	defer cancel()

	conn, err := net.DialUDP(
		"udp",
		nil,
		&net.UDPAddr{
			IP:   net.ParseIP(target),
			Port: port,
		},
	)

	if err != nil {
		return UDPResult{
			Port:   port,
			State:  "unknown",
			Reason: err.Error(),
		}
	}

	defer conn.Close()

	_ = address

	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return UDPResult{
			Port:   port,
			State:  "unknown",
			Reason: err.Error(),
		}
	}

	payload := []byte{}

	if _, err := conn.Write(payload); err != nil {
		return UDPResult{
			Port:   port,
			State:  "unknown",
			Reason: err.Error(),
		}
	}

	buffer := make([]byte, 4096)

	for {
		select {
		case <-ctx.Done():
			return UDPResult{
				Port:   port,
				State:  "open|filtered",
				Reason: "no response received",
			}

		default:
			n, _, err := conn.ReadFromUDP(buffer)

			if err == nil && n >= 0 {
				return UDPResult{
					Port:   port,
					State:  "open",
					Reason: "UDP response received",
				}
			}

			if err != nil {
				if isUDPTimeout(err) {
					return UDPResult{
						Port:   port,
						State:  "open|filtered",
						Reason: "no response received",
					}
				}

				return UDPResult{
					Port:   port,
					State:  "unknown",
					Reason: err.Error(),
				}
			}
		}
	}
}

func isUDPTimeout(err error) bool {
	if netErr, ok := err.(net.Error); ok {
		return netErr.Timeout()
	}

	return false
}
