package commands

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sort"
	"strings"
	"sync"
	"time"
)

// ============================================================
// DATABASE RECON
// ============================================================
//
// Modes:
//
//   database <target>
//       Passive DNS/SRV reconnaissance.
//
//   database <target> --verify
//       Passive reconnaissance + bounded TCP verification +
//       safe protocol identification.
//
// IMPORTANT:
//
// This command does NOT:
//   - authenticate
//   - brute-force credentials
//   - enumerate databases
//   - execute SQL
//   - exploit services
//   - attempt credential discovery
//
// --verify should only be used against infrastructure you are
// authorized to test.
// ============================================================

// ============================================================
// TYPES
// ============================================================

type DatabaseReconResult struct {
	Host       string
	Port       int
	Protocol   string
	Database   string
	Status     string
	Confidence string
	Evidence   string
	Live       bool
}

type databaseService struct {
	Name string
	Port int
	SRV  string
}

// ============================================================
// DATABASE SERVICE DEFINITIONS
// ============================================================

var databaseServices = []databaseService{
	{
		Name: "MySQL",
		Port: 3306,
		SRV:  "_mysql._tcp",
	},
	{
		Name: "PostgreSQL",
		Port: 5432,
		SRV:  "_postgresql._tcp",
	},
	{
		Name: "Redis",
		Port: 6379,
		SRV:  "_redis._tcp",
	},
	{
		Name: "MongoDB",
		Port: 27017,
		SRV:  "_mongodb._tcp",
	},
	{
		Name: "Microsoft SQL Server",
		Port: 1433,
		SRV:  "_mssql._tcp",
	},
	{
		Name: "Oracle",
		Port: 1521,
		SRV:  "",
	},
	{
		Name: "Cassandra",
		Port: 9042,
		SRV:  "",
	},
	{
		Name: "CouchDB",
		Port: 5984,
		SRV:  "",
	},
	{
		Name: "ArangoDB",
		Port: 8529,
		SRV:  "",
	},
}

// ============================================================
// MAIN COMMAND
// ============================================================

func DatabaseRecon(args []string) bool {

	if len(args) == 0 {
		fmt.Println("Usage: database <domain|host> [--verify]")
		return false
	}

	target := strings.TrimSpace(args[0])

	if target == "" {
		fmt.Println("Error: target cannot be empty")
		return false
	}

	// --------------------------------------------------------
	// OPTIONS
	// --------------------------------------------------------

	verify := false

	for _, arg := range args[1:] {

		switch strings.ToLower(strings.TrimSpace(arg)) {

		case "--verify":
			verify = true

		case "-v":
			verify = true

		default:
			fmt.Printf("Unknown option: %s\n", arg)
			fmt.Println("Usage: database <domain|host> [--verify]")
			return false
		}
	}

	// --------------------------------------------------------
	// NORMALIZE
	// --------------------------------------------------------

	target = normalizeDatabaseTarget(target)

	if !validDatabaseTarget(target) {

		fmt.Printf(
			"Error: invalid target: %s\n",
			target,
		)

		return false
	}

	// --------------------------------------------------------
	// RESOLVE
	// --------------------------------------------------------

	ips, err := lookupDatabaseIPs(target)

	if err != nil {

		fmt.Printf(
			"Warning: DNS resolution failed: %v\n",
			err,
		)
	}

	// --------------------------------------------------------
	// INFRASTRUCTURE
	// --------------------------------------------------------

	infrastructure := detectDatabaseInfrastructure(
		target,
		ips,
	)

	// --------------------------------------------------------
	// PASSIVE DISCOVERY
	// --------------------------------------------------------

	results := discoverPassiveDatabaseServices(
		target,
	)

	// --------------------------------------------------------
	// ACTIVE VERIFICATION
	// --------------------------------------------------------

	if verify && len(ips) > 0 {

		activeResults := verifyDatabaseServices(
			ips,
		)

		results = append(
			results,
			activeResults...,
		)
	}

	// --------------------------------------------------------
	// DEDUPLICATE
	// --------------------------------------------------------

	results = deduplicateDatabaseResults(
		results,
	)

	// --------------------------------------------------------
	// SORT
	// --------------------------------------------------------

	sortDatabaseResults(
		results,
	)

	// --------------------------------------------------------
	// REPORT
	// --------------------------------------------------------

	printDatabaseReconReport(
		target,
		verify,
		results,
		ips,
		infrastructure,
	)

	// --------------------------------------------------------
	// YOUR EXISTING FILE SAVE FUNCTION
	// --------------------------------------------------------
	//
	// Connect your existing save/report function here.
	//
	// Example:
	//
	// saveDatabaseReconResult(target, results)
	//
	// --------------------------------------------------------

	return true
}

