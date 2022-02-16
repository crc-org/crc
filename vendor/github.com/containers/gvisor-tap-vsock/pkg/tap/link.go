package tap

import (
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	log "github.com/sirupsen/logrus"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/buffer"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

type LinkEndpoint struct {
	debug      bool
	mtu        int
	mac        tcpip.LinkAddress
	ip         string
	virtualIPs map[string]struct{}

	dispatcher    stack.NetworkDispatcher
	networkSwitch NetworkSwitch
}

func NewLinkEndpoint(debug bool, mtu int, macAddress string, ip string, virtualIPs []string) (*LinkEndpoint, error) {
	linkAddr, err := net.ParseMAC(macAddress)
	if err != nil {
		return nil, err
	}
	set := make(map[string]struct{})
	for _, virtualIP := range virtualIPs {
		set[virtualIP] = struct{}{}
	}
	return &LinkEndpoint{
		debug:      debug,
		mtu:        mtu,
		mac:        tcpip.LinkAddress(linkAddr),
		ip:         ip,
		virtualIPs: set,
	}, nil
}

func (e *LinkEndpoint) ARPHardwareType() header.ARPHardwareType {
	return header.ARPHardwareEther
}

func (e *LinkEndpoint) Connect(networkSwitch NetworkSwitch) {
	e.networkSwitch = networkSwitch
}

func (e *LinkEndpoint) Attach(dispatcher stack.NetworkDispatcher) {
	e.dispatcher = dispatcher
}

func (e *LinkEndpoint) IsAttached() bool {
	return e.dispatcher != nil
}

func (e *LinkEndpoint) DeliverNetworkPacket(remote, local tcpip.LinkAddress, protocol tcpip.NetworkProtocolNumber, pkt *stack.PacketBuffer) {
	e.dispatcher.DeliverNetworkPacket(remote, local, protocol, pkt)
}

func (e *LinkEndpoint) AddHeader(local, remote tcpip.LinkAddress, protocol tcpip.NetworkProtocolNumber, pkt *stack.PacketBuffer) {
}

func (e *LinkEndpoint) Capabilities() stack.LinkEndpointCapabilities {
	return stack.CapabilityResolutionRequired | stack.CapabilityRXChecksumOffload
}

func (e *LinkEndpoint) LinkAddress() tcpip.LinkAddress {
	return e.mac
}

func (e *LinkEndpoint) MaxHeaderLength() uint16 {
	return uint16(header.EthernetMinimumSize)
}

func (e *LinkEndpoint) MTU() uint32 {
	return uint32(e.mtu)
}

func (e *LinkEndpoint) Wait() {
}

func (e *LinkEndpoint) WritePackets(pkts stack.PacketBufferList) (int, tcpip.Error) {
	n := 0
	for p := pkts.Front(); p != nil; p = p.Next() {
		if err := e.writePacket(p.EgressRoute, p.NetworkProtocolNumber, p); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

func (e *LinkEndpoint) writePacket(r stack.RouteInfo, protocol tcpip.NetworkProtocolNumber, pkt *stack.PacketBuffer) tcpip.Error {
	// Preserve the src address if it's set in the route.
	srcAddr := e.LinkAddress()
	if r.LocalLinkAddress != "" {
		srcAddr = r.LocalLinkAddress
	}
	eth := header.Ethernet(pkt.LinkHeader().Push(header.EthernetMinimumSize))
	eth.Encode(&header.EthernetFields{
		Type:    protocol,
		SrcAddr: srcAddr,
		DstAddr: r.RemoteLinkAddress,
	})

	h := header.ARP(pkt.NetworkHeader().View())
	if h.IsValid() &&
		h.Op() == header.ARPReply {
		ip := tcpip.Address(h.ProtocolAddressSender()).String()
		_, ok := e.virtualIPs[ip]
		if ip != e.IP() && !ok {
			log.Debugf("dropping spoofing packets from the gateway about IP %s", ip)
			return nil
		}
	}

	if e.debug {
		vv := buffer.NewVectorisedView(pkt.Size(), pkt.Views())
		packet := gopacket.NewPacket(vv.ToView(), layers.LayerTypeEthernet, gopacket.Default)
		log.Info(packet.String())
	}

	e.networkSwitch.DeliverNetworkPacket(r.RemoteLinkAddress, srcAddr, protocol, pkt)
	return nil
}

func (e *LinkEndpoint) WriteRawPacket(pkt *stack.PacketBuffer) tcpip.Error {
	return &tcpip.ErrNotSupported{}
}

func (e *LinkEndpoint) IP() string {
	return e.ip
}
