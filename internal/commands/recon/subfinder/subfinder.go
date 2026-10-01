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
	maxCTResponseSize    = 10 * 1024 * 1024
	maxConcurrentLookups = 20

	maxCTRetries     = 3
	initialCTBackoff = 500 * time.Millisecond
	maxCTBackoff     = 3 * time.Second

	// Number of common DNS names to test.
	maxCommonNames = 150
)

func ValidateDomain(domain string) error {
	if strings.TrimSpace(domain) == "" {
		return fmt.Errorf("domain is required")
	}

	if len(domain) > 253 {
		return fmt.Errorf("domain is too long")
	}

	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return fmt.Errorf("domain cannot start or end with a dot")
	}

	if strings.Contains(domain, "..") {
		return fmt.Errorf("domain contains empty labels")
	}

	for i, r := range domain {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '.' {
			continue
		}
		if i == 0 && r == '*' {
			continue
		}
		return fmt.Errorf("domain contains invalid character %q", r)
	}

	labels := strings.Split(domain, ".")
	for _, label := range labels {
		if label == "" {
			return fmt.Errorf("domain contains empty labels")
		}
		if len(label) > 63 {
			return fmt.Errorf("label %q exceeds 63 characters", label)
		}
		if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return fmt.Errorf("label %q cannot start or end with a hyphen", label)
		}
	}

	last := labels[len(labels)-1]
	if len(last) < 2 || len(last) > 63 {
		return fmt.Errorf("TLD is invalid")
	}

	return nil
}

// ============================================================
// RESULT STRUCTURES
// ============================================================

type SubdomainsResult struct {
	Target     string          `json:"target"`
	ScannedAt  time.Time       `json:"scanned_at"`
	Duration   time.Duration   `json:"duration"`
	Sources    []string        `json:"sources"`
	Subdomains []SubdomainInfo `json:"subdomains"`
	TotalFound int             `json:"total_found"`
	Errors     []string        `json:"errors,omitempty"`
}

type SubdomainInfo struct {
	Hostname string   `json:"hostname"`
	IPs      []string `json:"ips,omitempty"`
}

// ============================================================
// PROVIDER RESULT
// ============================================================

type providerResult struct {
	Name       string
	Subdomains []string
	Err        error
}

// ============================================================
// COMMON DNS NAMES
// ============================================================

var commonSubdomainNames = []string{
	"www",
	"www1",
	"www2",

	"mail",
	"mail1",
	"mail2",
	"email",
	"webmail",
	"smtp",
	"imap",
	"pop",
	"pop3",

	"ftp",
	"sftp",

	"api",
	"api1",
	"api2",
	"api-dev",
	"api-test",
	"api-stage",
	"api-staging",

	"app",
	"app1",
	"app2",

	"admin",
	"administrator",
	"portal",
	"dashboard",
	"panel",

	"login",
	"auth",
	"sso",
	"oauth",

	"dev",
	"development",
	"test",
	"testing",
	"stage",
	"staging",
	"uat",
	"qa",

	"demo",
	"beta",
	"alpha",

	"blog",
	"news",
	"forum",
	"community",
	"support",
	"help",
	"docs",
	"documentation",

	"cdn",
	"static",
	"assets",
	"media",
	"img",
	"images",
	"files",
	"download",
	"downloads",

	"status",
	"monitor",
	"monitoring",
	"metrics",
	"grafana",
	"kibana",

	"vpn",
	"remote",
	"gateway",
	"proxy",

	"ns1",
	"ns2",
	"ns3",
	"ns4",

	"mx",
	"mx1",
	"mx2",

	"db",
	"database",
	"mysql",
	"postgres",
	"postgresql",
	"redis",
	"mongo",
	"mongodb",

	"git",
	"gitlab",
	"github",
	"svn",

	"jenkins",
	"ci",
	"cd",
	"build",

	"docker",
	"registry",
	"harbor",

	"devops",
	"ops",

	"cloud",
	"storage",
	"backup",
	"backups",

	"internal",
	"intranet",
	"private",

	"server",
	"server1",
	"server2",

	"host",
	"host1",
	"host2",

	"secure",
	"www-secure",

	"members",
	"account",
	"accounts",
	"customer",
	"customers",

	"shop",
	"store",
	"payment",
	"payments",

	"calendar",
	"drive",
	"chat",
	"meet",

	"search",
	"mobile",
	"m",
	"statuspage",
}