// ============================================================
// TARGET NORMALIZATION
// ============================================================

func normalizeDatabaseTarget(target string) string {

	target = strings.TrimSpace(target)

	target = strings.TrimPrefix(
		target,
		"https://",
	)

	target = strings.TrimPrefix(
		target,
		"http://",
	)

	// Remove path.
	if index := strings.Index(
		target,
		"/",
	); index >= 0 {

		target = target[:index]
	}

	// Remove port if host:port.
	if host, _, err := net.SplitHostPort(target); err == nil {

		target = host
	}

	target = strings.TrimSpace(target)

	target = strings.TrimSuffix(
		target,
		".",
	)

	return strings.ToLower(target)
}

// ============================================================
// TARGET VALIDATION
// ============================================================

func validDatabaseTarget(target string) bool {

	if target == "" {
		return false
	}

	// IPv4 / IPv6.
	if net.ParseIP(target) != nil {
		return true
	}

	if len(target) > 253 {
		return false
	}

	if strings.ContainsAny(
		target,
		" \t\r\n",
	) {
		return false
	}

	labels := strings.Split(
		target,
		".",
	)

	if len(labels) < 2 {
		return false
	}

	for _, label := range labels {

		if label == "" {
			return false
		}

		if len(label) > 63 {
			return false
		}

		if strings.HasPrefix(label, "-") ||
			strings.HasSuffix(label, "-") {
			return false
		}

		for _, c := range label {

			if (c >= 'a' && c <= 'z') ||
				(c >= '0' && c <= '9') ||
				c == '-' {
				continue
			}

			return false
		}
	}

	return true
}

// ============================================================
// DNS RESOLUTION
// ============================================================

func lookupDatabaseIPs(
	target string,
) ([]string, error) {

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)

	defer cancel()

	ips, err := net.DefaultResolver.LookupHost(
		ctx,
		target,
	)

	if err != nil {
		return nil, err
	}

	unique := make(
		map[string]struct{},
	)

	for _, ip := range ips {

		ip = strings.TrimSpace(ip)

		if net.ParseIP(ip) == nil {
			continue
		}

		unique[ip] = struct{}{}
	}

	result := make(
		[]string,
		0,
		len(unique),
	)

	for ip := range unique {
		result = append(result, ip)
	}

	sort.Strings(result)

	return result, nil
}

// ============================================================
// PASSIVE SRV DISCOVERY
// ============================================================

func discoverPassiveDatabaseServices(
	target string,
) []DatabaseReconResult {

	var results []DatabaseReconResult

	for _, service := range databaseServices {

		if service.SRV == "" {
			continue
		}

		serviceName := strings.TrimPrefix(
			service.SRV,
			"_",
		)

		_, records, err := net.LookupSRV(
			serviceName,
			"tcp",
			target,
		)

		if err != nil {
			continue
		}

		for _, record := range records {

			host := strings.TrimSuffix(
				record.Target,
				".",
			)

			results = append(
				results,
				DatabaseReconResult{
					Host:       host,
					Port:       int(record.Port),
					Protocol:   "tcp",
					Database:   service.Name,
					Status:     "advertised",
					Confidence: "high",
					Evidence:   "public DNS SRV record",
					Live:       false,
				},
			)
		}
	}

	return results
}

