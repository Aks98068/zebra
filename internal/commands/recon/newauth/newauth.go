package newauth

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DefaultTimeout = 5 * time.Second
	MaxTimeout     = 30 * time.Second
	MaxWorkers     = 32
)

type Config struct {
	Timeout time.Duration
	Workers int
}

type ServiceResult struct {
	Host      string
	Port      int
	Reachable bool
	Protocol  string
	TLS       bool
	Duration  time.Duration
	Error     error
}

type Result struct {
	Target    string
	Services  []ServiceResult
	Duration  time.Duration
	Completed bool
	Error     error
}

func NetAuth(args []string) bool {
	if len(args) == 0 {
		PrintUsage()
		return false
	}

	if len(args) < 2 || len(args) > 6 {
		fmt.Println("[!] invalid arguments")
		PrintUsage()
		return false
	}

	host := strings.TrimSpace(args[0])

	port, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Println("[!] invalid port")
		return false
	}

	cfg := Config{
		Timeout: DefaultTimeout,
		Workers: 4,
	}

	for i := 2; i < len(args); i++ {
		switch args[i] {

		case "--workers":
			if i+1 >= len(args) {
				fmt.Println("[!] --workers requires a value")
				return false
			}

			cfg.Workers, err = strconv.Atoi(args[i+1])
			if err != nil {
				fmt.Println("[!] invalid worker count")
				return false
			}

			i++

		case "--timeout":
			if i+1 >= len(args) {
				fmt.Println("[!] --timeout requires a value")
				return false
			}

			seconds, parseErr := strconv.Atoi(args[i+1])
			if parseErr != nil {
				fmt.Println("[!] invalid timeout")
				return false
			}

			cfg.Timeout = time.Duration(seconds) *
				time.Second

			i++

		default:
			fmt.Printf("[!] unknown option: %s\n", args[i])
			PrintUsage()
			return false
		}
	}

	if err := ValidateHost(host); err != nil {
		fmt.Printf("[!] invalid host: %v\n", err)
		return false
	}

	if err := ValidatePort(port); err != nil {
		fmt.Printf("[!] invalid port: %v\n", err)
		return false
	}

	if err := ValidateConfig(cfg); err != nil {
		fmt.Printf("[!] invalid configuration: %v\n", err)
		return false
	}

	fmt.Println()
	fmt.Println("======================================")
	fmt.Println("      ZEBRA NETWORK AUTH ASSESSMENT")
	fmt.Println("======================================")
	fmt.Printf("[*] Target  : %s:%d\n", host, port)
	fmt.Printf("[*] Workers : %d\n", cfg.Workers)
	fmt.Printf("[*] Timeout : %s\n", cfg.Timeout)
	fmt.Println("======================================")

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	result, err := Scan(
		ctx,
		host,
		[]int{port},
		cfg,
	)

	if err != nil {
		fmt.Printf("[!] scan failed: %v\n", err)
		return false
	}

	PrintResult(result)

	return result.Completed
}

