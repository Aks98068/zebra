package networkscan

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

func HTTP(
	target string,
	port int,
	timeout time.Duration,
) (string, string) {
	address := net.JoinHostPort(
		target,
		strconv.Itoa(port),
	)

	conn, err := net.DialTimeout(
		"tcp",
		address,
		timeout,
	)

	if err != nil {
		return "", ""
	}

	defer conn.Close()

	_ = conn.SetDeadline(
		time.Now().Add(timeout),
	)

	request := fmt.Sprintf(
		"HEAD / HTTP/1.1\r\nHost: %s\r\nConnection: close\r\nUser-Agent: ZEBRA/1.0\r\n\r\n",
		target,
	)

	if _, err := conn.Write([]byte(request)); err != nil {
		return "", ""
	}

	reader := bufio.NewReader(conn)

	var response strings.Builder

	for i := 0; i < 32; i++ {
		line, err := reader.ReadString('\n')

		if err != nil {
			break
		}

		response.WriteString(line)

		if strings.TrimSpace(line) == "" {
			break
		}
	}

	data := response.String()

	if data == "" {
		return "", ""
	}

	server := header(data, "Server")

	if server != "" {
		return "http", clean(server)
	}

	return "http", clean(data)
}

func header(
	response string,
	name string,
) string {
	for _, line := range strings.Split(
		response,
		"\n",
	) {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(
			strings.ToLower(line),
			strings.ToLower(name),
		) {
			parts := strings.SplitN(
				line,
				":",
				2,
			)

			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	return ""
}

func clean(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\n", " ")

	if len(value) > 512 {
		value = value[:512]
	}

	return strings.TrimSpace(value)
}

func HTTPS(
	target string,
	port int,
	timeout time.Duration,
) (string, string) {
	address := net.JoinHostPort(
		target,
		strconv.Itoa(port),
	)

	conn, err := tls.DialWithDialer(
		&net.Dialer{
			Timeout: timeout,
		},
		"tcp",
		address,
		&tls.Config{
			ServerName:         target,
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: true,
		},
	)

	if err != nil {
		return "", ""
	}

	defer conn.Close()

	_ = conn.SetDeadline(
		time.Now().Add(timeout),
	)

	request := fmt.Sprintf(
		"HEAD / HTTP/1.1\r\nHost: %s\r\nConnection: close\r\nUser-Agent: ZEBRA/1.0\r\n\r\n",
		target,
	)

	if _, err := conn.Write([]byte(request)); err != nil {
		return "", ""
	}

	buffer := make([]byte, 4096)

	n, err := conn.Read(buffer)

	if err != nil || n == 0 {
		return "", ""
	}

	response := strings.TrimSpace(
		string(buffer[:n]),
	)

	server := header(
		response,
		"Server",
	)

	if server != "" {
		return "https", server
	}

	return "https", response
}

func SSH(
	target string,
	port int,
	timeout time.Duration,
) (string, string) {
	address := net.JoinHostPort(
		target,
		strconv.Itoa(port),
	)

	conn, err := net.DialTimeout(
		"tcp",
		address,
		timeout,
	)

	if err != nil {
		return "", ""
	}

	defer conn.Close()

	_ = conn.SetReadDeadline(
		time.Now().Add(timeout),
	)

	reader := bufio.NewReader(conn)

	line, err := reader.ReadString('\n')

	if err != nil {
		return "", ""
	}

	banner := strings.TrimSpace(line)

	if strings.HasPrefix(
		strings.ToUpper(banner),
		"SSH-",
	) {
		return "ssh", banner
	}

	return "", banner
}

func SMTP(
	target string,
	port int,
	timeout time.Duration,
) (string, string) {
	address := net.JoinHostPort(
		target,
		strconv.Itoa(port),
	)

	conn, err := net.DialTimeout(
		"tcp",
		address,
		timeout,
	)

	if err != nil {
		return "", ""
	}

	defer conn.Close()

	_ = conn.SetReadDeadline(
		time.Now().Add(timeout),
	)

	reader := bufio.NewReader(conn)

	line, err := reader.ReadString('\n')

	if err != nil {
		return "", ""
	}

	banner := strings.TrimSpace(line)

	if strings.HasPrefix(banner, "220") ||
		strings.Contains(
			strings.ToLower(banner),
			"smtp",
		) {
		return "smtp", banner
	}

	return "", banner
}