// ============================================================
// ACTIVE SERVICE VERIFICATION
// ============================================================

func verifyDatabaseServices(
	ips []string,
) []DatabaseReconResult {

	var results []DatabaseReconResult

	// --------------------------------------------------------
	// Limit concurrent connections.
	// --------------------------------------------------------

	const maxWorkers = 4

	type targetPort struct {
		ip   string
		port int
		name string
	}

	var jobs []targetPort

	for _, ip := range ips {

		for _, service := range databaseServices {

			jobs = append(
				jobs,
				targetPort{
					ip:   ip,
					port: service.Port,
					name: service.Name,
				},
			)
		}
	}

	jobChannel := make(
		chan targetPort,
		len(jobs),
	)

	resultChannel := make(
		chan DatabaseReconResult,
		len(jobs),
	)

	for _, job := range jobs {
		jobChannel <- job
	}

	close(jobChannel)

	var wg sync.WaitGroup

	worker := func() {

		defer wg.Done()

		for job := range jobChannel {

			result := verifyDatabaseEndpoint(
				job.ip,
				job.port,
				job.name,
			)

			// Only return interesting results.
			//
			// We don't print every closed port as a
			// database service.
			if result.Status != "closed" {
				resultChannel <- result
			}
		}
	}

	workers := maxWorkers

	if len(jobs) < workers {
		workers = len(jobs)
	}

	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go worker()
	}

	wg.Wait()

	close(resultChannel)

	for result := range resultChannel {
		results = append(
			results,
			result,
		)
	}

	return results
}

// ============================================================
// ENDPOINT VERIFICATION
// ============================================================

func verifyDatabaseEndpoint(
	host string,
	port int,
	candidate string,
) DatabaseReconResult {

	result := DatabaseReconResult{
		Host:       host,
		Port:       port,
		Protocol:   "tcp",
		Database:   candidate,
		Status:     "unknown",
		Confidence: "none",
		Live:       false,
	}

	// --------------------------------------------------------
	// TCP CONNECT
	// --------------------------------------------------------

	conn, err := (&net.Dialer{
		Timeout: 2 * time.Second,
	}).Dial(
		"tcp",
		net.JoinHostPort(
			host,
			fmt.Sprintf("%d", port),
		),
	)

	if err != nil {

		result.Status = classifyTCPError(
			err,
		)

		switch result.Status {

		case "timeout":
			result.Evidence = "TCP connection timed out"

		case "closed":
			result.Evidence = "TCP connection refused"

		default:
			result.Evidence = "TCP connection failed"
		}

		return result
	}

	defer conn.Close()

	result.Live = true
	result.Status = "reachable"
	result.Confidence = "low"
	result.Evidence = "TCP connection succeeded"

	// --------------------------------------------------------
	// SAFE PROTOCOL IDENTIFICATION
	// --------------------------------------------------------

	switch port {

	case 3306:

		if identifyMySQL(
			conn,
		) {

			result.Database = "MySQL"
			result.Status = "identified"
			result.Confidence = "high"
			result.Evidence =
				"MySQL server greeting detected"

		} else {

			result.Database = "Unknown"
			result.Evidence =
				"TCP reachable; MySQL protocol not confirmed"
		}

	case 6379:

		if identifyRedis(
			conn,
		) {

			result.Database = "Redis"
			result.Status = "identified"
			result.Confidence = "high"
			result.Evidence =
				"Redis PING/PONG response detected"

		} else {

			result.Database = "Unknown"
			result.Evidence =
				"TCP reachable; Redis protocol not confirmed"
		}

	case 5432:

		if identifyPostgreSQL(
			conn,
		) {

			result.Database = "PostgreSQL"
			result.Status = "identified"
			result.Confidence = "high"
			result.Evidence =
				"PostgreSQL protocol response detected"

		} else {

			result.Database = "Unknown"
			result.Evidence =
				"TCP reachable; PostgreSQL protocol not confirmed"
		}

	default:

		result.Database = candidate
		result.Evidence =
			"TCP reachable; protocol identification not implemented for this service"
	}

	return result
}

