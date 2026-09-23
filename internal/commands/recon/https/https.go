package https

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type HTTPConfig struct {
	Method          string
	Timeout         time.Duration
	FollowRedirects bool
	MaxRedirects    int
	UserAgent       string
	Headers         map[string]string
	Proxy           string
	InsecureTLS     bool
	BodyLimit       int64
	OutputFile      string
	Body            string
	BodyFile        string
	ContentType     string
}

type HTTPResult struct {
	TargetURL       string
	FinalURL        string
	Method          string
	StatusCode      int
	Status          string
	Protocol        string
	ResponseTime    time.Duration
	ContentType     string
	ContentLength   int64
	Headers         map[string][]string
	Cookies         []HTTPCookie
	Redirects       []HTTPRedirect
	TLS             *TLSInfo
	SecurityHeaders SecurityHeaders
	Title           string
	Technologies    []string
	Body            string
	BodyTruncated   bool
	URL             URLInfo
	DNS             DNSInfo
	Timing          TimingInfo
	Methods         MethodInfo
	Server          string
	PoweredBy       string
}

type HTTPCookie struct {
	Name     string
	Value    string
	Path     string
	Domain   string
	Expires  time.Time
	Secure   bool
	HTTPOnly bool
	SameSite string
}

type HTTPRedirect struct {
	FromURL    string
	ToURL      string
	StatusCode int
	Status     string
}

type TLSInfo struct {
	Version          string
	CipherSuite      string
	ServerName       string
	Negotiated       string
	PeerCertificates []CertificateInfo
}

type CertificateInfo struct {
	Subject   string
	Issuer    string
	Serial    string
	NotBefore time.Time
	NotAfter  time.Time
	DNSNames  []string
}

type SecurityHeaders struct {
	ContentSecurityPolicy     string
	StrictTransportSecurity   string
	XContentTypeOptions       string
	XFrameOptions             string
	ReferrerPolicy            string
	PermissionsPolicy         string
	CrossOriginOpenerPolicy   string
	CrossOriginResourcePolicy string
}

type URLInfo struct {
	Scheme   string
	Host     string
	Hostname string
	Port     string
	Path     string
	Query    string
}

type TimingInfo struct {
	DNSLookup    time.Duration
	TCPConnect   time.Duration
	TLSHandshake time.Duration
	TTFB         time.Duration
	Total        time.Duration
}

type MethodInfo struct {
	Advertised []string
	Source     string
}

type DNSInfo struct {
	Addresses []string
}

// ============================================================
// HTTP RECON
// ============================================================

