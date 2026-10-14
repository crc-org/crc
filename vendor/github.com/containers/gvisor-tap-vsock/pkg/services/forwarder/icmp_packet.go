package forwarder

import (
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
	netIcmp "golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"gvisor.dev/gvisor/pkg/buffer"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/checksum"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

// echoRequestDetails contains the extracted details from an ICMP echo request.
type echoRequestDetails struct {
	ident   uint16
	seq     uint16
	payload []byte
	srcAddr tcpip.Address
	dataBuf buffer.Buffer
}

// handlePingRequest handles forwarding an ICMP echo request (PING) from the VM
// to the external network and injecting the reply back into the VM.
func handlePingRequest(s *stack.Stack, r *ICMPForwarderRequest, destAddr tcpip.Address, icmpHeader header.ICMPv4, pkt *stack.PacketBuffer) {
	defer pkt.DecRef()

	// Extract ICMP echo request details
	details, err := extractEchoRequestDetails(r, icmpHeader, pkt)
	if err != nil {
		return
	}
	defer details.dataBuf.Release()

	// Create ICMP connection
	conn, err := createICMPConnection()
	if err != nil {
		return
	}
	defer conn.Close()

	// Send the echo request
	if err := sendEchoRequest(conn, destAddr, details.ident, details.seq, details.payload); err != nil {
		return
	}

	// Receive and parse the echo reply
	echoReply, err := receiveEchoReply(conn)
	if err != nil {
		return
	}

	// Validate the reply matches our request (on Linux, kernel uses socket port as echo ID)
	expectedIdent := getExpectedReplyIdent(conn, details.ident)
	if !validateEchoReply(echoReply, expectedIdent, details.seq) {
		return
	}

	// Forward the reply back to the VM's network stack. Use the VM's original
	// ident/seq so the VM's ping process can match the reply to its request
	// (on Linux the host reply has kernel-assigned ID; we must rewrite to VM's).
	forwardEchoReply(s, r, details.srcAddr, destAddr, details.ident, details.seq, echoReply.Data)
}

// extractEchoRequestDetails extracts the identifier, sequence, payload, and source address
// from an ICMP echo request packet.
func extractEchoRequestDetails(r *ICMPForwarderRequest, icmpHeader header.ICMPv4, pkt *stack.PacketBuffer) (*echoRequestDetails, error) {
	ident := icmpHeader.Ident()
	seq := icmpHeader.Sequence()

	// Extract payload data
	dataBuf := pkt.Data().ToBuffer()
	dataSize := int(dataBuf.Size())
	payload := make([]byte, dataSize)
	if dataSize > 0 {
		_, _ = dataBuf.ReadAt(payload, 0)
	}

	// Get source address from the request
	srcAddr := r.ID().RemoteAddress

	return &echoRequestDetails{
		ident:   ident,
		seq:     seq,
		payload: payload,
		srcAddr: srcAddr,
		dataBuf: dataBuf,
	}, nil
}

// sendEchoRequest creates and sends an ICMP echo request message.
func sendEchoRequest(conn *netIcmp.PacketConn, destAddr tcpip.Address, ident, seq uint16, payload []byte) error {
	// Create ICMP echo request message
	msg := &netIcmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &netIcmp.Echo{
			ID:   int(ident),
			Seq:  int(seq),
			Data: payload,
		},
	}

	// Marshal the message
	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		log.Debugf("Failed to marshal ICMP message: %v", err)
		return err
	}

	// Parse destination address
	dstIP := net.ParseIP(destAddr.String())
	if dstIP == nil {
		log.Debugf("Failed to parse destination address: %s", destAddr)
		return fmt.Errorf("failed to parse destination address: %s", destAddr)
	}

	// Create destination address based on platform
	dst := createDestinationAddr(dstIP)

	// Send the ping request
	_, err = conn.WriteTo(msgBytes, dst)
	if err != nil {
		log.Debugf("Failed to send ICMP echo request: %v", err)
		return err
	}

	return nil
}