// ============================================================
// TCP ERROR CLASSIFICATION
// ============================================================

func classifyTCPError(
	err error,
) string {

	if err == nil {
		return "reachable"
	}

	if strings.Contains(
		strings.ToLower(err.Error()),
		"timeout",
	) {

		return "timeout"
	}

	if strings.Contains(
		strings.ToLower(err.Error()),
		"refused",
	) {

		return "closed"
	}

	return "unreachable"
}

// ============================================================
// MYSQL IDENTIFICATION
// ============================================================
//
// MySQL sends a server greeting immediately after TCP
// connection.
//
// We only READ the greeting.
//
// No authentication.
// No credentials.
// No SQL.
// ============================================================

func identifyMySQL(
	conn net.Conn,
) bool {

	_ = conn.SetReadDeadline(
		time.Now().Add(
			1500 * time.Millisecond,
		),
	)

	reader := bufio.NewReader(
		conn,
	)

	// MySQL packet header:
	//
	// 3 bytes payload length
	// 1 byte sequence ID

	header := make(
		[]byte,
		4,
	)

	// FIX:
	// ReadFull belongs to the io package.
	if _, err := io.ReadFull(
		reader,
		header,
	); err != nil {

		return false
	}

	payloadLength := int(
		header[0],
	) |
		int(header[1])<<8 |
		int(header[2])<<16

	if payloadLength <= 0 ||
		payloadLength > 1<<20 {

		return false
	}

	payload := make(
		[]byte,
		payloadLength,
	)

	// FIX:
	// ReadFull belongs to the io package.
	if _, err := io.ReadFull(
		reader,
		payload,
	); err != nil {

		return false
	}

	// MySQL protocol version 10.
	return len(payload) > 0 &&
		payload[0] == 0x0a
}

// ============================================================
// REDIS IDENTIFICATION
// ============================================================
//
// PING is a read-only Redis command.
//
// The request is bounded and does not attempt authentication.
//
// IMPORTANT:
// This is still an active protocol probe, so --verify must
// only be used against systems you are authorized to test.
// ============================================================

func identifyRedis(
	conn net.Conn,
) bool {

	_ = conn.SetDeadline(
		time.Now().Add(
			1500 * time.Millisecond,
		),
	)

	// RESP PING.
	_, err := conn.Write(
		[]byte("*1\r\n$4\r\nPING\r\n"),
	)

	if err != nil {
		return false
	}

	reader := bufio.NewReader(
		conn,
	)

	response, err := reader.ReadString(
		'\n',
	)

	if err != nil {
		return false
	}

	response = strings.TrimSpace(
		response,
	)

	return strings.HasPrefix(
		response,
		"+PONG",
	)
}

// ============================================================
// POSTGRESQL IDENTIFICATION
// ============================================================
//
// Sends a minimal PostgreSQL StartupMessage.
//
// user/database are intentionally omitted.
//
// A PostgreSQL server should respond with an authentication,
// error, or other PostgreSQL protocol message.
//
// We do not authenticate.
// ============================================================

func identifyPostgreSQL(
	conn net.Conn,
) bool {

	_ = conn.SetDeadline(
		time.Now().Add(
			1500 * time.Millisecond,
		),
	)

	// PostgreSQL protocol version 3.0.
	//
	// StartupMessage:
	//
	// int32 length
	// int32 protocol version
	// parameters...
	//
	// Empty parameter list is sufficient for a protocol-level
	// response from many PostgreSQL servers.

	packet := make(
		[]byte,
		8,
	)

	binary.BigEndian.PutUint32(
		packet[0:4],
		8,
	)

	binary.BigEndian.PutUint32(
		packet[4:8],
		196608,
	)

	if _, err := conn.Write(
		packet,
	); err != nil {

		return false
	}

	header := make(
		[]byte,
		5,
	)

	if _, err := readFullConnection(
		conn,
		header,
	); err != nil {

		return false
	}

	// PostgreSQL backend messages begin with a message type.
	//
	// Typical responses:
	//
	// R = Authentication
	// E = ErrorResponse
	// S = ParameterStatus
	//
	switch header[0] {

	case 'R', 'E', 'S':
		return true

	default:
		return false
	}
}

