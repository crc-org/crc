package forwarder

import (
	"sync"

	log "github.com/sirupsen/logrus"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

type Forwarder struct {
	handler func(request *ICMPForwarderRequest)
	stack   *stack.Stack
}

func ICMP(s *stack.Stack, nat map[tcpip.Address]tcpip.Address, natLock *sync.Mutex) *Forwarder {
	return NewForwarder(s, func(r *ICMPForwarderRequest) {
		localAddress := r.ID().LocalAddress

		// Skip forwarding for addresses that should be handled locally
		if header.IsV4LoopbackAddress(localAddress) || localAddress == header.IPv4Broadcast {
			return
		}

		// Apply NAT translation if needed
		natLock.Lock()
		if replaced, ok := nat[localAddress]; ok {
			localAddress = replaced
		}
		natLock.Unlock()

		pkt := r.Packet()
		if pkt == nil {
			log.Warningf("Dropping ICMP packet from VM (no packet data)")
			return
		}

		// Check if this is an ICMP Echo Request (PING)
		transportHeader := pkt.TransportHeader().Slice()
		if len(transportHeader) < header.ICMPv4MinimumSize {
			log.Warningf("Dropping ICMP packet from VM (packet too short)")
			return
		}

		icmpHeader := header.ICMPv4(transportHeader)
		if icmpHeader.Type() != header.ICMPv4Echo {
			// Not a PING, drop it
			log.Warningf("Dropping ICMP packet from VM (type %d, not Echo Request)", icmpHeader.Type())
			return
		}

		// This is a PING request - forward it using unprivileged ICMP sockets
		go handlePingRequest(s, r, localAddress, icmpHeader, pkt)
	})
}

// HandlePacket handles all packets.
//
// This function is expected to be passed as an argument to the
// stack.SetTransportProtocolHandler function.
func (f *Forwarder) HandlePacket(id stack.TransportEndpointID, pkt *stack.PacketBuffer) bool {
	f.handler(NewICMPForwarderRequest(f.stack, id, pkt.IncRef()))
	return true
}

// NewForwarder allocates and initializes a new forwarder.
func NewForwarder(s *stack.Stack, handler func(*ICMPForwarderRequest)) *Forwarder {
	return &Forwarder{
		stack:   s,
		handler: handler,
	}
}
