package subfinder

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
)

const (
	defaultHTTPTimeout   = 10 * time.Second
	defaultDNSDeadline   = 5 * time.Second
	maxCTResponseSize    = 10 * 1024 * 1024 // 10 MB
	maxConcurrentLookups = 10
)

// ============================================================
// RESULT STRUCTURES
// ============================================================

type SubdomainsResult struct {
	Target     string           `json:"target"`
	ScannedAt  time.Time        `json:"scanned_at"`
	Duration   time.Duration    `json:"duration"`
	Sources    []string         `json:"sources"`
	Subdomains []SubdomainInfo  `json:"subdomains"`
	TotalFound int              `json:"total_found"`
	Errors     []string         `json:"errors,omitempty"`
}

type SubdomainInfo struct {
	Hostname string   `json:"hostname"`
	IPs      []string `json:"ips,omitempty"`
}

// ============================================================
// MAIN RECON FUNCTION
// ============================================================

func SubfinderRecon(args []string) bool {
	if len(args) == 0 {
		fmt.Println("subfinder <domain> [options]")
		return false
	}

	domain := strings.ToLower(strings.TrimSpace(args[0]))

	if err := ValidateDomain(domain); err != nil {
		fmt.Println("Error:", err)
		return false
	}

	start := time.Now()
	scannedAt := time.Now().UTC()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	fmt.Println()
	fmt.Println("Starting subdomain reconnaissance...")
	fmt.Println("Target:", domain)
	fmt.Println()

	// --------------------------------------------------------
	// DISCOVERY
	// --------------------------------------------------------

	subdomains, err := DiscoverSubdomains(ctx, domain)

	if err != nil {
		fmt.Println("Error:", err)
		return false
	}

	if len(subdomains) == 0 {
		fmt.Println("No subdomains discovered.")
		return true
	}

	// --------------------------------------------------------
	// DNS RESOLUTION
	// --------------------------------------------------------

	results, resolveErrors := ResolveSubdomains(ctx, subdomains)

	// --------------------------------------------------------
	// BUILD REPORT
	// --------------------------------------------------------

	report := SubdomainsResult{
		Target:     domain,
		ScannedAt:  scannedAt,
		Duration:   time.Since(start),
		Sources:    []string{"Certificate Transparency"},
		Subdomains: results,
		TotalFound: len(results),
		Errors:     resolveErrors,
	}

	// --------------------------------------------------------
	// SORT RESULTS
	// --------------------------------------------------------

	sort.Slice(
		report.Subdomains,
		func(i, j int) bool {
			return report.Subdomains[i].Hostname <
				report.Subdomains[j].Hostname
		},
	)

	// --------------------------------------------------------
	// SAVE REPORT
	// --------------------------------------------------------

	if err := SaveSubdomainReport(report); err != nil {
		fmt.Println("Error saving report:", err)
		return false
	}

	// --------------------------------------------------------
	// PRINT RESULT
	// --------------------------------------------------------

	PrintSubdomainResult(report)

	return true
}

// ============================================================
// DOMAIN VALIDATION
// ============================================================

func ValidateDomain(domain string) error {
	domain = strings.TrimSpace(domain)

	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	if strings.IndexFunc(domain, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsControl(r)
	}) != -1 {
		return fmt.Errorf(
			"domain contains invalid whitespace or control characters",
		)
	}

	if strings.Contains(domain, "://") {
		return fmt.Errorf(
			"invalid domain: URL scheme is not allowed",
		)
	}

	if strings.Contains(domain, "/") {
		return fmt.Errorf(
			"invalid domain: path is not allowed",
		)
	}

	if strings.Contains(domain, ":") {
		return fmt.Errorf(
			"invalid domain: port is not allowed",
		)
	}

	if net.ParseIP(domain) != nil {
		return fmt.Errorf(
			"IP addresses are not supported",
		)
	}

	if strings.HasSuffix(domain, ".") {
		domain = strings.TrimSuffix(domain, ".")
	}

	if !strings.Contains(domain, ".") {
		return fmt.Errorf(
			"invalid domain: domain must contain a dot",
		)
	}

	if len(domain) > 253 {
		return fmt.Errorf(
			"domain exceeds the maximum length of 253 characters",
		)
	}

	labels := strings.Split(domain, ".")

	for _, label := range labels {
		if label == "" {
			return fmt.Errorf(
				"domain contains an empty label",
			)
		}

		if len(label) > 63 {
			return fmt.Errorf(
				"domain label exceeds 63 characters",
			)
		}

		if strings.HasPrefix(label, "-") ||
			strings.HasSuffix(label, "-") {
			return fmt.Errorf(
				"domain labels cannot begin or end with '-'",
			)
		}

		for _, r := range label {
			if unicode.IsLetter(r) ||
				unicode.IsDigit(r) ||
				r == '-' ||
				r == '_' {
				continue
			}

			return fmt.Errorf(
				"domain contains invalid character %q",
				r,
			)
		}
	}

	return nil
}

// ============================================================
// PASSIVE SUBDOMAIN DISCOVERY
// ============================================================

