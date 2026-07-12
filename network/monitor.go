package network

import (
	"encoding/hex"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"
)

type NodeNetInfo struct {
	Latency     float64
	LatencyHist []float64
	OpenPorts   map[int]bool
	Path        string // "Direct", "DERP(...)", or "" if unknown
}

var commonPorts = []int{22, 80, 443, 3389, 5900}

func CheckPorts(ip string) map[int]bool {
	results := make(map[int]bool)
	timeout := 500 * time.Millisecond
	for _, port := range commonPorts {
		target := net.JoinHostPort(ip, fmt.Sprintf("%d", port))
		conn, err := net.DialTimeout("tcp", target, timeout)
		if err == nil {
			results[port] = true
			conn.Close()
		} else {
			results[port] = false
		}
	}
	return results
}

// WakeOnLan sends a magic packet to the given MAC address as a UDP broadcast
// on the local network.
func WakeOnLan(macAddr string) error {
	macAddr = strings.ReplaceAll(macAddr, ":", "")
	macAddr = strings.ReplaceAll(macAddr, "-", "")
	if len(macAddr) != 12 {
		return fmt.Errorf("invalid MAC address format")
	}

	macBytes, err := hex.DecodeString(macAddr)
	if err != nil {
		return err
	}

	packet := make([]byte, 0, 102)
	for i := 0; i < 6; i++ {
		packet = append(packet, 0xff)
	}
	for i := 0; i < 16; i++ {
		packet = append(packet, macBytes...)
	}

	addr, err := net.ResolveUDPAddr("udp", "255.255.255.255:9")
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(packet)
	return err
}

// WakeOnLanViaProxy sends the magic packet from a remote Tailscale node over
// SSH, for when the target device is on a different LAN than this machine.
func WakeOnLanViaProxy(proxyHost, macAddr string) error {
	pyCode := fmt.Sprintf(
		`import socket; socket.socket(socket.AF_INET, socket.SOCK_DGRAM).sendto(b'\xff'*6 + bytes.fromhex('%s'.replace(':', '').replace('-', ''))*16, ('255.255.255.255', 9))`,
		macAddr,
	)
	cmd := exec.Command("ssh", "-o", "ConnectTimeout=3", proxyHost, "python3", "-c", pyCode)
	return cmd.Run()
}
