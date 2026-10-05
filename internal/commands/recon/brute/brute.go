package brute

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	DefaultWorkers     = 4
	MaxWorkers         = 32
	DefaultMaxAttempts = uint64(1_000_000)
	MaxCandidateLength = 1024
	ProgressInterval   = time.Second
)

type Config struct {
	Workers     int
	MaxAttempts uint64
}

type Result struct {
	TargetHash string
	Algorithm  string
	Wordlist   string

	Found    bool
	Password string

	Attempts uint64
	Duration time.Duration

	Workers   int
	Completed bool
	Error     error
}

func BruteForce(args []string) bool {
	if len(args) == 0 {
		PrintUsage()
		return false
	}

	if len(args) < 2 || len(args) > 6 {
		fmt.Println("[!] invalid arguments")
		PrintUsage()
		return false
	}

	targetHash := strings.TrimSpace(args[0])
	wordlist := strings.TrimSpace(args[1])

	if err := ValidateHash(targetHash); err != nil {
		fmt.Printf("[!] invalid hash: %v\n", err)
		return false
	}

	if err := ValidateWordlist(wordlist); err != nil {
		fmt.Printf("[!] invalid wordlist: %v\n", err)
		return false
	}

	cfg := Config{
		Workers:     DefaultWorkers,
		MaxAttempts: DefaultMaxAttempts,
	}

	for i := 2; i < len(args); i++ {
		switch args[i] {

		case "--workers":
			if i+1 >= len(args) {
				fmt.Println("[!] --workers requires a value")
				return false
			}

			value, err := strconv.Atoi(args[i+1])
			if err != nil {
				fmt.Println("[!] invalid worker count")
				return false
			}

			cfg.Workers = value
			i++

		case "--max-attempts":
			if i+1 >= len(args) {
				fmt.Println("[!] --max-attempts requires a value")
				return false
			}

			value, err := strconv.ParseUint(
				args[i+1],
				10,
				64,
			)

			if err != nil {
				fmt.Println("[!] invalid maximum attempts")
				return false
			}

			cfg.MaxAttempts = value
			i++

		default:
			fmt.Printf("[!] unknown option: %s\n", args[i])
			PrintUsage()
			return false
		}
	}

	if err := ValidateConfig(cfg); err != nil {
		fmt.Printf("[!] invalid configuration: %v\n", err)
		return false
	}

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	fmt.Println()
	fmt.Println("======================================")
	fmt.Println("       ZEBRA OFFLINE HASH AUDIT")
	fmt.Println("======================================")
	fmt.Printf("[*] Algorithm    : SHA-256\n")
	fmt.Printf("[*] Workers      : %d\n", cfg.Workers)
	fmt.Printf("[*] Max attempts : %d\n", cfg.MaxAttempts)
	fmt.Println("======================================")

	result, err := Run(
		ctx,
		targetHash,
		wordlist,
		cfg,
	)

	if err != nil {
		fmt.Printf("[!] audit failed: %v\n", err)
		return false
	}

	PrintResult(result)

	return result.Found
}

func Run(
	ctx context.Context,
	targetHash string,
	wordlist string,
	cfg Config,
) (Result, error) {

	start := time.Now()

	result := Result{
		TargetHash: strings.ToLower(
			strings.TrimSpace(targetHash),
		),
		Algorithm: "SHA-256",
		Wordlist:  wordlist,
		Workers:   cfg.Workers,
	}

	if err := ValidateHash(targetHash); err != nil {
		return result, err
	}

	if err := ValidateWordlist(wordlist); err != nil {
		return result, err
	}

	if err := ValidateConfig(cfg); err != nil {
		return result, err
	}

	file, err := os.Open(wordlist)
	if err != nil {
		return result, fmt.Errorf(
			"open wordlist: %w",
			err,
		)
	}
	defer file.Close()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan string, cfg.Workers*2)

	var attempts atomic.Uint64
	var found atomic.Bool

	var (
		wg       sync.WaitGroup
		resultMu sync.Mutex
		password string
	)

	worker := func() {
		defer wg.Done()

		for {
			select {

			case <-ctx.Done():
				return

			case candidate, ok := <-jobs:
				if !ok {
					return
				}

				if found.Load() {
					return
				}

				count := attempts.Add(1)

				if count > cfg.MaxAttempts {
					cancel()
					return
				}

				candidate = strings.TrimSuffix(
					candidate,
					"\r",
				)

				if len(candidate) > MaxCandidateLength {
					continue
				}

				sum := sha256.Sum256(
					[]byte(candidate),
				)

				calculated := hex.EncodeToString(
					sum[:],
				)

				if calculated == result.TargetHash {
					if found.CompareAndSwap(
						false,
						true,
					) {
						resultMu.Lock()
						password = candidate
						resultMu.Unlock()

						cancel()
					}

					return
				}
			}
		}
	}

	wg.Add(cfg.Workers)

	for i := 0; i < cfg.Workers; i++ {
		go worker()
	}

	scanner := bufio.NewScanner(file)

	scanner.Buffer(
		make([]byte, 64*1024),
		MaxCandidateLength+1,
	)

