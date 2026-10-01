package networkscan

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sort"
	"sync"
	"syscall"
	"time"
)

type TCPResult struct {
	Port   int
	State  string
	Reason string
}

func RunTCPScan(config ScanConfig) ([]TCPResult, error) {
	if !config.TCP {
		return nil, nil
	}

	if len(config.Ports) == 0 {
		return nil, fmt.Errorf("no ports available for TCP scan")
	}

	workers := config.Workers

	if workers <= 0 {
		workers = 100
	}

	if workers > len(config.Ports) {
		workers = len(config.Ports)
	}

	jobs := make(chan int)
	results := make(chan TCPResult)

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for port := range jobs {
				results <- scanTCPPort(
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

	scanResults := make([]TCPResult, 0, len(config.Ports))

	for result := range results {
		scanResults = append(scanResults, result)
	}

	sort.Slice(scanResults, func(i, j int) bool {
		return scanResults[i].Port < scanResults[j].Port
	})

	return scanResults, nil
}

func scanTCPPort(
	target string,
	port int,
	timeout time.Duration,
) TCPResult {
	address := net.JoinHostPort(
		target,
		fmt.Sprintf("%d", port),
	)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		timeout,
	)
	defer cancel()

	dialer := net.Dialer{}

	conn, err := dialer.DialContext(
		ctx,
		"tcp",
		address,
	)

	if err == nil {
		_ = conn.Close()

		return TCPResult{
			Port:   port,
			State:  "open",
			Reason: "connection succeeded",
		}
	}

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return TCPResult{
			Port:   port,
			State:  "filtered",
			Reason: "connection timed out",
		}
	}

	if errors.Is(err, syscall.ECONNREFUSED) {
		return TCPResult{
			Port:   port,
			State:  "closed",
			Reason: "connection refused",
		}
	}

	return TCPResult{
		Port:   port,
		State:  "unknown",
		Reason: err.Error(),
	}
}
