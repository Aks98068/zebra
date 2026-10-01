package ping

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type PingConfig struct {
	Target  string
	Count   int
	Timeout time.Duration
}

type PingResult struct {
	Target      string
	Address     string
	Transmitted int
	Received    int
	Loss        float64
	MinRTT      time.Duration
	AvgRTT      time.Duration
	MaxRTT      time.Duration
	Replies     []PingReply
}

type PingReply struct {
	Sequence int
	Address  string
	RTT      time.Duration
	TTL      int
}

func PingRecon(args []string) bool {
	config, err := ParsePingArguments(args)
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}

	result, err := RunPing(config)
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}

	PrintPingResult(result)

	return true
}

func ParsePingArguments(args []string) (PingConfig, error) {
	config := PingConfig{
		Count:   4,
		Timeout: 2 * time.Second,
	}

	if len(args) == 0 {
		return PingConfig{}, fmt.Errorf(
			"target must be specified",
		)
	}

	if strings.HasPrefix(args[0], "-") {
		return PingConfig{}, fmt.Errorf(
			"target must be specified before options",
		)
	}

	config.Target = strings.TrimSpace(args[0])

	if config.Target == "" {
		return PingConfig{}, fmt.Errorf(
			"target cannot be empty",
		)
	}

	for i := 1; i < len(args); i++ {
		arg := strings.TrimSpace(args[i])

		switch arg {
		case "-c":
			if i+1 >= len(args) {
				return PingConfig{}, fmt.Errorf(
					"missing value after -c",
				)
			}

			i++

			count, err := strconv.Atoi(
				strings.TrimSpace(args[i]),
			)

			if err != nil || count <= 0 {
				return PingConfig{}, fmt.Errorf(
					"invalid ping count: %q",
					args[i],
				)
			}

			if count > 10000 {
				return PingConfig{}, fmt.Errorf(
					"ping count cannot exceed 10000",
				)
			}

			config.Count = count

		case "-W", "--timeout":
			if i+1 >= len(args) {
				return PingConfig{}, fmt.Errorf(
					"missing timeout value after %s",
					arg,
				)
			}

			i++

			timeout, err := time.ParseDuration(
				strings.TrimSpace(args[i]),
			)

			if err != nil || timeout <= 0 {
				return PingConfig{}, fmt.Errorf(
					"invalid timeout: %q",
					args[i],
				)
			}

			config.Timeout = timeout

		case "-h", "--help":
			return PingConfig{}, fmt.Errorf(
				"help requested",
			)

		default:
			return PingConfig{}, fmt.Errorf(
				"unknown option: %q",
				arg,
			)
		}
	}

	return config, nil
}

func RunPing(config PingConfig) (PingResult, error) {
	address, err := resolveTarget(config.Target)
	if err != nil {
		return PingResult{}, err
	}

	ip := net.ParseIP(address)

	if ip == nil {
		return PingResult{}, fmt.Errorf(
			"invalid resolved address: %s",
			address,
		)
	}

	result := PingResult{
		Target:  config.Target,
		Address: address,
		Replies: make([]PingReply, 0, config.Count),
	}

	if ip.To4() != nil {
		return runIPv4Ping(
			config,
			address,
			result,
		)
	}

	return runIPv6Ping(
		config,
		address,
		result,
	)
}

func resolveTarget(target string) (string, error) {
	if ip := net.ParseIP(target); ip != nil {
		return ip.String(), nil
	}

	ips, err := net.LookupIP(target)
	if err != nil {
		return "", fmt.Errorf(
			"failed to resolve %q: %w",
			target,
			err,
		)
	}

	if len(ips) == 0 {
		return "", fmt.Errorf(
			"no IP address found for %q",
			target,
		)
	}

	for _, ip := range ips {
		if ip.To4() != nil {
			return ip.To4().String(), nil
		}
	}

	return ips[0].String(), nil
}

func runIPv4Ping(
	config PingConfig,
	address string,
	result PingResult,
) (PingResult, error) {
	conn, err := icmp.ListenPacket(
		"ip4:icmp",
		"0.0.0.0",
	)

	if err != nil {
		return result, fmt.Errorf(
			"failed to create ICMP socket: %w",
			err,
		)
	}

	defer conn.Close()

	destination := &net.IPAddr{
		IP: net.ParseIP(address),
	}

	for sequence := 1; sequence <= config.Count; sequence++ {
		result.Transmitted++

		start := time.Now()

		message := icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &icmp.Echo{
				ID:   int(time.Now().UnixNano() & 0xffff),
				Seq:  sequence,
				Data: []byte("ZEBRA"),
			},
		}

		packet, err := message.Marshal(nil)
		if err != nil {
			continue
		}

		if err := conn.SetReadDeadline(
			time.Now().Add(config.Timeout),
		); err != nil {
			return result, err
		}

		if _, err := conn.WriteTo(
			packet,
			destination,
		); err != nil {
			continue
		}

		reply, err := readIPv4Reply(
			conn,
			sequence,
			config.Timeout,
			start,
		)

		if err != nil {
			fmt.Printf(
				"Request timeout for icmp_seq %d\n",
				sequence,
			)

			continue
		}

		result.Received++

		reply.RTT = time.Since(start)

		result.Replies = append(
			result.Replies,
			reply,
		)

		fmt.Printf(
			"%d bytes from %s: icmp_seq=%d time=%s",
			5,
			reply.Address,
			reply.Sequence,
			reply.RTT.Round(time.Microsecond),
		)

		if reply.TTL > 0 {
			fmt.Printf(
				" ttl=%d",
				reply.TTL,
			)
		}

		fmt.Println()

		if sequence < config.Count {
			time.Sleep(1 * time.Second)
		}
	}

	finalizeResult(&result)

	return result, nil
}

