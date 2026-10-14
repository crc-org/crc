//go:build !windows

package forwarder

import (
	"net"
	"runtime"
	"time"

	log "github.com/sirupsen/logrus"
	netIcmp "golang.org/x/net/icmp"
)

// createICMPConnection creates an ICMP connection using unprivileged ICMP sockets (udp4) on Linux/macOS.
func createICMPConnection() (*netIcmp.PacketConn, error) {
	conn, err := netIcmp.ListenPacket("udp4", "0.0.0.0")
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

// createDestinationAddr creates a destination address for Unix unprivileged sockets.
func createDestinationAddr(dstIP net.IP) net.Addr {
	// Linux/macOS use net.UDPAddr for unprivileged sockets
	return &net.UDPAddr{IP: dstIP, Port: 0}
}

// extractICMPData extracts ICMP data from the received bytes.
// On Linux/macOS unprivileged sockets, it returns the data as-is.
func extractICMPData(replyBytes []byte) ([]byte, error) {
	// Linux/macOS unprivileged sockets return just the ICMP data
	return replyBytes, nil
}

// getExpectedReplyIdent returns the ICMP echo identifier to expect in the reply.
// On Linux, the kernel overwrites the echo ID with the socket's local port for
// unprivileged ICMP sockets, so we must use that for validation. On macOS the
// kernel preserves the ID we send.
func getExpectedReplyIdent(conn *netIcmp.PacketConn, sentIdent uint16) uint16 {
	if runtime.GOOS != "linux" {
		return sentIdent
	}
	addr := conn.LocalAddr()
	udpAddr, ok := addr.(*net.UDPAddr)
	if !ok || udpAddr == nil {
		return sentIdent
	}
	port := udpAddr.Port
	if port < 0 || port > 0xFFFF {
		return sentIdent
	}
	return uint16(port)
}