func DiscoverSubdomains(
	ctx context.Context,
	domain string,
) ([]string, error) {

	client := &http.Client{
		Timeout: defaultHTTPTimeout,
	}

	// Certificate Transparency query.
	url := "https://crt.sh/?q=%25." +
		domain +
		"&output=json"

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create CT request: %w",
			err,
		)
	}

	req.Header.Set(
		"User-Agent",
		"Zebra/1.0.0 subdomain-recon",
	)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(
			"certificate transparency request failed: %w",
			err,
		)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"certificate transparency returned HTTP status %d",
			resp.StatusCode,
		)
	}

	// Prevent an unexpectedly large response from being
	// loaded into memory.
	limitedReader := io.LimitReader(
		resp.Body,
		maxCTResponseSize+1,
	)

	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read certificate transparency response: %w",
			err,
		)
	}

	if int64(len(body)) > maxCTResponseSize {
		return nil, fmt.Errorf(
			"certificate transparency response exceeds size limit",
		)
	}

	// --------------------------------------------------------
	// CT JSON STRUCTURE
	// --------------------------------------------------------

	var records []struct {
		NameValue string `json:"name_value"`
	}

	if err := json.Unmarshal(body, &records); err != nil {
		return nil, fmt.Errorf(
			"failed to parse certificate transparency response: %w",
			err,
		)
	}

	// --------------------------------------------------------
	// NORMALIZATION + DEDUPLICATION
	// --------------------------------------------------------

	seen := make(map[string]struct{})

	for _, record := range records {

		// A CT record can contain multiple names separated
		// by newline characters.
		names := strings.Split(
			record.NameValue,
			"\n",
		)

		for _, name := range names {

			name = normalizeHostname(name)

			if name == "" {
				continue
			}

			// Only keep the requested domain and its
			// subdomains.
			if name != domain &&
				!strings.HasSuffix(name, "."+domain) {
				continue
			}

			if err := validateDiscoveredHostname(
				name,
				domain,
			); err != nil {
				continue
			}

			if _, exists := seen[name]; exists {
				continue
			}

			seen[name] = struct{}{}
		}
	}

	// Convert set to slice.
	subdomains := make(
		[]string,
		0,
		len(seen),
	)

	for name := range seen {
		subdomains = append(
			subdomains,
			name,
		)
	}

	// Deterministic output.
	sort.Strings(subdomains)

	return subdomains, nil
}

// ============================================================
// HOSTNAME NORMALIZATION
// ============================================================

func normalizeHostname(hostname string) string {
	hostname = strings.TrimSpace(hostname)

	hostname = strings.ToLower(hostname)

	hostname = strings.TrimPrefix(
		hostname,
		"*.",
	)

	hostname = strings.TrimSuffix(
		hostname,
		".",
	)

	return hostname
}

// ============================================================
// VALIDATE DISCOVERED HOSTNAME
// ============================================================

func validateDiscoveredHostname(
	hostname string,
	domain string,
) error {

	if hostname == "" {
		return fmt.Errorf(
			"empty hostname",
		)
	}

	if len(hostname) > 253 {
		return fmt.Errorf(
			"hostname too long",
		)
	}

	if hostname != domain &&
		!strings.HasSuffix(
			hostname,
			"."+domain,
		) {
		return fmt.Errorf(
			"hostname is outside target domain",
		)
	}

	labels := strings.Split(
		hostname,
		".",
	)

	for _, label := range labels {

		if label == "" {
			return fmt.Errorf(
				"hostname contains empty label",
			)
		}

		if len(label) > 63 {
			return fmt.Errorf(
				"hostname label is too long",
			)
		}

		if strings.HasPrefix(label, "-") ||
			strings.HasSuffix(label, "-") {
			return fmt.Errorf(
				"hostname label has invalid hyphen placement",
			)
		}
	}

	return nil
}

// ============================================================
// DNS RESOLUTION
// ============================================================

func ResolveSubdomains(
	ctx context.Context,
	subdomains []string,
) ([]SubdomainInfo, []string) {

	if len(subdomains) == 0 {
		return []SubdomainInfo{}, nil
	}

	results := make(
		[]SubdomainInfo,
		0,
		len(subdomains),
	)

	errorsFound := make(
		[]string,
		0,
	)

	var mu sync.Mutex
	var wg sync.WaitGroup

	semaphore := make(
		chan struct{},
		maxConcurrentLookups,
	)

	for _, hostname := range subdomains {

		hostname := hostname

		wg.Add(1)

		go func() {
			defer wg.Done()

			select {
			case semaphore <- struct{}{}:
			case <-ctx.Done():
				mu.Lock()

				errorsFound = append(
					errorsFound,
					fmt.Sprintf(
						"%s: %v",
						hostname,
						ctx.Err(),
					),
				)

				mu.Unlock()

				return
			}

			defer func() {
				<-semaphore
			}()

			lookupCtx, cancel := context.WithTimeout(
				ctx,
				defaultDNSDeadline,
			)
			defer cancel()

			ips, err := net.DefaultResolver.LookupHost(
				lookupCtx,
				hostname,
			)

			if err != nil {

				mu.Lock()

				errorsFound = append(
					errorsFound,
					fmt.Sprintf(
						"%s: DNS resolution failed: %v",
						hostname,
						err,
					),
				)

				mu.Unlock()

				return
			}

			ips = uniqueStrings(ips)

			mu.Lock()

			results = append(
				results,
				SubdomainInfo{
					Hostname: hostname,
					IPs:      ips,
				},
			)

			mu.Unlock()
		}()
	}

	wg.Wait()

	sort.Slice(
		results,
		func(i, j int) bool {
			return results[i].Hostname <
				results[j].Hostname
		},
	)

	sort.Strings(errorsFound)

	return results, errorsFound
}