// ============================================================
// SAFE CONNECTION READ
// ============================================================

func readFullConnection(
	conn net.Conn,
	buffer []byte,
) ([]byte, error) {

	offset := 0

	for offset < len(buffer) {

		n, err := conn.Read(
			buffer[offset:],
		)

		if err != nil {
			return nil, err
		}

		if n == 0 {
			return nil, fmt.Errorf(
				"empty read",
			)
		}

		offset += n
	}

	return buffer, nil
}

// ============================================================
// INFRASTRUCTURE DETECTION
// ============================================================

func detectDatabaseInfrastructure(
	target string,
	ips []string,
) string {

	// --------------------------------------------------------
	// First inspect CNAME.
	// --------------------------------------------------------

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)

	defer cancel()

	cname, err := net.DefaultResolver.LookupCNAME(
		ctx,
		target,
	)

	if err == nil {

		cname = strings.ToLower(
			strings.TrimSuffix(
				cname,
				".",
			),
		)

		switch {

		case strings.Contains(
			cname,
			"cloudflare",
		):
			return "Cloudflare / reverse proxy indicator"

		case strings.Contains(
			cname,
			"cloudfront",
		):
			return "Amazon CloudFront indicator"

		case strings.Contains(
			cname,
			"akamai",
		):
			return "Akamai indicator"

		case strings.Contains(
			cname,
			"fastly",
		):
			return "Fastly indicator"
		}
	}

	// --------------------------------------------------------
	// Check known Cloudflare IPv4 ranges.
	//
	// This is intentionally only an indicator.
	// It does not prove that Cloudflare is being used.
	// --------------------------------------------------------

	for _, ipString := range ips {

		ip := net.ParseIP(ipString)

		if ip == nil {
			continue
		}

		if isKnownCloudflareIP(ip) {
			return "Cloudflare edge IP indicator"
		}
	}

	return ""
}

// ============================================================
// CLOUDFLARE IP INDICATOR
// ============================================================

func isKnownCloudflareIP(
	ip net.IP,
) bool {

	// Common Cloudflare IPv4 ranges.
	//
	// These are only indicators for report enrichment.

	ranges := []string{
		"173.245.48.0/20",
		"103.21.244.0/22",
		"103.22.200.0/22",
		"103.31.4.0/22",
		"141.101.64.0/18",
		"108.162.192.0/18",
		"190.93.240.0/20",
		"188.114.96.0/20",
		"197.234.240.0/22",
		"198.41.128.0/17",
		"162.158.0.0/15",
		"104.16.0.0/13",
		"104.24.0.0/14",
		"172.64.0.0/13",
		"131.0.72.0/22",
	}

	for _, cidr := range ranges {

		_, network, err := net.ParseCIDR(
			cidr,
		)

		if err != nil {
			continue
		}

		if network.Contains(ip) {
			return true
		}
	}

	return false
}

// ============================================================
// DEDUPLICATION
// ============================================================

func deduplicateDatabaseResults(
	results []DatabaseReconResult,
) []DatabaseReconResult {

	seen := make(
		map[string]bool,
	)

	unique := make(
		[]DatabaseReconResult,
		0,
		len(results),
	)

	for _, result := range results {

		key := fmt.Sprintf(
			"%s|%d|%s",
			strings.ToLower(
				result.Host,
			),
			result.Port,
			strings.ToLower(
				result.Database,
			),
		)

		if seen[key] {
			continue
		}

		seen[key] = true

		unique = append(
			unique,
			result,
		)
	}

	return unique
}