func HTTPRecon(args []string) bool {
	if len(args) == 0 {
		fmt.Println("http <url> [options]")
		return false
	}

	target := args[0]

	method := "GET"
	var timeout time.Duration

	followRedirects := true
	maxRedirects := 10

	userAgent := ""
	headers := make(map[string]string)

	proxy := ""
	insecureTLS := false

	bodyLimit := int64(0)

	outputFile := ""
	requestBody := ""
	bodyFile := ""
	contentType := ""

	for i := 1; i < len(args); i++ {
		switch args[i] {

		case "-method":
			if i+1 >= len(args) {
				fmt.Println("missing value for method")
				return false
			}

			methodValue := strings.ToUpper(args[i+1])

			switch methodValue {
			case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "TRACE", "CONNECT":
				method = methodValue

			default:
				fmt.Println("invalid HTTP method:", methodValue)
				return false
			}

			i++

		case "-body":
			if i+1 >= len(args) {
				fmt.Println("missing value for body")
				return false
			}
			requestBody = args[i+1]
			i++

		case "-body-file":
			if i+1 >= len(args) {
				fmt.Println("missing value for body file")
				return false
			}
			bodyFile = args[i+1]
			i++

		case "-content-type":
			if i+1 >= len(args) {
				fmt.Println("missing value for content type")
				return false
			}
			contentType = args[i+1]
			i++

		case "-timeout":
			if i+1 >= len(args) {
				fmt.Println("missing value for timeout")
				return false
			}

			parsedTimeout, err := time.ParseDuration(args[i+1])

			if err != nil || parsedTimeout <= 0 {
				fmt.Println("invalid timeout")
				return false
			}

			timeout = parsedTimeout
			i++

		case "-follow":
			followRedirects = true

		case "-no-follow":
			followRedirects = false

		case "-max-redirects":
			if i+1 >= len(args) {
				fmt.Println("missing value for max redirects")
				return false
			}

			parsedRedirects, err := strconv.Atoi(args[i+1])

			if err != nil || parsedRedirects < 0 {
				fmt.Println("invalid max redirects")
				return false
			}

			maxRedirects = parsedRedirects
			i++

		case "-user-agent":
			if i+1 >= len(args) {
				fmt.Println("missing value for user agent")
				return false
			}

			userAgent = args[i+1]
			i++

		case "-header":
			if i+1 >= len(args) {
				fmt.Println("missing value for header")
				return false
			}

			key, value, err := parseHeader(args[i+1])

			if err != nil {
				fmt.Println("invalid header:", err)
				return false
			}

			headers[key] = value
			i++

		case "-proxy":
			if i+1 >= len(args) {
				fmt.Println("missing value for proxy")
				return false
			}

			proxy = args[i+1]
			i++

		case "-insecure":
			insecureTLS = true

		case "-body-limit":
			if i+1 >= len(args) {
				fmt.Println("missing value for body limit")
				return false
			}

			limit, err := parseBodyLimit(args[i+1])

			if err != nil {
				fmt.Println("invalid body limit:", err)
				return false
			}

			bodyLimit = limit
			i++

		case "-o":
			if i+1 >= len(args) {
				fmt.Println("missing value for output")
				return false
			}

			outputFile = args[i+1]
			i++

		default:
			fmt.Println("invalid option:", args[i])
			return false
		}
	}

	config := HTTPConfig{
		Method:          method,
		Timeout:         timeout,
		FollowRedirects: followRedirects,
		MaxRedirects:    maxRedirects,
		UserAgent:       userAgent,
		Headers:         headers,
		Proxy:           proxy,
		InsecureTLS:     insecureTLS,
		BodyLimit:       bodyLimit,
		OutputFile:      outputFile,
		Body:            requestBody,
		BodyFile:        bodyFile,
		ContentType:     contentType,
	}

	if config.Body != "" && config.BodyFile != "" {
		fmt.Println("use either -body or -body-file, not both")
		return false
	}

	if config.BodyFile != "" {
		data, err := os.ReadFile(config.BodyFile)
		if err != nil {
			fmt.Println("body file error:", err)
			return false
		}
		config.Body = string(data)
	}

	if config.ContentType == "" && config.Body != "" {
		config.ContentType = "application/json"
	}

	targetURL, err := parseURL(target)

	if err != nil {
		fmt.Println("URL error:", err)
		return false
	}

	dnsInfo := resolveDNS(targetURL.Hostname())

	transport, err := buildTransport(config)

	if err != nil {
		fmt.Println("transport error:", err)
		return false
	}

	redirects := make([]HTTPRedirect, 0)

	client := buildClient(
		config,
		transport,
		&redirects,
	)

	req, err := buildRequest(
		&config,
		targetURL,
	)

	if err != nil {
		fmt.Println("request error:", err)
		return false
	}

	resp, elapsed, timing, err := executeRequest(
		client,
		req,
	)

	if err != nil {
		fmt.Println("request failed:", err)
		return false
	}

	result, err := collectResponse(
		resp,
		config.BodyLimit,
	)

	if err != nil {
		fmt.Println("response error:", err)
		return false
	}

	result.TargetURL = targetURL.String()
	result.Method = config.Method
	result.ResponseTime = elapsed
	result.Redirects = redirects
	result.Timing = timing
	result.URL = makeURLInfo(targetURL)
	result.DNS = dnsInfo
	result.Server = headerValue(result.Headers, "Server")
	result.PoweredBy = headerValue(result.Headers, "X-Powered-By")
	result.Methods = analyzeMethods(result.Headers)

	if resp.Request != nil {
		result.FinalURL = resp.Request.URL.String()
	}

	if resp.TLS != nil {
		result.TLS = collectTLS(resp.TLS)
	}

	result.SecurityHeaders =
		analyzeSecurityHeaders(result.Headers)

	result.Title =
		extractHTMLTitle(
			result.Body,
			result.ContentType,
		)

	result.Technologies =
		detectTechnologies(
			result.Headers,
			result.Body,
		)

	printResult(result)

	if config.OutputFile != "" {
		err := saveResult(
			config.OutputFile,
			result,
		)

		if err != nil {
			fmt.Println("output error:", err)
			return false
		}

		fmt.Println()
		fmt.Println("Result saved to:", config.OutputFile)
	}

	return false
}

func makeURLInfo(target *url.URL) URLInfo {
	port := target.Port()
	if port == "" {
		if target.Scheme == "https" {
			port = "443"
		} else if target.Scheme == "http" {
			port = "80"
		}
	}

	return URLInfo{
		Scheme:   target.Scheme,
		Host:     target.Host,
		Hostname: target.Hostname(),
		Port:     port,
		Path:     target.EscapedPath(),
		Query:    target.RawQuery,
	}
}

func resolveDNS(host string) DNSInfo {
	ips, err := net.LookupIP(host)
	if err != nil {
		return DNSInfo{}
	}

	seen := make(map[string]bool)
	addresses := make([]string, 0, len(ips))

	for _, ip := range ips {
		value := ip.String()
		if !seen[value] {
			seen[value] = true
			addresses = append(addresses, value)
		}
	}

	return DNSInfo{Addresses: addresses}
}

func headerValue(headers map[string][]string, name string) string {
	for key, values := range headers {
		if strings.EqualFold(key, name) && len(values) > 0 {
			return strings.Join(values, ", ")
		}
	}
	return ""
}

func analyzeMethods(headers map[string][]string) MethodInfo {
	if allow := headerValue(headers, "Allow"); allow != "" {
		return MethodInfo{
			Advertised: splitMethods(allow),
			Source:     "Allow",
		}
	}

	if allow := headerValue(headers, "Access-Control-Allow-Methods"); allow != "" {
		return MethodInfo{
			Advertised: splitMethods(allow),
			Source:     "Access-Control-Allow-Methods",
		}
	}

	return MethodInfo{}
}

func splitMethods(value string) []string {
	parts := strings.Split(value, ",")
	methods := make([]string, 0, len(parts))
	seen := make(map[string]bool)

	for _, part := range parts {
		method := strings.ToUpper(strings.TrimSpace(part))
		if method == "" || seen[method] {
			continue
		}
		seen[method] = true
		methods = append(methods, method)
	}

	return methods
}

// ============================================================
// URL
// ============================================================

