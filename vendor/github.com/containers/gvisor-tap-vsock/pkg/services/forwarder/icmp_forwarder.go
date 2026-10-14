package forwarder

import (
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/waiter"
)

// ICMPForwarderRequest represents a request to forward an ICMP packet.
type ICMPForwarderRequest struct {
	stack *stack.Stack
	id    stack.TransportEndpointID
	pkt   *stack.PacketBuffer
}

// NewICMPForwarderRequest creates a new ICMP forwarder request.
func NewICMPForwarderRequest(s *stack.Stack, id stack.TransportEndpointID, pkt *stack.PacketBuffer) *ICMPForwarderRequest {
	return &ICMPForwarderRequest{
		stack: s,
		id:    id,
		pkt:   pkt,
	}
}

// ID returns the 4-tuple (src address, src port, dst address, dst port) that
// represents the connection request.
func (f *ICMPForwarderRequest) ID() stack.TransportEndpointID {
	return f.id
}

// Packet returns the packet buffer associated with this forwarder request.
func (f *ICMPForwarderRequest) Packet() *stack.PacketBuffer {
	return f.pkt
}

// CreateEndpoint creates a new endpoint for this forwarder request.
func (f *ICMPForwarderRequest) CreateEndpoint(s *stack.Stack, wq *waiter.Queue) (tcpip.Endpoint, tcpip.Error) {
	// Use the stack's public NewEndpoint API instead of linkname
	return s.NewEndpoint(f.pkt.TransportProtocolNumber, f.pkt.NetworkProtocolNumber, wq)
}