func Scan(
	ctx context.Context,
	host string,
	ports []int,
	cfg Config,
) (Result, error) {

	start := time.Now()

	result := Result{
		Target: host,
	}

	if err := ValidateHost(host); err != nil {
		return result, err
	}

	if err := ValidateConfig(cfg); err != nil {
		return result, err
	}

	jobs := make(chan int)

	results := make(chan ServiceResult)

	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()

		for {
			select {

			case <-ctx.Done():
				return

			case port, ok := <-jobs:
				if !ok {
					return
				}

				service := Probe(
					ctx,
					host,
					port,
					cfg.Timeout,
				)

				select {
				case results <- service:

				case <-ctx.Done():
					return
				}
			}
		}
	}

	workerCount := cfg.Workers

	if workerCount > len(ports) {
		workerCount = len(ports)
	}

	wg.Add(workerCount)

	for i := 0; i < workerCount; i++ {
		go worker()
	}

	go func() {
		defer close(jobs)

		for _, port := range ports {
			select {

			case jobs <- port:

			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	for service := range results {
		result.Services = append(
			result.Services,
			service,
		)
	}

	result.Duration = time.Since(start)
	result.Completed = true

	return result, nil
}

func Probe(
	ctx context.Context,
	host string,
	port int,
	timeout time.Duration,
) ServiceResult {

	start := time.Now()

	result := ServiceResult{
		Host: host,
		Port: port,
	}

	address := net.JoinHostPort(
		host,
		strconv.Itoa(port),
	)

	dialer := net.Dialer{
		Timeout: timeout,
	}

	conn, err := dialer.DialContext(
		ctx,
		"tcp",
		address,
	)

	if err != nil {
		result.Duration = time.Since(start)
		result.Error = err
		return result
	}

	defer conn.Close()

	result.Reachable = true
	result.Protocol = IdentifyProtocol(port)

	result.Duration = time.Since(start)

	return result
}

func IdentifyProtocol(port int) string {
	switch port {

	case 20, 21:
		return "FTP"

	case 22:
		return "SSH"

	case 23:
		return "TELNET"

	case 25, 587:
		return "SMTP"

	case 53:
		return "DNS"

	case 80:
		return "HTTP"

	case 110:
		return "POP3"

	case 143:
		return "IMAP"

	case 443:
		return "HTTPS"

	case 445:
		return "SMB"

	case 993:
		return "IMAPS"

	case 995:
		return "POP3S"

	default:
		return "UNKNOWN"
	}
}

func ValidateHost(host string) error {
	host = strings.TrimSpace(host)

	if host == "" {
		return errors.New(
			"host cannot be empty",
		)
	}

	if strings.ContainsAny(
		host,
		" \r\n\t",
	) {
		return errors.New(
			"host contains invalid whitespace",
		)
	}

	return nil
}

func ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return errors.New(
			"port must be between 1 and 65535",
		)
	}

	return nil
}

func ValidateConfig(cfg Config) error {
	if cfg.Workers < 1 {
		return errors.New(
			"workers must be greater than zero",
		)
	}

	if cfg.Workers > MaxWorkers {
		return fmt.Errorf(
			"workers cannot exceed %d",
			MaxWorkers,
		)
	}

	if cfg.Timeout <= 0 {
		return errors.New(
			"timeout must be greater than zero",
		)
	}

	if cfg.Timeout > MaxTimeout {
		return fmt.Errorf(
			"timeout cannot exceed %s",
			MaxTimeout,
		)
	}

	return nil
}

func PrintResult(result Result) {
	fmt.Println()
	fmt.Println("======================================")
	fmt.Println("          NETWORK RESULT")
	fmt.Println("======================================")

	for _, service := range result.Services {

		fmt.Printf(
			"[%s] %s:%d",
			status(service.Reachable),
			service.Host,
			service.Port,
		)

		if service.Protocol != "" {
			fmt.Printf(
				"  protocol=%s",
				service.Protocol,
			)
		}

		fmt.Printf(
			"  duration=%s\n",
			service.Duration.Round(time.Millisecond),
		)
	}

	fmt.Println()
	fmt.Printf(
		"[*] Duration : %s\n",
		result.Duration.Round(time.Millisecond),
	)

	fmt.Printf(
		"[*] Status   : %s\n",
		completionStatus(result.Completed),
	)

	fmt.Println("======================================")
}

func status(reachable bool) string {
	if reachable {
		return "+"
	}

	return "-"
}

func completionStatus(completed bool) string {
	if completed {
		return "COMPLETE"
	}

	return "INCOMPLETE"
}

func PrintUsage() {
	fmt.Println(
		"Usage: netauth <host> <port> [options]",
	)
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println(
		"  --workers <n>   Number of parallel probes",
	)
	fmt.Println(
		"  --timeout <s>   Connection timeout in seconds",
	)
}