scan:
	for scanner.Scan() {

		select {
		case <-ctx.Done():
			break scan

		case jobs <- scanner.Text():
		}
	}

	close(jobs)
	wg.Wait()

	if err := scanner.Err(); err != nil {
		return result, fmt.Errorf(
			"read wordlist: %w",
			err,
		)
	}

	result.Attempts = attempts.Load()
	result.Duration = time.Since(start)
	result.Found = found.Load()

	if result.Found {
		result.Completed = true

		resultMu.Lock()
		result.Password = password
		resultMu.Unlock()

		return result, nil
	}

	if result.Attempts >= cfg.MaxAttempts {
		result.Completed = false
		result.Error = errors.New(
			"maximum attempt limit reached",
		)

		return result, nil
	}

	result.Completed = true

	return result, nil
}

func ValidateHash(hash string) error {
	hash = strings.TrimSpace(hash)

	if hash == "" {
		return errors.New(
			"hash cannot be empty",
		)
	}

	if len(hash) != 64 {
		return errors.New(
			"SHA-256 hash must contain 64 hexadecimal characters",
		)
	}

	for _, c := range hash {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return errors.New(
				"hash contains invalid hexadecimal characters",
			)
		}
	}

	return nil
}

func ValidateWordlist(path string) error {
	path = strings.TrimSpace(path)

	if path == "" {
		return errors.New(
			"wordlist path cannot be empty",
		)
	}

	if strings.ContainsAny(
		path,
		"\r\n\t",
	) {
		return errors.New(
			"wordlist path contains invalid characters",
		)
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf(
			"wordlist is not accessible: %w",
			err,
		)
	}

	if info.IsDir() {
		return errors.New(
			"wordlist path is a directory",
		)
	}

	if info.Size() == 0 {
		return errors.New(
			"wordlist is empty",
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

	if cfg.MaxAttempts == 0 {
		return errors.New(
			"maximum attempts must be greater than zero",
		)
	}

	return nil
}

func PrintResult(result Result) {
	fmt.Println()
	fmt.Println("======================================")
	fmt.Println("          AUDIT RESULT")
	fmt.Println("======================================")

	fmt.Printf(
		"[*] Algorithm : %s\n",
		result.Algorithm,
	)

	fmt.Printf(
		"[*] Attempts  : %d\n",
		result.Attempts,
	)

	fmt.Printf(
		"[*] Workers   : %d\n",
		result.Workers,
	)

	fmt.Printf(
		"[*] Duration  : %s\n",
		result.Duration.Round(time.Millisecond),
	)

	if result.Found {
		fmt.Println("[+] Status    : MATCH FOUND")
		fmt.Printf(
			"[+] Password  : %s\n",
			result.Password,
		)
	} else {
		fmt.Println("[-] Status    : NO MATCH")

		if result.Error != nil {
			fmt.Printf(
				"[!] Reason    : %v\n",
				result.Error,
			)
		}
	}

	fmt.Println("======================================")
}

func PrintUsage() {
	fmt.Println(
		"Usage: brute <sha256-hash> <wordlist> [options]",
	)
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println(
		"  --workers <n>       Parallel workers",
	)
	fmt.Println(
		"  --max-attempts <n> Maximum candidates",
	)
}
