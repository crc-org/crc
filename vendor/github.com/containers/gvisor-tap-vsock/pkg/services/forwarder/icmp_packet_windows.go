//go:build windows

package forwarder

import (
	"fmt"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
	netIcmp "golang.org/x/net/icmp"
)

// createICMPConnection creates an ICMP connection using privileged raw sockets (ip4:icmp) on Windows.
func createICMPConnection() (*netIcmp.PacketConn, error) {
	conn, err := netIcmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		log.Debugf("Failed to create ICMP connection: %v", err)
		return nil, err
	}

	// Set read deadline
	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		conn.Close()
		log.Debugf("Failed to set read deadline: %v", err)
		return nil, err
	}

	return conn, nil
}

// createDestinationAddr creates a destination address for Windows raw sockets.
func createDestinationAddr(dstIP net.IP) net.Addr {
	// Windows requires net.IPAddr for raw sockets
	return &net.IPAddr{IP: dstIP}
}

// extractICMPData extracts ICMP data from the received bytes.
// Some Windows versions return IPv4 with the IP header; others return ICMP only (e.g. echo
// reply type 0 makes the first byte 0x00, which is not IPv4 version 4). If the buffer
// starts with an IPv4 header, strip it; otherwise return the payload as-is.
func extractICMPData(replyBytes []byte) ([]byte, error) {
	if len(replyBytes) == 0 {
		return nil, fmt.Errorf("reply packet empty")
	}
	version := (replyBytes[0] >> 4) & 0x0F
	if version != 4 {
		return replyBytes, nil
	}
	if len(replyBytes) < 20 {
		return nil, fmt.Errorf("reply packet too short for IPv4: %d bytes", len(replyBytes))
	}
	ihl := int(replyBytes[0]&0x0F) * 4
	if ihl < 20 || ihl > len(replyBytes) {
		return nil, fmt.Errorf("invalid IP header length: %d", ihl)
	}
	return replyBytes[ihl:], nil
}

// getExpectedReplyIdent returns the ICMP echo identifier to expect in the reply.
// On Windows (raw sockets) the kernel preserves the ID we send.
func getExpectedReplyIdent(_ *netIcmp.PacketConn, sentIdent uint16) uint16 {
	return sentIdent
}