// ============================================================
// MAIN RECON FUNCTION
// ============================================================

func SubfinderRecon(args []string) bool {
	if len(args) == 0 {
		fmt.Println("subfinder <domain>")
		return false
	}

	domain := strings.ToLower(
		strings.TrimSpace(args[0]),
	)

	if err := ValidateDomain(domain); err != nil {
		fmt.Println("Error:", err)
		return false
	}

	start := time.Now()
	scannedAt := time.Now().UTC()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		45*time.Second,
	)
	defer cancel()

	fmt.Println()
	fmt.Println("Starting subdomain reconnaissance...")
	fmt.Println("Target:", domain)
	fmt.Println()

	// --------------------------------------------------------
	// DISCOVERY
	// --------------------------------------------------------

	subdomains, discoveryErrors, sources :=
		DiscoverSubdomains(ctx, domain)

	// --------------------------------------------------------
	// DNS RESOLUTION
	// --------------------------------------------------------

	results, resolveErrors := ResolveSubdomains(
		ctx,
		subdomains,
	)

	// --------------------------------------------------------
	// COMBINE ERRORS
	// --------------------------------------------------------

	allErrors := make(
		[]string,
		0,
		len(discoveryErrors)+len(resolveErrors),
	)

	allErrors = append(
		allErrors,
		discoveryErrors...,
	)

	allErrors = append(
		allErrors,
		resolveErrors...,
	)

	sort.Strings(allErrors)

	// --------------------------------------------------------
	// BUILD REPORT
	// --------------------------------------------------------

	report := SubdomainsResult{
		Target:     domain,
		ScannedAt:  scannedAt,
		Duration:   time.Since(start),
		Sources:    sources,
		Subdomains: results,
		TotalFound: len(results),
		Errors:     allErrors,
	}

	sort.Slice(
		report.Subdomains,
		func(i, j int) bool {
			return report.Subdomains[i].Hostname <
				report.Subdomains[j].Hostname
		},
	)

	// --------------------------------------------------------
	// SAVE
	// --------------------------------------------------------

	if err := SaveSubdomainReport(report); err != nil {
		fmt.Println("Error saving report:", err)
		return false
	}

	// --------------------------------------------------------
	// PRINT
	// --------------------------------------------------------

	PrintSubdomainResult(report)

	return true
}

// ============================================================
// DISCOVERY ENGINE
// ============================================================

func DiscoverSubdomains(
	ctx context.Context,
	domain string,
) ([]string, []string, []string) {

	var (
		wg sync.WaitGroup
		mu sync.Mutex

		allSubdomains = make(map[string]struct{})
		errorsFound   []string
		sources       []string
	)

	// --------------------------------------------------------
	// CERTIFICATE TRANSPARENCY
	// --------------------------------------------------------

	wg.Add(1)

	go func() {
		defer wg.Done()

		results, err := DiscoverFromCT(
			ctx,
			domain,
		)

		mu.Lock()
		defer mu.Unlock()

		sources = append(
			sources,
			"Certificate Transparency",
		)

		if err != nil {
			errorsFound = append(
				errorsFound,
				fmt.Sprintf(
					"Certificate Transparency: %v",
					err,
				),
			)

			return
		}

		for _, hostname := range results {
			allSubdomains[hostname] = struct{}{}
		}
	}()

	// --------------------------------------------------------
	// DNS COMMON-NAME DISCOVERY
	// --------------------------------------------------------

	wg.Add(1)

	go func() {
		defer wg.Done()

		results, err := DiscoverFromDNS(
			ctx,
			domain,
		)

		mu.Lock()
		defer mu.Unlock()

		sources = append(
			sources,
			"DNS Common Names",
		)

		if err != nil {
			errorsFound = append(
				errorsFound,
				fmt.Sprintf(
					"DNS Common Names: %v",
					err,
				),
			)

			return
		}

		for _, hostname := range results {
			allSubdomains[hostname] = struct{}{}
		}
	}()

	wg.Wait()

	// --------------------------------------------------------
	// CONVERT MAP TO SLICE
	// --------------------------------------------------------

	subdomains := make(
		[]string,
		0,
		len(allSubdomains),
	)

	for hostname := range allSubdomains {
		subdomains = append(
			subdomains,
			hostname,
		)
	}

	sort.Strings(subdomains)
	sort.Strings(errorsFound)
	sort.Strings(sources)

	return subdomains, errorsFound, sources
}