func parseURL(target string) (*url.URL, error) {
	if !strings.Contains(target, "://") {
		target = "https://" + target
	}

	parsed, err := url.Parse(target)

	if err != nil {
		return nil, err
	}

	switch parsed.Scheme {
	case "http", "https":

	default:
		return nil, fmt.Errorf(
			"unsupported URL scheme: %s",
			parsed.Scheme,
		)
	}

	if parsed.Host == "" {
		return nil, fmt.Errorf(
			"target host is empty",
		)
	}

	return parsed, nil
}

// ============================================================
// TRANSPORT
// ============================================================

func buildTransport(
	config HTTPConfig,
) (*http.Transport, error) {

	transport := http.DefaultTransport.(*http.Transport).
		Clone()

	if config.Proxy != "" {

		proxyURL, err := url.Parse(
			config.Proxy,
		)

		if err != nil {
			return nil, err
		}

		switch proxyURL.Scheme {
		case "http", "https":

		default:
			return nil, fmt.Errorf(
				"unsupported proxy scheme: %s",
				proxyURL.Scheme,
			)
		}

		if proxyURL.Host == "" {
			return nil, fmt.Errorf(
				"proxy host is empty",
			)
		}

		transport.Proxy =
			http.ProxyURL(proxyURL)
	}

	if config.InsecureTLS {

		transport.TLSClientConfig =
			&tls.Config{
				InsecureSkipVerify: true,
			}
	}

	if config.Timeout > 0 {

		transport.TLSHandshakeTimeout =
			config.Timeout

		transport.ResponseHeaderTimeout =
			config.Timeout

		transport.ExpectContinueTimeout =
			config.Timeout
	}

	return transport, nil
}

// ============================================================
// CLIENT
// ============================================================

func buildClient(
	config HTTPConfig,
	transport *http.Transport,
	redirects *[]HTTPRedirect,
) *http.Client {

	client := &http.Client{
		Transport: transport,
	}

	if config.Timeout > 0 {
		client.Timeout = config.Timeout
	}

	client.CheckRedirect =
		func(
			req *http.Request,
			via []*http.Request,
		) error {

			if len(via) > 0 {

				previous :=
					via[len(via)-1]

				statusCode := 0
				status := ""

				if previous.Response != nil {
					statusCode =
						previous.Response.StatusCode

					status =
						previous.Response.Status
				}

				*redirects =
					append(
						*redirects,
						HTTPRedirect{
							FromURL:    previous.URL.String(),
							ToURL:      req.URL.String(),
							StatusCode: statusCode,
							Status:     status,
						},
					)
			}

			if !config.FollowRedirects {
				return http.ErrUseLastResponse
			}

			if len(via) >= config.MaxRedirects {
				return fmt.Errorf(
					"maximum redirects exceeded: %d",
					config.MaxRedirects,
				)
			}

			return nil
		}

	return client
}

// ============================================================
// REQUEST
// ============================================================

func buildRequest(
	config *HTTPConfig,
	target *url.URL,
) (*http.Request, error) {

	var body io.Reader

	if config.Body != "" {
		body = strings.NewReader(config.Body)
	}

	req, err := http.NewRequest(
		config.Method,
		target.String(),
		body,
	)

	if err != nil {
		return nil, err
	}

	if config.UserAgent != "" {
		req.Header.Set("User-Agent", config.UserAgent)
	}

	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	if config.ContentType != "" && config.Body != "" {
		req.Header.Set("Content-Type", config.ContentType)
	}

	return req, nil
}

// ============================================================
// EXECUTE
// ============================================================