// ============================================================
// SORTING
// ============================================================

func sortDatabaseResults(
	results []DatabaseReconResult,
) {

	sort.Slice(
		results,
		func(i, j int) bool {

			if results[i].Port != results[j].Port {
				return results[i].Port <
					results[j].Port
			}

			if results[i].Host != results[j].Host {
				return results[i].Host <
					results[j].Host
			}

			return results[i].Database <
				results[j].Database
		},
	)
}

// ============================================================
// REPORT
// ============================================================

func printDatabaseReconReport(
	target string,
	verify bool,
	results []DatabaseReconResult,
	ips []string,
	infrastructure string,
) {

	fmt.Println()

	fmt.Println("DATABASE RECON")
	fmt.Println("────────────────────────────────────────")
	fmt.Println()

	fmt.Printf(
		"Target       %s\n",
		target,
	)

	if verify {
		fmt.Println(
			"Mode         active verification",
		)
	} else {
		fmt.Println(
			"Mode         passive",
		)
	}

	fmt.Println()

	// --------------------------------------------------------
	// Infrastructure
	// --------------------------------------------------------

	fmt.Println("Infrastructure")

	if infrastructure != "" {

		fmt.Printf(
			"  %s\n",
			infrastructure,
		)

	} else {

		fmt.Println(
			"  No common CDN/proxy indicator detected",
		)
	}

	fmt.Println()

	// --------------------------------------------------------
	// Services
	// --------------------------------------------------------

	fmt.Println("Database Services")
	fmt.Println()

	if len(results) == 0 {

		fmt.Println(
			"No database service identified.",
		)

	} else {

		fmt.Printf(
			"%-18s %-7s %-20s %-14s %-12s\n",
			"Host",
			"Port",
			"Database",
			"TCP State",
			"Confidence",
		)

		fmt.Println(
			"────────────────────────────────────────────────────────────────",
		)

		for _, result := range results {

			fmt.Printf(
				"%-18s %-7d %-20s %-14s %-12s\n",
				shortDatabaseHost(
					result.Host,
				),
				result.Port,
				result.Database,
				result.Status,
				result.Confidence,
			)

			fmt.Printf(
				"  Evidence: %s\n",
				result.Evidence,
			)

			fmt.Println()
		}
	}

	// --------------------------------------------------------
	// IP addresses
	// --------------------------------------------------------

	fmt.Println("Target IPs:")

	if len(ips) == 0 {

		fmt.Println(
			"  None resolved",
		)

	} else {

		for _, ip := range ips {

			fmt.Printf(
				"  %s\n",
				ip,
			)
		}
	}

	fmt.Println()

	// --------------------------------------------------------
	// Interpretation
	// --------------------------------------------------------

	if len(results) == 0 {

		fmt.Println("Conclusion")

		if verify {

			fmt.Println(
				"  No tested database endpoint was reachable.",
			)

		} else {

			fmt.Println(
				"  No database-specific service was exposed through passive DNS.",
			)

			fmt.Println(
				"  This does not prove that the application has no database.",
			)
		}

		fmt.Println()
	}

	if infrastructure != "" {

		fmt.Println("Note")

		fmt.Println(
			"  The public domain appears to use CDN/proxy infrastructure.",
		)

		fmt.Println(
			"  The resolved IP may therefore represent the edge rather",
		)

		fmt.Println(
			"  than the application's origin server.",
		)

		fmt.Println()
	}

	fmt.Println(
		"No authentication attempted.",
	)

	fmt.Println()
}

// ============================================================
// DISPLAY HELPERS
// ============================================================

func shortDatabaseHost(
	host string,
) string {

	host = strings.TrimSpace(host)

	if len(host) <= 18 {
		return host
	}

	return host[:15] + "..."
}
