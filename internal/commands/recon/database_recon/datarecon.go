package database_recon

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
	"sync/atomic"
	"time"
)

// ============================================================
// DATABASE RECON
// ============================================================
//
// Passive:
//   database <target>
//
// Full authorized scan:
//   database <target> --full
//
// Full mode:
//   - Resolves A / AAAA records
//   - Scans TCP ports 1-65535
//   - Identifies OPEN ports
//   - Fingerprints database protocols
//   - Extracts publicly exposed protocol information
//
// It does NOT:
//   - authenticate
//   - brute-force credentials
//   - execute SQL
//   - enumerate tables
//   - enumerate records
//   - exploit services
// ============================================================

const (
	firstTCPPort = 1
	lastTCPPort  = 65535

	tcpScanWorkers = 100
	fpWorkers      = 20

	connectTimeout = 700 * time.Millisecond
	probeTimeout   = 1800 * time.Millisecond

	maxBannerSize = 8192
)

// ============================================================
// TYPES
// ============================================================

// DatabaseReconResult represents one discovered service.
type DatabaseReconResult struct {
	IP         string
	Port       int
	Protocol   string
	Database   string
	Version    string
	Status     string
	Confidence string
	Evidence   string
	TLS        string
}

// DatabaseEndpoint is an OPEN TCP endpoint that will be
// fingerprinted.
type DatabaseEndpoint struct {
	IP   string
	Port int
}

// databaseFingerprint contains protocol-level evidence.
type databaseFingerprint struct {
	Database   string
	Version    string
	Confidence string
	Evidence   string
	TLS        string
}

// databaseScanStats contains complete TCP scan statistics.
type databaseScanStats struct {
	Tested   int64
	Open     int64
	Closed   int64
	Filtered int64
	Errors   int64
}

// ============================================================
// MAIN
// ============================================================

func DatabaseRecon(args []string) bool {

	if len(args) == 0 {
		fmt.Println("Usage: database <domain|host|IP> [--full]")
		return false
	}

	target := normalizeTarget(args[0])

	if !validTarget(target) {
		fmt.Printf("Invalid target: %s\n", target)
		return false
	}

	fullScan := false

	for _, arg := range args[1:] {

		switch strings.ToLower(
			strings.TrimSpace(arg),
		) {

		case "--full":
			fullScan = true

		default:
			fmt.Printf("Unknown option: %s\n", arg)
			fmt.Println("Usage: database <domain|host|IP> [--full]")
			return false
		}
	}

	// --------------------------------------------------------
	// DNS
	// --------------------------------------------------------

	ips, err := lookupIPs(target)

	if err != nil {
		fmt.Printf("DNS resolution failed: %v\n", err)
		return false
	}

	if len(ips) == 0 {
		fmt.Println("No IP addresses resolved.")
		return false
	}

	// --------------------------------------------------------
	// Infrastructure
	// --------------------------------------------------------

	infrastructure := detectInfrastructure(
		target,
		ips,
	)

	// --------------------------------------------------------
	// SRV records
	// --------------------------------------------------------

	passive := discoverSRV(target)

	// --------------------------------------------------------
	// Passive mode
	// --------------------------------------------------------

	if !fullScan {

		printPassiveReport(
			target,
			ips,
			infrastructure,
			passive,
		)

		return true
	}

	// --------------------------------------------------------
	// Full scan
	// --------------------------------------------------------

	fmt.Println()
	fmt.Println("Starting full TCP scan...")
	fmt.Printf("Targets      %d address(es)\n", len(ips))
	fmt.Println("Ports        1-65535")
	fmt.Printf("Workers      %d\n", tcpScanWorkers)
	fmt.Println()

	openEndpoints, stats := scanAllTCPPorts(
		ips,
	)

	// --------------------------------------------------------
	// Fingerprint
	// --------------------------------------------------------

	fmt.Println()
	fmt.Println("Fingerprinting open services...")

	active := fingerprintOpenEndpoints(
		openEndpoints,
	)

	// --------------------------------------------------------
	// Merge
	// --------------------------------------------------------

	results := append(
		passive,
		active...,
	)

	results = deduplicateResults(results)

	sortResults(results)

	// --------------------------------------------------------
	// Report
	// --------------------------------------------------------

	printFullReport(
		target,
		ips,
		infrastructure,
		results,
		stats,
	)

	return true
}

// ============================================================
// NORMALIZE TARGET
// ============================================================