func executeRequest(
	client *http.Client,
	req *http.Request,
) (*http.Response, time.Duration, TimingInfo, error) {

	var timing TimingInfo
	var dnsStart, connectStart, tlsStart, requestStart time.Time
	var firstByte bool

	trace := &httptrace.ClientTrace{
		DNSStart: func(httptrace.DNSStartInfo) {
			dnsStart = time.Now()
		},
		DNSDone: func(httptrace.DNSDoneInfo) {
			if !dnsStart.IsZero() {
				timing.DNSLookup = time.Since(dnsStart)
			}
		},
		ConnectStart: func(_, _ string) {
			connectStart = time.Now()
		},
		ConnectDone: func(_, _ string, _ error) {
			if !connectStart.IsZero() && timing.TCPConnect == 0 {
				timing.TCPConnect = time.Since(connectStart)
			}
		},
		TLSHandshakeStart: func() {
			tlsStart = time.Now()
		},
		TLSHandshakeDone: func(tls.ConnectionState, error) {
			if !tlsStart.IsZero() {
				timing.TLSHandshake = time.Since(tlsStart)
			}
		},
		GotFirstResponseByte: func() {
			if !firstByte {
				firstByte = true
				timing.TTFB = time.Since(requestStart)
			}
		},
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	requestStart = time.Now()
	resp, err := client.Do(req)
	timing.Total = time.Since(requestStart)

	if err != nil {
		return nil, timing.Total, timing, err
	}

	if timing.TTFB == 0 {
		timing.TTFB = timing.Total
	}

	return resp, timing.Total, timing, nil
}

// ============================================================
// RESPONSE
// ============================================================

func collectResponse(
	resp *http.Response,
	bodyLimit int64,
) (HTTPResult, error) {

	var result HTTPResult

	if resp == nil {
		return result, fmt.Errorf(
			"response is nil",
		)
	}

	defer resp.Body.Close()

	result.StatusCode =
		resp.StatusCode

	result.Status =
		resp.Status

	result.Protocol =
		resp.Proto

	result.Headers =
		make(map[string][]string)

	for key, values := range resp.Header {

		copied := make(
			[]string,
			len(values),
		)

		copy(copied, values)

		result.Headers[key] = copied
	}

	result.ContentType =
		resp.Header.Get("Content-Type")

	result.ContentLength =
		resp.ContentLength

	result.Cookies =
		collectCookies(resp.Cookies())

	if bodyLimit <= 0 {
		bodyLimit = 1024 * 1024
	}

	reader := io.LimitReader(
		resp.Body,
		bodyLimit+1,
	)

	data, err := io.ReadAll(reader)

	if err != nil {
		return result, err
	}

	if int64(len(data)) > bodyLimit {

		result.BodyTruncated = true

		data =
			data[:int(bodyLimit)]
	}

	result.Body =
		string(data)

	return result, nil
}

// ============================================================
// COOKIES
// ============================================================

func collectCookies(
	cookies []*http.Cookie,
) []HTTPCookie {

	result := make(
		[]HTTPCookie,
		0,
		len(cookies),
	)

	for _, cookie := range cookies {

		sameSite := ""

		switch cookie.SameSite {

		case http.SameSiteDefaultMode:
			sameSite = "Default"

		case http.SameSiteLaxMode:
			sameSite = "Lax"

		case http.SameSiteStrictMode:
			sameSite = "Strict"

		case http.SameSiteNoneMode:
			sameSite = "None"
		}

		result = append(
			result,
			HTTPCookie{
				Name:     cookie.Name,
				Value:    cookie.Value,
				Path:     cookie.Path,
				Domain:   cookie.Domain,
				Expires:  cookie.Expires,
				Secure:   cookie.Secure,
				HTTPOnly: cookie.HttpOnly,
				SameSite: sameSite,
			},
		)
	}

	return result
}

// ============================================================
// TLS
// ============================================================

func collectTLS(
	state *tls.ConnectionState,
) *TLSInfo {

	if state == nil {
		return nil
	}

	result := &TLSInfo{
		Version: tlsVersionName(
			state.Version,
		),

		CipherSuite: tls.CipherSuiteName(
			state.CipherSuite,
		),

		ServerName: state.ServerName,

		Negotiated: state.NegotiatedProtocol,
	}

	for _, cert := range state.PeerCertificates {

		result.PeerCertificates =
			append(
				result.PeerCertificates,
				CertificateInfo{
					Subject: cert.Subject.String(),

					Issuer: cert.Issuer.String(),

					Serial: cert.SerialNumber.String(),

					NotBefore: cert.NotBefore,

					NotAfter: cert.NotAfter,

					DNSNames: cert.DNSNames,
				},
			)
	}

	return result
}

// ============================================================
// TLS VERSION
// ============================================================

func tlsVersionName(
	version uint16,
) string {

	switch version {

	case tls.VersionTLS10:
		return "TLS 1.0"

	case tls.VersionTLS11:
		return "TLS 1.1"

	case tls.VersionTLS12:
		return "TLS 1.2"

	case tls.VersionTLS13:
		return "TLS 1.3"

	default:
		return fmt.Sprintf(
			"Unknown (0x%04x)",
			version,
		)
	}
}

// ============================================================
// SECURITY HEADERS
// ============================================================

func analyzeSecurityHeaders(
	headers map[string][]string,
) SecurityHeaders {

	get := func(name string) string {

		for key, values := range headers {

			if strings.EqualFold(
				key,
				name,
			) &&
				len(values) > 0 {

				return values[0]
			}
		}

		return ""
	}

	return SecurityHeaders{

		ContentSecurityPolicy: get("Content-Security-Policy"),

		StrictTransportSecurity: get("Strict-Transport-Security"),

		XContentTypeOptions: get("X-Content-Type-Options"),

		XFrameOptions: get("X-Frame-Options"),

		ReferrerPolicy: get("Referrer-Policy"),

		PermissionsPolicy: get("Permissions-Policy"),

		CrossOriginOpenerPolicy: get("Cross-Origin-Opener-Policy"),

		CrossOriginResourcePolicy: get("Cross-Origin-Resource-Policy"),
	}
}

// ============================================================
// HTML TITLE
// ============================================================

func extractHTMLTitle(
	body string,
	contentType string,
) string {

	contentType =
		strings.ToLower(contentType)

	if !strings.Contains(
		contentType,
		"text/html",
	) {
		return ""
	}

	doc, err := html.Parse(
		strings.NewReader(body),
	)

	if err != nil {
		return ""
	}

	var findTitle func(*html.Node) string

	findTitle =
		func(node *html.Node) string {

			if node.Type ==
				html.ElementNode &&
				strings.EqualFold(
					node.Data,
					"title",
				) {

				if node.FirstChild != nil {
					return strings.TrimSpace(
						node.FirstChild.Data,
					)
				}

				return ""
			}

			for child :=
				node.FirstChild; child != nil; child = child.NextSibling {

				title :=
					findTitle(child)

				if title != "" {
					return title
				}
			}

			return ""
		}

	return findTitle(doc)
}

// ============================================================
// TECHNOLOGY DETECTION
// ============================================================

func detectTechnologies(
	headers map[string][]string,
	body string,
) []string {

	technologies :=
		make([]string, 0)

	add := func(name string) {

		for _, existing := range technologies {

			if existing == name {
				return
			}
		}

		technologies =
			append(
				technologies,
				name,
			)
	}

	getHeader :=
		func(name string) string {

			for key, values := range headers {

				if strings.EqualFold(
					key,
					name,
				) &&
					len(values) > 0 {

					return strings.Join(
						values,
						" ",
					)
				}
			}

			return ""
		}

	server :=
		strings.ToLower(
			getHeader("Server"),
		)

	powered :=
		strings.ToLower(
			getHeader("X-Powered-By"),
		)

	bodyLower :=
		strings.ToLower(body)

	if strings.Contains(
		server,
		"nginx",
	) {
		add("Nginx")
	}

	if strings.Contains(
		server,
		"apache",
	) {
		add("Apache")
	}

	if strings.Contains(
		server,
		"cloudflare",
	) {
		add("Cloudflare")
	}

	if strings.Contains(
		powered,
		"php",
	) {
		add("PHP")
	}

	if strings.Contains(
		powered,
		"express",
	) {
		add("Express")
	}

	if strings.Contains(
		bodyLower,
		"__next_data__",
	) {
		add("Next.js")
	}

	if strings.Contains(
		bodyLower,
		"/_next/static/",
	) {
		add("Next.js")
	}

	if strings.Contains(
		bodyLower,
		"wp-content/",
	) {
		add("WordPress")
	}

	if strings.Contains(bodyLower, "react-dom") ||
		strings.Contains(bodyLower, "reactroot") ||
		strings.Contains(bodyLower, "data-reactroot") {
		add("React")
	}

	if strings.Contains(bodyLower, "vue") ||
		strings.Contains(bodyLower, "__vue__") ||
		strings.Contains(bodyLower, "data-v-") {
		add("Vue.js")
	}

	if strings.Contains(bodyLower, "ng-version") ||
		strings.Contains(bodyLower, "angular") {
		add("Angular")
	}

	if strings.Contains(bodyLower, "bootstrap") ||
		strings.Contains(bodyLower, "bootstrap.min.css") {
		add("Bootstrap")
	}

	if strings.Contains(bodyLower, "jquery") ||
		strings.Contains(bodyLower, "jquery.min.js") {
		add("jQuery")
	}

	if strings.Contains(server, "iis") {
		add("Microsoft IIS")
	}

	if strings.Contains(powered, "asp.net") {
		add("ASP.NET")
	}

	if strings.Contains(powered, "node") {
		add("Node.js")
	}

	if strings.Contains(powered, "django") {
		add("Django")
	}

	return technologies
}

// ============================================================
// TERMINAL OUTPUT
// ============================================================

func printResult(
	result HTTPResult,
) {

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("           HTTP RECON RESULT")
	fmt.Println("========================================")

	fmt.Println("Target:         ", result.TargetURL)
	fmt.Println("Final URL:      ", result.FinalURL)
	fmt.Println("Method:         ", result.Method)
	fmt.Println("Status:         ", result.Status)
	fmt.Println("Status Code:    ", result.StatusCode)
	fmt.Println("Protocol:       ", result.Protocol)
	fmt.Println("Response Time:  ", result.ResponseTime)
	fmt.Println("Content-Type:   ", result.ContentType)
	fmt.Println("Content-Length: ", result.ContentLength)

	fmt.Println()
	fmt.Println("--------------- URL --------------------")
	fmt.Println("Scheme:          ", result.URL.Scheme)
	fmt.Println("Host:            ", result.URL.Host)
	fmt.Println("Hostname:        ", result.URL.Hostname)
	fmt.Println("Port:            ", result.URL.Port)
	fmt.Println("Path:            ", result.URL.Path)
	if result.URL.Query != "" {
		fmt.Println("Query:           ", result.URL.Query)
	} else {
		fmt.Println("Query:            None")
	}

	fmt.Println()
	fmt.Println("--------------- DNS --------------------")
	if len(result.DNS.Addresses) == 0 {
		fmt.Println("No DNS addresses detected")
	} else {
		for _, address := range result.DNS.Addresses {
			fmt.Println(address)
		}
	}

	fmt.Println()
	fmt.Println("--------------- TIMING -----------------")
	fmt.Println("DNS Lookup:      ", result.Timing.DNSLookup)
	fmt.Println("TCP Connect:     ", result.Timing.TCPConnect)
	fmt.Println("TLS Handshake:   ", result.Timing.TLSHandshake)
	fmt.Println("TTFB:            ", result.Timing.TTFB)
	fmt.Println("Total:           ", result.Timing.Total)

	fmt.Println()
	fmt.Println("--------------- SERVER -----------------")
	if result.Server != "" {
		fmt.Println("Server:          ", result.Server)
	} else {
		fmt.Println("Server:           Not disclosed")
	}
	if result.PoweredBy != "" {
		fmt.Println("X-Powered-By:     ", result.PoweredBy)
	} else {
		fmt.Println("X-Powered-By:      Not disclosed")
	}

	fmt.Println()
	fmt.Println("----------- HTTP METHODS ----------------")
	if len(result.Methods.Advertised) == 0 {
		fmt.Println("Advertised:       Not disclosed")
		fmt.Println("Source:           No Allow or Access-Control-Allow-Methods header")
	} else {
		fmt.Println("Advertised:      ", strings.Join(result.Methods.Advertised, ", "))
		fmt.Println("Source:           ", result.Methods.Source)
	}

	fmt.Println()
	fmt.Println("--------------- HEADERS ----------------")

	if len(result.Headers) == 0 {

		fmt.Println("None")

	} else {

		for key, values := range result.Headers {

			fmt.Printf(
				"%s: %s\n",
				key,
				strings.Join(
					values,
					", ",
				),
			)
		}
	}

	fmt.Println()
	fmt.Println("--------------- COOKIES ----------------")

	if len(result.Cookies) == 0 {

		fmt.Println("None")

	} else {

		for _, cookie := range result.Cookies {

			fmt.Printf(
				"%s=%s",
				cookie.Name,
				cookie.Value,
			)

			fmt.Printf(
				" Secure=%t",
				cookie.Secure,
			)

			fmt.Printf(
				" HttpOnly=%t",
				cookie.HTTPOnly,
			)

			if cookie.SameSite != "" {

				fmt.Printf(
					" SameSite=%s",
					cookie.SameSite,
				)
			}

			fmt.Println()
		}
	}

	fmt.Println()
	fmt.Println("--------------- REDIRECTS --------------")

	if len(result.Redirects) == 0 {

		fmt.Println("None")

	} else {

		for i, redirect := range result.Redirects {

			fmt.Printf(
				"%d. %s -> %s [%d %s]\n",
				i+1,
				redirect.FromURL,
				redirect.ToURL,
				redirect.StatusCode,
				redirect.Status,
			)
		}
	}

	fmt.Println()
	fmt.Println("--------------- TLS --------------------")

	if result.TLS == nil {

		fmt.Println("Not available")

	} else {

		fmt.Println(
			"Version:",
			result.TLS.Version,
		)

		fmt.Println(
			"Cipher Suite:",
			result.TLS.CipherSuite,
		)

		fmt.Println(
			"Server Name:",
			result.TLS.ServerName,
		)

		fmt.Println(
			"Negotiated Protocol:",
			result.TLS.Negotiated,
		)

		for i, cert := range result.TLS.PeerCertificates {

			fmt.Printf(
				"Certificate %d:\n",
				i+1,
			)

			fmt.Println(
				"  Subject:",
				cert.Subject,
			)

			fmt.Println(
				"  Issuer:",
				cert.Issuer,
			)

			fmt.Println(
				"  Serial:",
				cert.Serial,
			)

			fmt.Println(
				"  Valid From:",
				cert.NotBefore,
			)

			fmt.Println(
				"  Valid Until:",
				cert.NotAfter,
			)

			if len(cert.DNSNames) > 0 {

				fmt.Println(
					"  DNS Names:",
					strings.Join(
						cert.DNSNames,
						", ",
					),
				)
			}
		}
	}

	fmt.Println()
	fmt.Println("----------- SECURITY HEADERS -----------")

	printSecurityHeader(
		"Content-Security-Policy",
		result.SecurityHeaders.ContentSecurityPolicy,
	)

	printSecurityHeader(
		"Strict-Transport-Security",
		result.SecurityHeaders.StrictTransportSecurity,
	)

	printSecurityHeader(
		"X-Content-Type-Options",
		result.SecurityHeaders.XContentTypeOptions,
	)

	printSecurityHeader(
		"X-Frame-Options",
		result.SecurityHeaders.XFrameOptions,
	)

	printSecurityHeader(
		"Referrer-Policy",
		result.SecurityHeaders.ReferrerPolicy,
	)

	printSecurityHeader(
		"Permissions-Policy",
		result.SecurityHeaders.PermissionsPolicy,
	)

	printSecurityHeader(
		"Cross-Origin-Opener-Policy",
		result.SecurityHeaders.CrossOriginOpenerPolicy,
	)

	printSecurityHeader(
		"Cross-Origin-Resource-Policy",
		result.SecurityHeaders.CrossOriginResourcePolicy,
	)

	fmt.Println()
	fmt.Println("--------------- PAGE -------------------")

	if result.Title != "" {

		fmt.Println(
			"Title:",
			result.Title,
		)

	} else {

		fmt.Println(
			"Title: Not detected",
		)
	}

	fmt.Println()
	fmt.Println("----------- TECHNOLOGY HINTS -----------")

	if len(result.Technologies) == 0 {

		fmt.Println("None detected")

	} else {

		for _, tech := range result.Technologies {

			fmt.Println("-", tech)
		}
	}

	fmt.Println()
	fmt.Println("--------------- BODY -------------------")

	if result.BodyTruncated {

		fmt.Println(
			"Body: truncated",
		)

	} else {

		fmt.Println(
			"Body: complete",
		)
	}

	fmt.Println()
	fmt.Println("========================================")
}

// ============================================================
// SECURITY HEADER OUTPUT
// ============================================================

func printSecurityHeader(
	name string,
	value string,
) {

	if value == "" {

		fmt.Printf(
			"%s: MISSING\n",
			name,
		)

		return
	}

	fmt.Printf(
		"%s: %s\n",
		name,
		value,
	)
}

// ============================================================
// OUTPUT
// ============================================================

func saveResult(
	filename string,
	result HTTPResult,
) error {

	extension :=
		strings.ToLower(
			fileExtension(filename),
		)

	switch extension {

	case ".json":

		data, err :=
			json.MarshalIndent(
				result,
				"",
				"    ",
			)

		if err != nil {
			return err
		}

		return writeOutputFile(
			filename,
			data,
		)

	default:

		content :=
			formatTextResult(result)

		return writeOutputFile(
			filename,
			[]byte(content),
		)
	}
}

// ============================================================
// FILE EXTENSION
// ============================================================

func fileExtension(
	filename string,
) string {

	for i := len(filename) - 1; i >= 0; i-- {

		if filename[i] == '.' {
			return filename[i:]
		}

		if filename[i] == '/' ||
			filename[i] == '\\' {

			break
		}
	}

	return ""
}

// ============================================================
// WRITE OUTPUT
// ============================================================

func writeOutputFile(
	filename string,
	data []byte,
) error {

	dir :=
		"."

	if index :=
		strings.LastIndexAny(
			filename,
			"/\\",
		); index >= 0 {

		dir = filename[:index]

		if dir == "" {
			dir = "."
		}
	}

	if dir != "." {

		if err :=
			os.MkdirAll(
				dir,
				0755,
			); err != nil {

			return err
		}
	}

	return os.WriteFile(
		filename,
		data,
		0644,
	)
}

// ============================================================
// TEXT FORMAT
// ============================================================

func formatTextResult(
	result HTTPResult,
) string {

	var builder strings.Builder

	builder.WriteString(
		"========================================\n",
	)

	builder.WriteString(
		"           HTTP RECON RESULT\n",
	)

	builder.WriteString(
		"========================================\n",
	)

	fmt.Fprintf(
		&builder,
		"Target: %s\n",
		result.TargetURL,
	)

	fmt.Fprintf(
		&builder,
		"Final URL: %s\n",
		result.FinalURL,
	)

	fmt.Fprintf(
		&builder,
		"Method: %s\n",
		result.Method,
	)

	fmt.Fprintf(
		&builder,
		"Status: %s\n",
		result.Status,
	)

	fmt.Fprintf(
		&builder,
		"Status Code: %d\n",
		result.StatusCode,
	)

	fmt.Fprintf(
		&builder,
		"Protocol: %s\n",
		result.Protocol,
	)

	fmt.Fprintf(
		&builder,
		"Response Time: %s\n",
		result.ResponseTime,
	)

	fmt.Fprintf(
		&builder,
		"Content-Type: %s\n",
		result.ContentType,
	)

	fmt.Fprintf(
		&builder,
		"Content-Length: %d\n",
		result.ContentLength,
	)

	builder.WriteString("\n")
	builder.WriteString("--------------- URL --------------------\n")
	fmt.Fprintf(&builder, "Scheme: %s\n", result.URL.Scheme)
	fmt.Fprintf(&builder, "Host: %s\n", result.URL.Host)
	fmt.Fprintf(&builder, "Hostname: %s\n", result.URL.Hostname)
	fmt.Fprintf(&builder, "Port: %s\n", result.URL.Port)
	fmt.Fprintf(&builder, "Path: %s\n", result.URL.Path)
	if result.URL.Query != "" {
		fmt.Fprintf(&builder, "Query: %s\n", result.URL.Query)
	} else {
		builder.WriteString("Query: None\n")
	}

	builder.WriteString("\n")
	builder.WriteString("--------------- DNS --------------------\n")
	if len(result.DNS.Addresses) == 0 {
		builder.WriteString("No DNS addresses detected\n")
	} else {
		for _, address := range result.DNS.Addresses {
			fmt.Fprintf(&builder, "%s\n", address)
		}
	}

	builder.WriteString("\n")
	builder.WriteString("--------------- TIMING -----------------\n")
	fmt.Fprintf(&builder, "DNS Lookup: %s\n", result.Timing.DNSLookup)
	fmt.Fprintf(&builder, "TCP Connect: %s\n", result.Timing.TCPConnect)
	fmt.Fprintf(&builder, "TLS Handshake: %s\n", result.Timing.TLSHandshake)
	fmt.Fprintf(&builder, "TTFB: %s\n", result.Timing.TTFB)
	fmt.Fprintf(&builder, "Total: %s\n", result.Timing.Total)

	builder.WriteString("\n")
	builder.WriteString("--------------- SERVER -----------------\n")
	if result.Server != "" {
		fmt.Fprintf(&builder, "Server: %s\n", result.Server)
	} else {
		builder.WriteString("Server: Not disclosed\n")
	}
	if result.PoweredBy != "" {
		fmt.Fprintf(&builder, "X-Powered-By: %s\n", result.PoweredBy)
	} else {
		builder.WriteString("X-Powered-By: Not disclosed\n")
	}

	builder.WriteString("\n")
	builder.WriteString("----------- HTTP METHODS ----------------\n")
	if len(result.Methods.Advertised) == 0 {
		builder.WriteString("Advertised: Not disclosed\n")
		builder.WriteString("Source: No Allow or Access-Control-Allow-Methods header\n")
	} else {
		fmt.Fprintf(&builder, "Advertised: %s\n", strings.Join(result.Methods.Advertised, ", "))
		fmt.Fprintf(&builder, "Source: %s\n", result.Methods.Source)
	}

	builder.WriteString("\n")
	builder.WriteString("--------------- HEADERS ----------------\n")

	for key, values := range result.Headers {

		fmt.Fprintf(
			&builder,
			"%s: %s\n",
			key,
			strings.Join(
				values,
				", ",
			),
		)
	}

	builder.WriteString("\n")
	builder.WriteString("--------------- COOKIES ----------------\n")

	for _, cookie := range result.Cookies {

		fmt.Fprintf(
			&builder,
			"%s=%s Secure=%t HttpOnly=%t SameSite=%s\n",
			cookie.Name,
			cookie.Value,
			cookie.Secure,
			cookie.HTTPOnly,
			cookie.SameSite,
		)
	}

	builder.WriteString("\n")
	builder.WriteString("--------------- REDIRECTS --------------\n")

	for i, redirect := range result.Redirects {

		fmt.Fprintf(
			&builder,
			"%d. %s -> %s [%d %s]\n",
			i+1,
			redirect.FromURL,
			redirect.ToURL,
			redirect.StatusCode,
			redirect.Status,
		)
	}

	builder.WriteString("\n")
	builder.WriteString("--------------- TLS --------------------\n")

	if result.TLS != nil {

		fmt.Fprintf(
			&builder,
			"Version: %s\n",
			result.TLS.Version,
		)

		fmt.Fprintf(
			&builder,
			"Cipher Suite: %s\n",
			result.TLS.CipherSuite,
		)

		fmt.Fprintf(
			&builder,
			"Server Name: %s\n",
			result.TLS.ServerName,
		)

		fmt.Fprintf(
			&builder,
			"Negotiated Protocol: %s\n",
			result.TLS.Negotiated,
		)
	}

	builder.WriteString("\n")
	builder.WriteString("----------- SECURITY HEADERS -----------\n")

	writeTextSecurityHeader(
		&builder,
		"Content-Security-Policy",
		result.SecurityHeaders.ContentSecurityPolicy,
	)

	writeTextSecurityHeader(
		&builder,
		"Strict-Transport-Security",
		result.SecurityHeaders.StrictTransportSecurity,
	)

	writeTextSecurityHeader(
		&builder,
		"X-Content-Type-Options",
		result.SecurityHeaders.XContentTypeOptions,
	)

	writeTextSecurityHeader(
		&builder,
		"X-Frame-Options",
		result.SecurityHeaders.XFrameOptions,
	)

	writeTextSecurityHeader(
		&builder,
		"Referrer-Policy",
		result.SecurityHeaders.ReferrerPolicy,
	)

	writeTextSecurityHeader(
		&builder,
		"Permissions-Policy",
		result.SecurityHeaders.PermissionsPolicy,
	)

	writeTextSecurityHeader(
		&builder,
		"Cross-Origin-Opener-Policy",
		result.SecurityHeaders.CrossOriginOpenerPolicy,
	)

	writeTextSecurityHeader(
		&builder,
		"Cross-Origin-Resource-Policy",
		result.SecurityHeaders.CrossOriginResourcePolicy,
	)

	builder.WriteString("\n")
	builder.WriteString("--------------- PAGE -------------------\n")

	fmt.Fprintf(
		&builder,
		"Title: %s\n",
		result.Title,
	)

	builder.WriteString("\n")
	builder.WriteString("----------- TECHNOLOGY HINTS -----------\n")

	for _, technology := range result.Technologies {

		fmt.Fprintf(
			&builder,
			"- %s\n",
			technology,
		)
	}

	builder.WriteString("\n")
	builder.WriteString("--------------- BODY -------------------\n")

	if result.BodyTruncated {
		builder.WriteString(
			"Body: truncated\n",
		)
	} else {
		builder.WriteString(
			"Body: complete\n",
		)
	}

	return builder.String()
}

// ============================================================
// TEXT SECURITY HEADER
// ============================================================

func writeTextSecurityHeader(
	builder *strings.Builder,
	name string,
	value string,
) {

	if value == "" {

		fmt.Fprintf(
			builder,
			"%s: MISSING\n",
			name,
		)

		return
	}

	fmt.Fprintf(
		builder,
		"%s: %s\n",
		name,
		value,
	)
}

// ============================================================
// HEADER PARSER
// ============================================================

func parseHeader(
	input string,
) (string, string, error) {

	parts :=
		strings.SplitN(
			input,
			":",
			2,
		)

	if len(parts) != 2 {

		return "",
			"",
			fmt.Errorf(
				"expected KEY:VALUE",
			)
	}

	key :=
		strings.TrimSpace(
			parts[0],
		)

	value :=
		strings.TrimSpace(
			parts[1],
		)

	if key == "" {

		return "",
			"",
			fmt.Errorf(
				"header name is empty",
			)
	}

	if value == "" {

		return "",
			"",
			fmt.Errorf(
				"header value is empty",
			)
	}

	return key, value, nil
}

// ============================================================
// BODY LIMIT
// ============================================================

func parseBodyLimit(
	input string,
) (int64, error) {

	value :=
		strings.TrimSpace(
			strings.ToUpper(input),
		)

	if value == "" {
		return 0, fmt.Errorf(
			"body limit is empty",
		)
	}

	var multiplier int64 = 1
	var number string

	switch {

	case strings.HasSuffix(
		value,
		"KB",
	):

		multiplier = 1024

		number =
			strings.TrimSuffix(
				value,
				"KB",
			)

	case strings.HasSuffix(
		value,
		"MB",
	):

		multiplier =
			1024 * 1024

		number =
			strings.TrimSuffix(
				value,
				"MB",
			)

	case strings.HasSuffix(
		value,
		"GB",
	):

		multiplier =
			1024 * 1024 * 1024

		number =
			strings.TrimSuffix(
				value,
				"GB",
			)

	case strings.HasSuffix(
		value,
		"TB",
	):

		multiplier =
			1024 * 1024 *
				1024 * 1024

		number =
			strings.TrimSuffix(
				value,
				"TB",
			)

	case strings.HasSuffix(
		value,
		"B",
	):

		number =
			strings.TrimSuffix(
				value,
				"B",
			)

	default:

		return 0, fmt.Errorf(
			"unsupported size unit; use B, KB, MB, GB, or TB",
		)
	}

	number =
		strings.TrimSpace(number)

	if number == "" {

		return 0, fmt.Errorf(
			"size value is missing",
		)
	}

	parsed, err :=
		strconv.ParseInt(
			number,
			10,
			64,
		)

	if err != nil {

		return 0, fmt.Errorf(
			"invalid size value: %s",
			number,
		)
	}

	if parsed <= 0 {

		return 0, fmt.Errorf(
			"body limit must be greater than zero",
		)
	}

	if parsed > math.MaxInt64/multiplier {
		return 0, fmt.Errorf("body limit is too large")
	}

	return parsed * multiplier, nil
}
