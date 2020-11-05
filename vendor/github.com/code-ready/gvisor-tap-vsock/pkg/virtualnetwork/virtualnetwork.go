package virtualnetwork

import (
	"encoding/json"
	"math"
	"net"
	"net/http"
	"os"

	"github.com/code-ready/gvisor-tap-vsock/pkg/tap"
	"github.com/code-ready/gvisor-tap-vsock/pkg/types"
	"github.com/pkg/errors"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/link/sniffer"
	"gvisor.dev/gvisor/pkg/tcpip/network/arp"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/icmp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
)

type VirtualNetwork struct {
	configuration *types.Configuration
	stack         *stack.Stack
	networkSwitch *tap.Switch
}

func New(configuration *types.Configuration) (*VirtualNetwork, error) {
	_, subnet, err := net.ParseCIDR(configuration.Subnet)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse subnet cidr")
	}

	var endpoint stack.LinkEndpoint

	tapEndpoint := tap.NewLinkEndpoint(configuration.Debug, configuration.GatewayMacAddress)
	networkSwitch := tap.NewSwitch(configuration.Debug, configuration.MTU, tap.NewIPPool(subnet))
	tapEndpoint.Connect(networkSwitch)
	networkSwitch.Connect(configuration.GatewayIP, tapEndpoint)

	if configuration.CaptureFile != "" {
		_ = os.Remove(configuration.CaptureFile)
		fd, err := os.Create(configuration.CaptureFile)
		if err != nil {
			return nil, errors.Wrap(err, "cannot create capture file")
		}
		endpoint, err = sniffer.NewWithWriter(tapEndpoint, fd, math.MaxUint32)
		if err != nil {
			return nil, errors.Wrap(err, "cannot create sniffer")
		}
	} else {
		endpoint = tapEndpoint
	}

	stack, err := createStack(configuration, endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create network stack")
	}

	if err := addServices(configuration, stack); err != nil {
		return nil, errors.Wrap(err, "cannot add network services")
	}

	return &VirtualNetwork{
		configuration: configuration,
		stack:         stack,
		networkSwitch: networkSwitch,
	}, nil
}

func (n *VirtualNetwork) BytesSent() uint64 {
	if n.networkSwitch == nil {
		return 0
	}
	return n.networkSwitch.Sent
}

func (n *VirtualNetwork) BytesReceived() uint64 {
	if n.networkSwitch == nil {
		return 0
	}
	return n.networkSwitch.Received
}

func (n *VirtualNetwork) Mux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]uint64{n.networkSwitch.Sent, n.networkSwitch.Received})
	})
	mux.HandleFunc("/cam", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(n.networkSwitch.CAM())
	})
	mux.HandleFunc("/leases", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(n.networkSwitch.IPs.Leases())
	})
	mux.HandleFunc("/connect", func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
			return
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		n.networkSwitch.Accept(conn)
	})
	return mux
}

func createStack(configuration *types.Configuration, endpoint stack.LinkEndpoint) (*stack.Stack, error) {
	s := stack.New(stack.Options{
		UseNeighborCache: true,
		NetworkProtocols: []stack.NetworkProtocolFactory{
			ipv4.NewProtocol,
			arp.NewProtocol,
		},
		TransportProtocols: []stack.TransportProtocolFactory{
			tcp.NewProtocol,
			udp.NewProtocol,
			icmp.NewProtocol4,
		},
	})

	if err := s.CreateNIC(1, endpoint); err != nil {
		return nil, errors.New(err.String())
	}

	if err := s.AddAddress(1, arp.ProtocolNumber, "arp"); err != nil {
		return nil, errors.New(err.String())
	}

	if err := s.AddAddress(1, ipv4.ProtocolNumber, tcpip.Address(net.ParseIP(configuration.GatewayIP).To4())); err != nil {
		return nil, errors.New(err.String())
	}

	s.SetPromiscuousMode(1, true)

	_, parsedSubnet, err := net.ParseCIDR(configuration.Subnet)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse cidr")
	}

	subnet, err := tcpip.NewSubnet(tcpip.Address(parsedSubnet.IP), tcpip.AddressMask(parsedSubnet.Mask))
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse subnet")
	}
	s.SetRouteTable([]tcpip.Route{
		{
			Destination: subnet,
			Gateway:     "",
			NIC:         1,
		},
	})

	return s, nil
}