// ============================================================
// CERTIFICATE TRANSPARENCY DISCOVERY
// ============================================================

func DiscoverFromCT(
	ctx context.Context,
	domain string,
) ([]string, error) {

	client := &http.Client{
		Timeout: defaultHTTPTimeout,

		CheckRedirect: func(
			req *http.Request,
			via []*http.Request,
		) error {
			return http.ErrUseLastResponse
		},
	}

	url := "https://crt.sh/?q=%25." +
		domain +
		"&output=json"

	var lastErr error

	for attempt := 1; attempt <= maxCTRetries; attempt++ {

		if err := ctx.Err(); err != nil {
			return nil, err
		}

		results, retryable, err :=
			queryCT(
				ctx,
				client,
				url,
				domain,
			)

		if err == nil {
			return results, nil
		}

		lastErr = err

		if !retryable {
			return nil, err
		}

		if attempt == maxCTRetries {
			break
		}

		backoff := initialCTBackoff

		for i := 1; i < attempt; i++ {
			backoff *= 2

			if backoff >= maxCTBackoff {
				backoff = maxCTBackoff
				break
			}
		}

		fmt.Printf(
			"Certificate Transparency attempt %d/%d failed: %v\n",
			attempt,
			maxCTRetries,
			err,
		)

		fmt.Printf(
			"Retrying in %s...\n",
			backoff.Round(time.Millisecond),
		)

		timer := time.NewTimer(backoff)

		select {
		case <-timer.C:

		case <-ctx.Done():

			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}

			return nil, ctx.Err()
		}
	}

	return nil, fmt.Errorf(
		"provider unavailable after %d attempts: %v",
		maxCTRetries,
		lastErr,
	)
}

// ============================================================
// CT HTTP REQUEST
// ============================================================

func queryCT(
	ctx context.Context,
	client *http.Client,
	url string,
	domain string,
) ([]string, bool, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		nil,
	)

	if err != nil {
		return nil, false, fmt.Errorf(
			"failed to create request: %w",
			err,
		)
	}

	req.Header.Set(
		"User-Agent",
		"Zebra/1.0.0 subdomain-recon",
	)

	req.Header.Set(
		"Accept",
		"application/json",
	)

	resp, err := client.Do(req)

	if err != nil {
		return nil, true, fmt.Errorf(
			"request failed: %w",
			err,
		)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		results, err := parseCTResponse(
			resp.Body,
			domain,
		)

		return results, false, err
	}

	switch resp.StatusCode {

	case http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
		http.StatusTooManyRequests:

		return nil, true, fmt.Errorf(
			"HTTP status %d",
			resp.StatusCode,
		)

	default:

		return nil, false, fmt.Errorf(
			"HTTP status %d",
			resp.StatusCode,
		)
	}
}

// ============================================================
// CT RESPONSE PARSER
// ============================================================