// ============================================================
// STRING DEDUPLICATION
// ============================================================

func uniqueStrings(values []string) []string {

	seen := make(
		map[string]struct{},
		len(values),
	)

	result := make(
		[]string,
		0,
		len(values),
	)

	for _, value := range values {

		value = strings.TrimSpace(value)

		if value == "" {
			continue
		}

		if _, exists := seen[value]; exists {
			continue
		}

		seen[value] = struct{}{}

		result = append(
			result,
			value,
		)
	}

	sort.Strings(result)

	return result
}

// ============================================================
// SAVE REPORT
// ============================================================

func SaveSubdomainReport(
	result SubdomainsResult,
) error {

	reportDirectory := filepath.Join(
		"reports",
		"subfinder",
	)

	if err := os.MkdirAll(
		reportDirectory,
		0750,
	); err != nil {
		return fmt.Errorf(
			"failed to create report directory: %w",
			err,
		)
	}

	filename := fmt.Sprintf(
		"%s_%s.json",
		safeReportName(result.Target),
		result.ScannedAt.UTC().Format(
			"20060102_150405",
		),
	)

	finalPath := filepath.Join(
		reportDirectory,
		filename,
	)

	tempFile, err := os.CreateTemp(
		reportDirectory,
		".subfinder-*.tmp",
	)
	if err != nil {
		return fmt.Errorf(
			"failed to create temporary report: %w",
			err,
		)
	}

	tempPath := tempFile.Name()

	defer os.Remove(tempPath)

	encoder := json.NewEncoder(tempFile)
	encoder.SetIndent("", "    ")

	if err := encoder.Encode(result); err != nil {

		tempFile.Close()

		return fmt.Errorf(
			"failed to encode report: %w",
			err,
		)
	}

	if err := tempFile.Sync(); err != nil {

		tempFile.Close()

		return fmt.Errorf(
			"failed to sync report: %w",
			err,
		)
	}

	if err := tempFile.Close(); err != nil {
		return fmt.Errorf(
			"failed to close temporary report: %w",
			err,
		)
	}

	if err := os.Rename(
		tempPath,
		finalPath,
	); err != nil {
		return fmt.Errorf(
			"failed to finalize report: %w",
			err,
		)
	}

	fmt.Println(
		"Report saved:",
		finalPath,
	)

	return nil
}

// ============================================================
// SAFE REPORT FILENAME
// ============================================================

func safeReportName(value string) string {

	value = strings.TrimSpace(
		value,
	)

	if value == "" {
		return "unknown"
	}

	var builder strings.Builder

	for _, r := range value {

		switch {

		case unicode.IsLetter(r):
			builder.WriteRune(r)

		case unicode.IsDigit(r):
			builder.WriteRune(r)

		case r == '.' ||
			r == '-' ||
			r == '_':
			builder.WriteRune(r)

		default:
			builder.WriteRune('_')
		}
	}

	name := builder.String()

	if name == "" {
		return "unknown"
	}

	return name
}

// ============================================================
// PRINT RESULTS
// ============================================================

func PrintSubdomainResult(
	result SubdomainsResult,
) {

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("          SUBDOMAIN RECON")
	fmt.Println("========================================")

	fmt.Println(
		"Target:",
		result.Target,
	)

	fmt.Println(
		"Found:",
		result.TotalFound,
	)

	fmt.Println(
		"Duration:",
		result.Duration.Round(
			time.Millisecond,
		),
	)

	fmt.Println(
		"Sources:",
		strings.Join(
			result.Sources,
			", ",
		),
	)

	fmt.Println()

	for _, subdomain := range result.Subdomains {

		fmt.Println(
			"[+]",
			subdomain.Hostname,
		)

		if len(subdomain.IPs) == 0 {

			fmt.Println(
				"    IPs: none",
			)

			continue
		}

		fmt.Println(
			"    IPs:",
			strings.Join(
				subdomain.IPs,
				", ",
			),
		)
	}

	if len(result.Errors) > 0 {

		fmt.Println()
		fmt.Println(
			"Warnings:",
		)

		for _, scanError := range result.Errors {

			fmt.Println(
				"  -",
				scanError,
			)
		}
	}

	fmt.Println()
	fmt.Println("========================================")
}