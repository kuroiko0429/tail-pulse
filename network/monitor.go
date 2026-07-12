package network

import (
	"fmt"
	"net"
	"os/exec"
	"time"
)

type NodeNetInfo struct {
	Latency     float64
	LatencyHist []float64
	OpenPorts   map[int]bool
}

var commonPorts = []int{22, 80, 443, 3389, 5900}

func CheckPorts(ip string) map[int]bool {
	results := make(map[int]bool)
	timeout := 500 * time.Millisecond
	for _, port := range commonPorts {
		target := fmt.Sprintf("%s:%d", ip, port)
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

func Ping(ip string) float64 {
	start := time.Now()
	cmd := exec.Command("ping", "-c", "1", "-W", "1", ip)
	err := cmd.Run()
	if err != nil {
		return 0
	}
	return float64(time.Since(start).Milliseconds())
}

// WakeOnLan sends a magic packet to a MAC address via a broadcast IP
// For Tailscale, this might involve discovering the MAC, which is complex.
// We'll provide a stub or basic UDP broadcast string.
func WakeOnLan(macAddr string, broadcastIP string) error {
	// Simple stub for WoL, in real scenarios requires net.Dial("udp", broadcast)
	return fmt.Errorf("WoL not fully implemented for tailscale subnets yet")
}