func parseCTResponse(
	reader io.Reader,
	domain string,
) ([]string, error) {

	limitedReader := io.LimitReader(
		reader,
		maxCTResponseSize+1,
	)

	body, err := io.ReadAll(
		limitedReader,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to read response: %w",
			err,
		)
	}

	if int64(len(body)) > maxCTResponseSize {
		return nil, fmt.Errorf(
			"response exceeds 10 MB limit",
		)
	}

	var records []struct {
		NameValue string `json:"name_value"`
	}

	if err := json.Unmarshal(
		body,
		&records,
	); err != nil {
		return nil, fmt.Errorf(
			"invalid CT JSON response: %w",
			err,
		)
	}

	seen := make(
		map[string]struct{},
	)

	for _, record := range records {

		for _, name := range strings.Split(
			record.NameValue,
			"\n",
		) {

			name = normalizeHostname(name)

			if name == "" {
				continue
			}

			if name != domain &&
				!strings.HasSuffix(
					name,
					"."+domain,
				) {
				continue
			}

			if err := validateDiscoveredHostname(
				name,
				domain,
			); err != nil {
				continue
			}

			seen[name] = struct{}{}
		}
	}

	results := make(
		[]string,
		0,
		len(seen),
	)

	for hostname := range seen {
		results = append(
			results,
			hostname,
		)
	}

	sort.Strings(results)

	return results, nil
}

// ============================================================
// DNS COMMON-NAME DISCOVERY
// ============================================================

func DiscoverFromDNS(
	ctx context.Context,
	domain string,
) ([]string, error) {

	names := commonSubdomainNames

	if len(names) > maxCommonNames {
		names = names[:maxCommonNames]
	}

	var (
		wg sync.WaitGroup
		mu sync.Mutex

		found = make(map[string]struct{})
	)

	semaphore := make(
		chan struct{},
		maxConcurrentLookups,
	)

	for _, name := range names {

		name := strings.TrimSpace(
			name,
		)

		if name == "" {
			continue
		}

		hostname := name + "." + domain

		wg.Add(1)

		go func() {

			defer wg.Done()

			select {
			case semaphore <- struct{}{}:

			case <-ctx.Done():
				return
			}

			defer func() {
				<-semaphore
			}()

			lookupCtx, cancel :=
				context.WithTimeout(
					ctx,
					2*time.Second,
				)

			defer cancel()

			ips, err :=
				net.DefaultResolver.LookupHost(
					lookupCtx,
					hostname,
				)

			if err != nil {
				return
			}

			if len(ips) == 0 {
				return
			}

			mu.Lock()

			found[hostname] = struct{}{}

			mu.Unlock()

		}()
	}

	wg.Wait()

	results := make(
		[]string,
		0,
		len(found),
	)

	for hostname := range found {
		results = append(
			results,
			hostname,
		)
	}

	sort.Strings(results)

	return results, nil
}

// ============================================================
// HOSTNAME NORMALIZATION
// ============================================================

func normalizeHostname(
	hostname string,
) string {

	hostname = strings.TrimSpace(
		hostname,
	)

	hostname = strings.ToLower(
		hostname,
	)

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
				"invalid hyphen placement",
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
				"invalid character %q",
				r,
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

	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)

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

			lookupCtx, cancel :=
				context.WithTimeout(
					ctx,
					defaultDNSDeadline,
				)

			defer cancel()

			ips, err :=
				net.DefaultResolver.LookupHost(
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

func uniqueStrings(
	values []string,
) []string {

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

		value = strings.TrimSpace(
			value,
		)

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

	encoder := json.NewEncoder(
		tempFile,
	)

	encoder.SetIndent(
		"",
		"    ",
	)

	if err := encoder.Encode(
		result,
	); err != nil {

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

func safeReportName(
	value string,
) string {

	value = strings.TrimSpace(value)

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

	if len(result.Subdomains) == 0 {

		fmt.Println(
			"No DNS-confirmed subdomains discovered.",
		)

	} else {

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

		fmt.Println()
		fmt.Println(
			"Recon completed with provider warnings.",
		)

	} else {

		fmt.Println()
		fmt.Println(
			"Recon completed successfully.",
		)
	}

	fmt.Println()
	fmt.Println("========================================")
}