func readIPv4Reply(
	conn net.PacketConn,
	sequence int,
	timeout time.Duration,
	start time.Time,
) (PingReply, error) {
	buffer := make([]byte, 1500)

	for time.Since(start) < timeout {
		n, peer, err := conn.ReadFrom(
			buffer,
		)

		if err != nil {
			return PingReply{}, err
		}

		message, err := icmp.ParseMessage(
			1,
			buffer[:n],
		)

		if err != nil {
			continue
		}

		if message.Type != ipv4.ICMPTypeEchoReply {
			continue
		}

		echo, ok := message.Body.(*icmp.Echo)
		if !ok {
			continue
		}

		if echo.Seq != sequence {
			continue
		}

		address := peer.String()

		if host, _, err := net.SplitHostPort(
			address,
		); err == nil {
			address = host
		}

		return PingReply{
			Sequence: sequence,
			Address:  address,
			TTL:      extractTTL(message),
		}, nil
	}

	return PingReply{}, fmt.Errorf(
		"timeout waiting for reply",
	)
}

func runIPv6Ping(
	config PingConfig,
	address string,
	result PingResult,
) (PingResult, error) {
	conn, err := icmp.ListenPacket(
		"udp6",
		"[::]:0",
	)

	if err != nil {
		return result, fmt.Errorf(
			"failed to create IPv6 ICMP socket: %w",
			err,
		)
	}

	defer conn.Close()

	destination := &net.UDPAddr{
		IP: net.ParseIP(address),
	}

	for sequence := 1; sequence <= config.Count; sequence++ {
		result.Transmitted++

		start := time.Now()

		message := icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &icmp.Echo{
				ID:   int(time.Now().UnixNano() & 0xffff),
				Seq:  sequence,
				Data: []byte("ZEBRA"),
			},
		}

		packet, err := message.Marshal(nil)
		if err != nil {
			continue
		}

		if err := conn.SetReadDeadline(
			time.Now().Add(config.Timeout),
		); err != nil {
			return result, err
		}

		if _, err := conn.WriteTo(
			packet,
			destination,
		); err != nil {
			continue
		}

		buffer := make([]byte, 1500)

		n, peer, err := conn.ReadFrom(
			buffer,
		)

		if err != nil {
			fmt.Printf(
				"Request timeout for icmp_seq %d\n",
				sequence,
			)

			continue
		}

		reply, err := icmp.ParseMessage(
			58,
			buffer[:n],
		)

		if err != nil {
			continue
		}

		if reply.Type != ipv4.ICMPTypeEchoReply {
			continue
		}

		result.Received++

		rtt := time.Since(start)

		peerAddress := peer.String()

		result.Replies = append(
			result.Replies,
			PingReply{
				Sequence: sequence,
				Address:  peerAddress,
				RTT:      rtt,
			},
		)

		fmt.Printf(
			"%d bytes from %s: icmp_seq=%d time=%s\n",
			5,
			peerAddress,
			sequence,
			rtt.Round(time.Microsecond),
		)

		if sequence < config.Count {
			time.Sleep(1 * time.Second)
		}
	}

	finalizeResult(&result)

	return result, nil
}

func extractTTL(message *icmp.Message) int {
	_ = message
	return 0
}

func finalizeResult(result *PingResult) {
	if result.Transmitted == 0 {
		return
	}

	result.Loss = float64(
		result.Transmitted-result.Received,
	) / float64(result.Transmitted) * 100

	if len(result.Replies) == 0 {
		return
	}

	var total time.Duration

	result.MinRTT = result.Replies[0].RTT
	result.MaxRTT = result.Replies[0].RTT

	for _, reply := range result.Replies {
		total += reply.RTT

		if reply.RTT < result.MinRTT {
			result.MinRTT = reply.RTT
		}

		if reply.RTT > result.MaxRTT {
			result.MaxRTT = reply.RTT
		}
	}

	result.AvgRTT = total / time.Duration(
		len(result.Replies),
	)
}

func PrintPingResult(result PingResult) {
	fmt.Println()

	fmt.Printf(
		"PING %s (%s)\n",
		result.Target,
		result.Address,
	)

	fmt.Println()

	fmt.Printf(
		"--- %s ping statistics ---\n",
		result.Target,
	)

	fmt.Printf(
		"%d packets transmitted, %d packets received, %.1f%% packet loss\n",
		result.Transmitted,
		result.Received,
		result.Loss,
	)

	if result.Received > 0 {
		fmt.Printf(
			"rtt min/avg/max = %s/%s/%s\n",
			result.MinRTT.Round(time.Microsecond),
			result.AvgRTT.Round(time.Microsecond),
			result.MaxRTT.Round(time.Microsecond),
		)
	}

	fmt.Println()
}