// receiveEchoReply reads and parses an ICMP echo reply from the connection.
func receiveEchoReply(conn *netIcmp.PacketConn) (*netIcmp.Echo, error) {
	// Read the reply
	replyBytes := make([]byte, 1500)
	n, _, err := conn.ReadFrom(replyBytes)
	if err != nil {
		log.Debugf("Failed to receive ICMP echo reply: %v", err)
		return nil, err
	}

	// Extract ICMP data (skip IP header on Windows)
	replyData, err := extractICMPData(replyBytes[:n])
	if err != nil {
		return nil, err
	}

	// Parse the reply
	replyMsg, err := netIcmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), replyData)
	if err != nil {
		log.Debugf("Failed to parse ICMP reply: %v", err)
		return nil, err
	}

	// Check if it's an echo reply
	if replyMsg.Type != ipv4.ICMPTypeEchoReply {
		log.Debugf("Received ICMP message type %v, expected Echo Reply", replyMsg.Type)
		return nil, fmt.Errorf("unexpected ICMP message type: %v", replyMsg.Type)
	}

	echoReply, ok := replyMsg.Body.(*netIcmp.Echo)
	if !ok {
		log.Debugf("ICMP reply body is not an Echo")
		return nil, fmt.Errorf("ICMP reply body is not an Echo")
	}

	return echoReply, nil
}

// validateEchoReply verifies that the echo reply matches the original request.
func validateEchoReply(echoReply *netIcmp.Echo, expectedIdent, expectedSeq uint16) bool {
	if echoReply.ID != int(expectedIdent) || echoReply.Seq != int(expectedSeq) {
		log.Debugf("ICMP reply ID/Seq mismatch: got ID=%d Seq=%d, expected ID=%d Seq=%d",
			echoReply.ID, echoReply.Seq, expectedIdent, expectedSeq)
		return false
	}
	return true
}

// forwardEchoReply creates an ICMP echo reply packet and forwards it back to the VM.
func forwardEchoReply(s *stack.Stack, r *ICMPForwarderRequest, dstAddr tcpip.Address, srcAddr tcpip.Address, ident, seq uint16, data []byte) {
	// Create ICMP echo reply header
	icmpHeaderSize := header.ICMPv4MinimumSize
	icmpBuf := make([]byte, icmpHeaderSize+len(data))
	icmpHdr := header.ICMPv4(icmpBuf)

	icmpHdr.SetType(header.ICMPv4EchoReply)
	icmpHdr.SetCode(0)
	icmpHdr.SetIdent(ident)
	icmpHdr.SetSequence(seq)

	// Copy data
	if len(data) > 0 {
		copy(icmpBuf[icmpHeaderSize:], data)
	}

	// Calculate checksum
	icmpHdr.SetChecksum(0)
	icmpHdr.SetChecksum(^checksum.Checksum(icmpBuf, 0))

	// Get the original packet's network info
	origPkt := r.Packet()
	if origPkt == nil {
		return
	}

	// Find route to send the reply back
	// Use srcAddr (the address we pinged) as the local address so the reply appears
	// to come from the address the VM originally pinged, not from the gateway
	route, err := s.FindRoute(origPkt.NICID, srcAddr, dstAddr, header.IPv4ProtocolNumber, false)
	if err != nil {
		log.Debugf("Failed to find route for ICMP reply: %v", err)
		return
	}
	defer route.Release()

	// Create packet buffer with ICMP reply
	payload := buffer.MakeWithData(icmpBuf)
	pkt := stack.NewPacketBuffer(stack.PacketBufferOptions{
		ReserveHeaderBytes: int(route.MaxHeaderLength()),
		Payload:            payload,
	})
	defer pkt.DecRef()

	pkt.NetworkProtocolNumber = header.IPv4ProtocolNumber
	pkt.TransportProtocolNumber = header.ICMPv4ProtocolNumber

	// Write the packet
	params := stack.NetworkHeaderParams{
		Protocol: header.ICMPv4ProtocolNumber,
		TTL:      64,
		TOS:      0,
	}
	if err := route.WritePacket(params, pkt); err != nil {
		log.Debugf("Failed to forward ICMP echo reply: %v", err)
		return
	}

	log.Debugf("Successfully forwarded ICMP echo reply to %s", dstAddr)
}