func normalizeTarget(target string) string {

	target = strings.TrimSpace(target)

	target = strings.TrimPrefix(
		target,
		"https://",
	)

	target = strings.TrimPrefix(
		target,
		"http://",
	)

	if index := strings.Index(
		target,
		"/",
	); index >= 0 {
		target = target[:index]
	}

	// Handle host:port.
	if host, _, err := net.SplitHostPort(target); err == nil {
		target = host
	}

	target = strings.TrimSpace(target)
	target = strings.TrimSuffix(target, ".")

	return strings.ToLower(target)
}

// ============================================================
// VALIDATE TARGET
// ============================================================

func validTarget(target string) bool {

	if target == "" {
		return false
	}

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

		if label == "" || len(label) > 63 {
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
// DNS
// ============================================================

func lookupIPs(target string) ([]string, error) {

	if ip := net.ParseIP(target); ip != nil {
		return []string{ip.String()}, nil
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)

	defer cancel()

	hosts, err := net.DefaultResolver.LookupHost(
		ctx,
		target,
	)

	if err != nil {
		return nil, err
	}

	unique := make(
		map[string]struct{},
	)

	for _, host := range hosts {

		ip := net.ParseIP(host)

		if ip == nil {
			continue
		}

		unique[ip.String()] = struct{}{}
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
// SRV DISCOVERY
// ============================================================

func discoverSRV(
	target string,
) []DatabaseReconResult {

	type srvService struct {
		name    string
		service string
	}

	services := []srvService{
		{
			name:    "MySQL",
			service: "mysql",
		},
		{
			name:    "PostgreSQL",
			service: "postgresql",
		},
		{
			name:    "Redis",
			service: "redis",
		},
		{
			name:    "MongoDB",
			service: "mongodb",
		},
		{
			name:    "Microsoft SQL Server",
			service: "mssql",
		},
	}

	var results []DatabaseReconResult

	for _, service := range services {

		_, records, err := net.LookupSRV(
			service.service,
			"tcp",
			target,
		)

		if err != nil {
			continue
		}

		for _, record := range records {

			// IMPORTANT:
			// net.LookupSRV returns []*net.SRV.
			//
			// Therefore:
			//     record.Target
			//
			// not:
			//     srv.Target

			host := strings.TrimSuffix(
				record.Target,
				".",
			)

			results = append(
				results,
				DatabaseReconResult{
					IP:         host,
					Port:       int(record.Port),
					Protocol:   "TCP",
					Database:   service.name,
					Status:     "ADVERTISED",
					Confidence: "HIGH",
					Evidence:   "Public DNS SRV record",
					TLS:        "unknown",
				},
			)
		}
	}

	return results
}

// ============================================================
// FULL TCP SCAN
// ============================================================

func scanAllTCPPorts(
	ips []string,
) ([]DatabaseEndpoint, databaseScanStats) {

	var stats databaseScanStats

	var endpoints []DatabaseEndpoint

	var mutex sync.Mutex

	jobs := make(
		chan DatabaseEndpoint,
		tcpScanWorkers*2,
	)

	var wg sync.WaitGroup

	worker := func() {

		defer wg.Done()

		for endpoint := range jobs {

			atomic.AddInt64(
				&stats.Tested,
				1,
			)

			state := scanTCPPort(
				endpoint.IP,
				endpoint.Port,
			)

			switch state {

			case "OPEN":

				atomic.AddInt64(
					&stats.Open,
					1,
				)

				mutex.Lock()

				endpoints = append(
					endpoints,
					endpoint,
				)

				mutex.Unlock()

			case "CLOSED":

				atomic.AddInt64(
					&stats.Closed,
					1,
				)

			case "FILTERED":

				atomic.AddInt64(
					&stats.Filtered,
					1,
				)

			default:

				atomic.AddInt64(
					&stats.Errors,
					1,
				)
			}
		}
	}

	wg.Add(tcpScanWorkers)

	for i := 0; i < tcpScanWorkers; i++ {
		go worker()
	}

	for _, ip := range ips {

		for port := firstTCPPort; port <= lastTCPPort; port++ {

			jobs <- DatabaseEndpoint{
				IP:   ip,
				Port: port,
			}
		}
	}

	close(jobs)

	wg.Wait()

	sort.Slice(
		endpoints,
		func(i, j int) bool {

			if endpoints[i].IP != endpoints[j].IP {
				return endpoints[i].IP <
					endpoints[j].IP
			}

			return endpoints[i].Port <
				endpoints[j].Port
		},
	)

	return endpoints, stats
}

// ============================================================
// TCP PORT CHECK
// ============================================================

func scanTCPPort(
	ip string,
	port int,
) string {

	address := net.JoinHostPort(
		ip,
		fmt.Sprintf("%d", port),
	)

	dialer := net.Dialer{
		Timeout: connectTimeout,
	}

	conn, err := dialer.Dial(
		"tcp",
		address,
	)

	if err == nil {

		_ = conn.Close()

		return "OPEN"
	}

	if isTimeout(err) {
		return "FILTERED"
	}

	if isRefused(err) {
		return "CLOSED"
	}

	return "ERROR"
}

// ============================================================
// FINGERPRINT OPEN PORTS
// ============================================================

func fingerprintOpenEndpoints(
	endpoints []DatabaseEndpoint,
) []DatabaseReconResult {

	if len(endpoints) == 0 {
		return nil
	}

	workers := fpWorkers

	if len(endpoints) < workers {
		workers = len(endpoints)
	}

	jobs := make(
		chan DatabaseEndpoint,
		workers*2,
	)

	var wg sync.WaitGroup
	var mutex sync.Mutex

	results := make(
		[]DatabaseReconResult,
		0,
		len(endpoints),
	)

	worker := func() {

		defer wg.Done()

		for endpoint := range jobs {

			result := fingerprintEndpoint(
				endpoint,
			)

			mutex.Lock()

			results = append(
				results,
				result,
			)

			mutex.Unlock()
		}
	}

	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go worker()
	}

	for _, endpoint := range endpoints {
		jobs <- endpoint
	}

	close(jobs)

	wg.Wait()

	return results
}

// ============================================================
// ENDPOINT FINGERPRINT
// ============================================================

func fingerprintEndpoint(
	endpoint DatabaseEndpoint,
) DatabaseReconResult {

	result := DatabaseReconResult{
		IP:         endpoint.IP,
		Port:       endpoint.Port,
		Protocol:   "TCP",
		Database:   "Unknown",
		Status:     "OPEN",
		Confidence: "NONE",
		TLS:        "unknown",
		Evidence:   "TCP connection accepted",
	}

	address := net.JoinHostPort(
		endpoint.IP,
		fmt.Sprintf("%d", endpoint.Port),
	)

	conn, err := (&net.Dialer{
		Timeout: connectTimeout,
	}).Dial(
		"tcp",
		address,
	)

	if err != nil {

		result.Status = classifyTCPError(err)
		result.Evidence = err.Error()

		return result
	}

	defer conn.Close()

	// --------------------------------------------------------
	// Protocol fingerprinting
	// --------------------------------------------------------

	switch endpoint.Port {

	case 3306:
		if fp := fingerprintMySQL(conn); fp != nil {
			return applyFingerprint(result, *fp)
		}

	case 5432:
		if fp := fingerprintPostgreSQL(conn); fp != nil {
			return applyFingerprint(result, *fp)
		}

	case 6379, 6380:
		if fp := fingerprintRedis(conn); fp != nil {
			return applyFingerprint(result, *fp)
		}

	case 27017:
		if fp := fingerprintMongoDB(conn); fp != nil {
			return applyFingerprint(result, *fp)
		}

	case 1433:
		if fp := fingerprintMSSQL(conn); fp != nil {
			return applyFingerprint(result, *fp)
		}

	case 9042:
		if fp := fingerprintCassandra(conn); fp != nil {
			return applyFingerprint(result, *fp)
		}

	case 8529:
		if fp := fingerprintArangoDB(conn); fp != nil {
			return applyFingerprint(result, *fp)
		}
	}

	// --------------------------------------------------------
	// Generic banner
	// --------------------------------------------------------

	if banner := readGenericBanner(conn); banner != "" {

		result.Evidence =
			"TCP open; application banner: " +
				banner

		result.Confidence = "LOW"

		return result
	}

	result.Evidence =
		"TCP port open; database protocol not confirmed"

	return result
}

// ============================================================
// APPLY FINGERPRINT
// ============================================================

func applyFingerprint(
	result DatabaseReconResult,
	fp databaseFingerprint,
) DatabaseReconResult {

	result.Database = fp.Database
	result.Version = fp.Version
	result.Confidence = fp.Confidence
	result.Evidence = fp.Evidence
	result.TLS = fp.TLS
	result.Status = "IDENTIFIED"

	return result
}

// ============================================================
// MYSQL
// ============================================================

func fingerprintMySQL(
	conn net.Conn,
) *databaseFingerprint {

	_ = conn.SetReadDeadline(
		time.Now().Add(probeTimeout),
	)

	reader := bufio.NewReader(conn)

	header := make([]byte, 4)

	if _, err := io.ReadFull(
		reader,
		header,
	); err != nil {
		return nil
	}

	payloadLength :=
		int(header[0]) |
			int(header[1])<<8 |
			int(header[2])<<16

	if payloadLength <= 0 ||
		payloadLength > maxBannerSize {
		return nil
	}

	payload := make(
		[]byte,
		payloadLength,
	)

	if _, err := io.ReadFull(
		reader,
		payload,
	); err != nil {
		return nil
	}

	if len(payload) < 2 {
		return nil
	}

	// MySQL protocol version 10.
	if payload[0] != 0x0A {
		return nil
	}

	versionEnd := 1

	for versionEnd < len(payload) &&
		payload[versionEnd] != 0 {
		versionEnd++
	}

	version := ""

	if versionEnd > 1 {
		version = string(
			payload[1:versionEnd],
		)
	}

	tls := mysqlTLSCapability(payload)

	evidence := "MySQL protocol handshake"

	if version != "" {
		evidence += "; server version advertised"
	}

	return &databaseFingerprint{
		Database:   "MySQL",
		Version:    version,
		Confidence: "HIGH",
		Evidence:   evidence,
		TLS:        tls,
	}
}

// ============================================================
// MYSQL TLS CAPABILITY
// ============================================================

func mysqlTLSCapability(
	payload []byte,
) string {

	// Find server version terminator.
	index := bytesIndexByte(
		payload,
		0,
	)

	if index < 0 {
		return "unknown"
	}

	// Protocol:
	//
	// protocol version       1
	// server version         NUL
	// connection id          4
	// auth data              8
	// filler                 1
	// capability lower       2
	//
	pos := index + 1

	if len(payload) < pos+4+8+1+2 {
		return "unknown"
	}

	pos += 4
	pos += 8
	pos += 1

	lower := binary.LittleEndian.Uint16(
		payload[pos : pos+2],
	)

	// CLIENT_SSL = 0x0800
	if lower&0x0800 != 0 {
		return "supported"
	}

	return "not advertised"
}

// ============================================================
// POSTGRESQL
// ============================================================

func fingerprintPostgreSQL(
	conn net.Conn,
) *databaseFingerprint {

	_ = conn.SetDeadline(
		time.Now().Add(probeTimeout),
	)

	// StartupMessage:
	//
	// length
	// protocol 3.0
	// user
	// database
	//

	payload := make(
		[]byte,
		0,
		64,
	)

	protocol := make(
		[]byte,
		4,
	)

	binary.BigEndian.PutUint32(
		protocol,
		196608,
	)

	payload = append(
		payload,
		protocol...,
	)

	payload = append(
		payload,
		[]byte("user\x00zebra_probe\x00")...,
	)

	payload = append(
		payload,
		[]byte("database\x00zebra_probe\x00")...,
	)

	payload = append(
		payload,
		0,
	)

	packet := make(
		[]byte,
		4,
	)

	binary.BigEndian.PutUint32(
		packet,
		uint32(4+len(payload)),
	)

	packet = append(
		packet,
		payload...,
	)

	if _, err := conn.Write(packet); err != nil {
		return nil
	}

	response := make(
		[]byte,
		4096,
	)

	n, err := conn.Read(response)

	if err != nil && n == 0 {
		return nil
	}

	if n <= 0 {
		return nil
	}

	switch response[0] {

	case 'R':

		return &databaseFingerprint{
			Database:   "PostgreSQL",
			Confidence: "HIGH",
			Evidence:   "PostgreSQL authentication protocol response",
			TLS:        "unknown",
		}

	case 'E':

		return &databaseFingerprint{
			Database:   "PostgreSQL",
			Confidence: "HIGH",
			Evidence:   "PostgreSQL ErrorResponse protocol message",
			TLS:        "unknown",
		}

	case 'S', 'N':

		return &databaseFingerprint{
			Database:   "PostgreSQL",
			Confidence: "HIGH",
			Evidence:   "PostgreSQL backend protocol response",
			TLS:        "unknown",
		}
	}

	return nil
}

// ============================================================
// REDIS
// ============================================================

func fingerprintRedis(
	conn net.Conn,
) *databaseFingerprint {

	_ = conn.SetDeadline(
		time.Now().Add(probeTimeout),
	)

	_, err := conn.Write(
		[]byte(
			"*1\r\n$4\r\nPING\r\n",
		),
	)

	if err != nil {
		return nil
	}

	reader := bufio.NewReader(conn)

	response, err := reader.ReadString(
		'\n',
	)

	if err != nil {
		return nil
	}

	response = strings.TrimSpace(
		response,
	)

	if response == "+PONG" {

		return &databaseFingerprint{
			Database:   "Redis",
			Confidence: "HIGH",
			Evidence:   "Redis RESP PING/PONG response",
			TLS:        "unknown",
		}
	}

	if strings.HasPrefix(
		response,
		"-NOAUTH",
	) {

		return &databaseFingerprint{
			Database:   "Redis",
			Confidence: "HIGH",
			Evidence:   "Redis RESP authentication-required response",
			TLS:        "unknown",
		}
	}

	if strings.HasPrefix(
		response,
		"-ERR",
	) {

		return &databaseFingerprint{
			Database:   "Redis",
			Confidence: "HIGH",
			Evidence:   "Redis RESP error response",
			TLS:        "unknown",
		}
	}

	return nil
}

// ============================================================
// MONGODB
// ============================================================

func fingerprintMongoDB(
	conn net.Conn,
) *databaseFingerprint {

	_ = conn.SetDeadline(
		time.Now().Add(probeTimeout),
	)

	// BSON:
	//
	// {
	//   ismaster: true
	// }
	//

	bson := make(
		[]byte,
		0,
		32,
	)

	bson = append(
		bson,
		0, 0, 0, 0,
	)

	bson = append(
		bson,
		0x08,
	)

	bson = append(
		bson,
		[]byte("ismaster")...,
	)

	bson = append(
		bson,
		0,
		1,
		0,
	)

	binary.LittleEndian.PutUint32(
		bson[:4],
		uint32(len(bson)),
	)

	collection := []byte(
		"admin.$cmd\x00",
	)

	query := make(
		[]byte,
		0,
		64,
	)

	tmp := make([]byte, 4)

	// flags
	binary.LittleEndian.PutUint32(
		tmp,
		0,
	)

	query = append(query, tmp...)

	// collection
	query = append(
		query,
		collection...,
	)

	// numberToSkip
	binary.LittleEndian.PutUint32(
		tmp,
		0,
	)

	query = append(query, tmp...)

	// numberToReturn
	binary.LittleEndian.PutUint32(
		tmp,
		1,
	)

	query = append(query, tmp...)

	query = append(
		query,
		bson...,
	)

	messageLength := 16 + len(query)

	message := make(
		[]byte,
		messageLength,
	)

	// messageLength
	binary.LittleEndian.PutUint32(
		message[0:4],
		uint32(messageLength),
	)

	// requestID
	binary.LittleEndian.PutUint32(
		message[4:8],
		1,
	)

	// responseTo
	binary.LittleEndian.PutUint32(
		message[8:12],
		0,
	)

	// OP_QUERY = 2004
	binary.LittleEndian.PutUint32(
		message[12:16],
		2004,
	)

	copy(
		message[16:],
		query,
	)

	if _, err := conn.Write(message); err != nil {
		return nil
	}

	header := make(
		[]byte,
		16,
	)

	if _, err := io.ReadFull(
		conn,
		header,
	); err != nil {
		return nil
	}

	responseLength := int(
		binary.LittleEndian.Uint32(
			header[:4],
		),
	)

	if responseLength < 16 ||
		responseLength > maxBannerSize {
		return nil
	}

	body := make(
		[]byte,
		responseLength-16,
	)

	if _, err := io.ReadFull(
		conn,
		body,
	); err != nil {
		return nil
	}

	opcode := binary.LittleEndian.Uint32(
		header[12:16],
	)

	// OP_REPLY = 1
	if opcode == 1 {

		return &databaseFingerprint{
			Database:   "MongoDB",
			Confidence: "HIGH",
			Evidence:   "MongoDB wire protocol response",
			TLS:        "unknown",
		}
	}

	if containsBytes(
		body,
		[]byte("MongoDB"),
	) {

		return &databaseFingerprint{
			Database:   "MongoDB",
			Confidence: "HIGH",
			Evidence:   "MongoDB wire protocol evidence",
			TLS:        "unknown",
		}
	}

	return nil
}

// ============================================================
// MSSQL
// ============================================================

func fingerprintMSSQL(
	conn net.Conn,
) *databaseFingerprint {

	_ = conn.SetDeadline(
		time.Now().Add(probeTimeout),
	)

	// TDS PRELOGIN packet.
	//
	// Version option
	// Encryption option
	// Terminator
	//

	payload := []byte{
		0x00, 0x00, 0x0b, 0x00, 0x06,
		0x01, 0x00, 0x11, 0x00, 0x01,
		0xff,

		// Version data
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,

		// Encryption
		0x00,
	}

	packetLength := 8 + len(payload)

	packet := make(
		[]byte,
		packetLength,
	)

	packet[0] = 0x12
	packet[1] = 0x01

	binary.BigEndian.PutUint16(
		packet[2:4],
		uint16(packetLength),
	)

	packet[4] = 0
	packet[5] = 0
	packet[6] = 0
	packet[7] = 0

	copy(
		packet[8:],
		payload,
	)

	if _, err := conn.Write(packet); err != nil {
		return nil
	}

	response := make(
		[]byte,
		4096,
	)

	n, err := conn.Read(response)

	if err != nil && n == 0 {
		return nil
	}

	if n < 8 {
		return nil
	}

	// TDS PRELOGIN response.
	if response[0] != 0x12 &&
		response[0] != 0x04 {
		return nil
	}

	return &databaseFingerprint{
		Database:   "Microsoft SQL Server",
		Confidence: "HIGH",
		Evidence:   "TDS PRELOGIN protocol response",
		TLS:        "available through TDS negotiation",
	}
}

// ============================================================
// CASSANDRA
// ============================================================

func fingerprintCassandra(
	conn net.Conn,
) *databaseFingerprint {

	_ = conn.SetDeadline(
		time.Now().Add(probeTimeout),
	)

	// Cassandra native protocol STARTUP.
	//
	// version  = v4 request
	// flags
	// stream
	// opcode   = STARTUP
	// length
	//

	body := []byte{
		0x00,
		0x00,
	}

	frame := make(
		[]byte,
		9,
	)

	frame[0] = 0x04
	frame[1] = 0x00

	binary.BigEndian.PutUint16(
		frame[2:4],
		0,
	)

	frame[4] = 0x01

	binary.BigEndian.PutUint32(
		frame[5:9],
		uint32(len(body)),
	)

	frame = append(
		frame,
		body...,
	)

	if _, err := conn.Write(frame); err != nil {
		return nil
	}

	response := make(
		[]byte,
		1024,
	)

	n, err := conn.Read(response)

	if err != nil && n == 0 {
		return nil
	}

	if n < 9 {
		return nil
	}

	// Server responses have direction bit set.
	if response[0]&0x80 != 0 {

		return &databaseFingerprint{
			Database:   "Cassandra",
			Confidence: "HIGH",
			Evidence:   "Cassandra native protocol response",
			TLS:        "unknown",
		}
	}

	return nil
}

// ============================================================
// ARANGODB
// ============================================================

func fingerprintArangoDB(
	conn net.Conn,
) *databaseFingerprint {

	_ = conn.SetDeadline(
		time.Now().Add(probeTimeout),
	)

	request :=
		"HEAD / HTTP/1.1\r\n" +
			"Host: localhost\r\n" +
			"Connection: close\r\n\r\n"

	if _, err := conn.Write(
		[]byte(request),
	); err != nil {
		return nil
	}

	reader := bufio.NewReader(conn)

	statusLine, err := reader.ReadString(
		'\n',
	)

	if err != nil {
		return nil
	}

	statusLine = strings.TrimSpace(
		statusLine,
	)

	if !strings.HasPrefix(
		statusLine,
		"HTTP/",
	) {
		return nil
	}

	for i := 0; i < 30; i++ {

		line, err := reader.ReadString(
			'\n',
		)

		if err != nil {
			break
		}

		line = strings.TrimSpace(line)

		lower := strings.ToLower(line)

		if strings.HasPrefix(
			lower,
			"server:",
		) &&
			strings.Contains(
				lower,
				"arangodb",
			) {

			return &databaseFingerprint{
				Database:   "ArangoDB",
				Confidence: "HIGH",
				Evidence:   "HTTP Server header identifies ArangoDB",
				TLS:        "unknown",
			}
		}
	}

	return nil
}

// ============================================================
// GENERIC BANNER
// ============================================================

func readGenericBanner(
	conn net.Conn,
) string {

	_ = conn.SetReadDeadline(
		time.Now().Add(
			400 * time.Millisecond,
		),
	)

	buffer := make(
		[]byte,
		1024,
	)

	n, err := conn.Read(buffer)

	if err != nil || n <= 0 {
		return ""
	}

	return sanitizeBanner(
		string(buffer[:n]),
	)
}

// ============================================================
// SANITIZE BANNER
// ============================================================

func sanitizeBanner(
	value string,
) string {

	value = strings.ReplaceAll(
		value,
		"\r",
		" ",
	)

	value = strings.ReplaceAll(
		value,
		"\n",
		" ",
	)

	value = strings.ReplaceAll(
		value,
		"\x00",
		" ",
	)

	value = strings.TrimSpace(value)

	if len(value) > 120 {
		value = value[:120] + "..."
	}

	var builder strings.Builder

	for _, c := range value {

		if c >= 32 && c <= 126 {
			builder.WriteRune(c)
		}
	}

	return strings.TrimSpace(
		builder.String(),
	)
}

// ============================================================
// BYTE HELPERS
// ============================================================

func bytesIndexByte(
	data []byte,
	target byte,
) int {

	for i, value := range data {

		if value == target {
			return i
		}
	}

	return -1
}

func containsBytes(
	data []byte,
	needle []byte,
) bool {

	if len(needle) == 0 {
		return true
	}

	if len(data) < len(needle) {
		return false
	}

	for i := 0; i <= len(data)-len(needle); i++ {

		match := true

		for j := range needle {

			if data[i+j] != needle[j] {
				match = false
				break
			}
		}

		if match {
			return true
		}
	}

	return false
}

// ============================================================
// ERROR CLASSIFICATION
// ============================================================

func isTimeout(err error) bool {

	if err == nil {
		return false
	}

	if networkError, ok := err.(net.Error); ok {

		if networkError.Timeout() {
			return true
		}
	}

	return strings.Contains(
		strings.ToLower(err.Error()),
		"timeout",
	)
}

func isRefused(err error) bool {

	if err == nil {
		return false
	}

	return strings.Contains(
		strings.ToLower(err.Error()),
		"connection refused",
	)
}

func classifyTCPError(
	err error,
) string {

	if isTimeout(err) {
		return "FILTERED"
	}

	if isRefused(err) {
		return "CLOSED"
	}

	return "ERROR"
}

// ============================================================
// INFRASTRUCTURE DETECTION
// ============================================================

func detectInfrastructure(
	target string,
	ips []string,
) string {

	ctx, cancel := context.WithTimeout(
		context.Background(),
		4*time.Second,
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

		case strings.Contains(cname, "cloudflare"):
			return "Cloudflare / reverse proxy indicator"

		case strings.Contains(cname, "cloudfront"):
			return "Amazon CloudFront indicator"

		case strings.Contains(cname, "akamai"):
			return "Akamai indicator"

		case strings.Contains(cname, "fastly"):
			return "Fastly indicator"
		}
	}

	for _, ipString := range ips {

		ip := net.ParseIP(ipString)

		if ip == nil {
			continue
		}

		if isCloudflareIP(ip) {
			return "Cloudflare edge IP indicator"
		}
	}

	return ""
}

// ============================================================
// CLOUDFLARE RANGE CHECK
// ============================================================

func isCloudflareIP(
	ip net.IP,
) bool {

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

		_, network, err := net.ParseCIDR(cidr)

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
// DEDUPLICATE
// ============================================================

func deduplicateResults(
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
			"%s|%d|%s|%s",
			strings.ToLower(result.IP),
			result.Port,
			strings.ToLower(result.Database),
			strings.ToLower(result.Version),
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
// SORT
// ============================================================

func sortResults(
	results []DatabaseReconResult,
) {

	sort.Slice(
		results,
		func(i, j int) bool {

			if results[i].IP != results[j].IP {
				return results[i].IP <
					results[j].IP
			}

			if results[i].Port != results[j].Port {
				return results[i].Port <
					results[j].Port
			}

			return results[i].Database <
				results[j].Database
		},
	)
}

// ============================================================
// PASSIVE REPORT
// ============================================================

func printPassiveReport(
	target string,
	ips []string,
	infrastructure string,
	results []DatabaseReconResult,
) {

	fmt.Println()
	fmt.Println("DATABASE RECON")
	fmt.Println("════════════════════════════════════════════════")
	fmt.Println()

	fmt.Printf(
		"Target       %s\n",
		target,
	)

	fmt.Println("Mode         passive")

	fmt.Println()

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

	fmt.Println("Resolved Addresses")

	for _, ip := range ips {
		fmt.Printf(
			"  %s\n",
			ip,
		)
	}

	fmt.Println()

	fmt.Println("Database DNS Records")

	if len(results) == 0 {

		fmt.Println(
			"  None identified.",
		)

	} else {

		for _, result := range results {

			fmt.Printf(
				"  %s:%d  %s\n",
				result.IP,
				result.Port,
				result.Database,
			)

			fmt.Printf(
				"    Evidence: %s\n",
				result.Evidence,
			)
		}
	}

	fmt.Println()
}

// ============================================================
// FULL REPORT
// ============================================================

func printFullReport(
	target string,
	ips []string,
	infrastructure string,
	results []DatabaseReconResult,
	stats databaseScanStats,
) {

	fmt.Println()
	fmt.Println("DATABASE RECON")
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Printf(
		"Target       %s\n",
		target,
	)

	fmt.Println(
		"Mode         FULL TCP + DATABASE FINGERPRINT",
	)

	fmt.Println()

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

	fmt.Println("Resolved Addresses")

	for _, ip := range ips {
		fmt.Printf(
			"  %s\n",
			ip,
		)
	}

	fmt.Println()

	// --------------------------------------------------------
	// Scan statistics
	// --------------------------------------------------------

	fmt.Println("TCP Scan")

	fmt.Printf(
		"  Tested       %d\n",
		stats.Tested,
	)

	fmt.Printf(
		"  Open         %d\n",
		stats.Open,
	)

	fmt.Printf(
		"  Closed       %d\n",
		stats.Closed,
	)

	fmt.Printf(
		"  Filtered     %d\n",
		stats.Filtered,
	)

	fmt.Printf(
		"  Errors       %d\n",
		stats.Errors,
	)

	fmt.Println()

	// --------------------------------------------------------
	// Database services
	// --------------------------------------------------------

	fmt.Println("DATABASE SERVICES")
	fmt.Println("────────────────────────────────────────────────────────")

	foundDatabase := false

	for _, result := range results {

		if result.Database == "" ||
			result.Database == "Unknown" {
			continue
		}

		foundDatabase = true

		fmt.Printf(
			"%s:%d\n",
			result.IP,
			result.Port,
		)

		fmt.Printf(
			"  Database       %s\n",
			result.Database,
		)

		if result.Version != "" {

			fmt.Printf(
				"  Version        %s\n",
				result.Version,
			)

		} else {

			fmt.Println(
				"  Version        not exposed",
			)
		}

		fmt.Printf(
			"  Status         %s\n",
			result.Status,
		)

		fmt.Printf(
			"  Confidence     %s\n",
			result.Confidence,
		)

		fmt.Printf(
			"  TLS            %s\n",
			result.TLS,
		)

		fmt.Printf(
			"  Evidence       %s\n",
			result.Evidence,
		)

		fmt.Println()
	}

	if !foundDatabase {

		fmt.Println(
			"No database protocol was confirmed.",
		)

		fmt.Println()
	}

	// --------------------------------------------------------
	// Other open ports
	// --------------------------------------------------------

	fmt.Println("OTHER OPEN TCP SERVICES")
	fmt.Println("────────────────────────────────────────────────────────")

	foundOther := false

	for _, result := range results {

		if result.Database != "Unknown" {
			continue
		}

		foundOther = true

		fmt.Printf(
			"%s:%d\n",
			result.IP,
			result.Port,
		)

		fmt.Printf(
			"  %s\n",
			result.Evidence,
		)
	}

	if !foundOther {

		fmt.Println(
			"No unidentified open TCP services.",
		)
	}

	fmt.Println()

	// --------------------------------------------------------
	// Important interpretation
	// --------------------------------------------------------

	fmt.Println("NOTES")
	fmt.Println("────────────────────────────────────────────────────────")

	fmt.Println(
		"  OPEN does not automatically mean database.",
	)

	fmt.Println(
		"  A database is reported only when protocol evidence",
	)

	fmt.Println(
		"  supports the identification.",
	)

	fmt.Println(
		"  Version is shown only when publicly exposed.",
	)

	fmt.Println(
		"  No authentication or SQL queries were attempted.",
	)

	if infrastructure != "" {

		fmt.Println()

		fmt.Println(
			"  The target appears to use CDN/proxy infrastructure.",
		)

		fmt.Println(
			"  Public scan results may describe the edge rather",
		)

		fmt.Println(
			"  than the origin database server.",
		)
	}

	fmt.Println()
}
