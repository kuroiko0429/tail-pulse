package main

import (
	"encoding/hex"
	"fmt"
	"net"
	"strings"
)

func sendMagicPacket(macAddr string) error {
	macAddr = strings.ReplaceAll(macAddr, ":", "")
	macAddr = strings.ReplaceAll(macAddr, "-", "")
	if len(macAddr) != 12 {
		return fmt.Errorf("invalid MAC address format")
	}

	macBytes, err := hex.DecodeString(macAddr)
	if err != nil {
		return err
	}

	var packet []byte
	for i := 0; i < 6; i++ {
		packet = append(packet, 0xff)
	}
	for i := 0; i < 16; i++ {
		packet = append(packet, macBytes...)
	}

	// UDP broadcast
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
