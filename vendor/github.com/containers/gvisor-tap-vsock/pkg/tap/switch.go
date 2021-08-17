package tap

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"

	"github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/buffer"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

type VirtualDevice interface {
	DeliverNetworkPacket(remote, local tcpip.LinkAddress, protocol tcpip.NetworkProtocolNumber, pkt *stack.PacketBuffer)
	LinkAddress() tcpip.LinkAddress
	IP() string
}

type NetworkSwitch interface {
	DeliverNetworkPacket(remote, local tcpip.LinkAddress, protocol tcpip.NetworkProtocolNumber, pkt *stack.PacketBuffer)
}

type Switch struct {
	Sent     uint64
	Received uint64

	debug               bool
	maxTransmissionUnit int

	nextConnID int
	conns      map[int]net.Conn
	connLock   sync.Mutex

	cam     map[tcpip.LinkAddress]int
	camLock sync.RWMutex

	writeLock sync.Mutex

	gateway VirtualDevice

	protocol protocol
}

func NewSwitch(debug bool, mtu int, protocol types.Protocol) *Switch {
	return &Switch{
		debug:               debug,
		maxTransmissionUnit: mtu,
		conns:               make(map[int]net.Conn),
		cam:                 make(map[tcpip.LinkAddress]int),
		protocol:            protocolImplementation(protocol),
	}
}

func (e *Switch) CAM() map[string]int {
	e.camLock.RLock()
	defer e.camLock.RUnlock()
	ret := make(map[string]int)
	for address, port := range e.cam {
		ret[address.String()] = port
	}
	return ret
}

func (e *Switch) Connect(ep VirtualDevice) {
	e.gateway = ep
}

func (e *Switch) DeliverNetworkPacket(remote, local tcpip.LinkAddress, protocol tcpip.NetworkProtocolNumber, pkt *stack.PacketBuffer) {
	if err := e.tx(local, remote, pkt); err != nil {
		log.Error(err)
	}
}

func (e *Switch) Accept(ctx context.Context, conn net.Conn) error {
	log.Infof("new connection from %s to %s", conn.RemoteAddr().String(), conn.LocalAddr().String())
	id, failed := e.connect(conn)
	if failed {
		log.Error("connection failed")
		return conn.Close()

	}

	defer func() {
		e.connLock.Lock()
		defer e.connLock.Unlock()
		e.disconnect(id, conn)
	}()
	if err := e.rx(ctx, id, conn); err != nil {
		log.Error(errors.Wrapf(err, "cannot receive packets from %s, disconnecting", conn.RemoteAddr().String()))
		return err
	}
	return nil
}

func (e *Switch) connect(conn net.Conn) (int, bool) {
	e.connLock.Lock()
	defer e.connLock.Unlock()

	id := e.nextConnID
	e.nextConnID++

	e.conns[id] = conn
	return id, false
}

func (e *Switch) tx(src, dst tcpip.LinkAddress, pkt *stack.PacketBuffer) error {
	size := e.protocol.Buf()
	e.protocol.Write(size, pkt.Size())

	e.writeLock.Lock()
	defer e.writeLock.Unlock()

	e.connLock.Lock()
	defer e.connLock.Unlock()

	if dst == header.EthernetBroadcastAddress {
		e.camLock.RLock()
		srcID, ok := e.cam[src]
		if !ok {
			srcID = -1
		}
		e.camLock.RUnlock()
		for id, conn := range e.conns {
			if id == srcID {
				continue
			}
			if _, err := conn.Write(size); err != nil {
				e.disconnect(id, conn)
				return err
			}
			for _, view := range pkt.Views() {
				if _, err := conn.Write(view); err != nil {
					e.disconnect(id, conn)
					return err
				}
			}

			atomic.AddUint64(&e.Sent, uint64(pkt.Size()))
		}
	} else {
		e.camLock.RLock()
		id, ok := e.cam[dst]
		if !ok {
			e.camLock.RUnlock()
			return nil
		}
		e.camLock.RUnlock()
		conn := e.conns[id]
		if _, err := conn.Write(size); err != nil {
			e.disconnect(id, conn)
			return err
		}
		for _, view := range pkt.Views() {
			if _, err := conn.Write(view); err != nil {
				e.disconnect(id, conn)
				return err
			}
		}
		atomic.AddUint64(&e.Sent, uint64(pkt.Size()))
	}
	return nil
}

func (e *Switch) disconnect(id int, conn net.Conn) {
	e.camLock.Lock()
	defer e.camLock.Unlock()

	for address, targetConn := range e.cam {
		if targetConn == id {
			delete(e.cam, address)
		}
	}
	_ = conn.Close()
	delete(e.conns, id)
}

func (e *Switch) rx(ctx context.Context, id int, conn net.Conn) error {
	sizeBuf := e.protocol.Buf()
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		default:
			// passthrough
		}
		n, err := io.ReadFull(conn, sizeBuf)
		if err != nil {
			return errors.Wrap(err, "cannot read size from socket")
		}
		size := int(e.protocol.Read(sizeBuf))

		buf := make([]byte, size)
		n, err = io.ReadFull(conn, buf)
		if err != nil {
			return errors.Wrap(err, "cannot read packet from socket")
		}
		if n == 0 || n != size {
			return fmt.Errorf("unexpected size %d != %d", n, size)
		}

		if e.debug {
			packet := gopacket.NewPacket(buf, layers.LayerTypeEthernet, gopacket.Default)
			log.Info(packet.String())
		}

		view := buffer.View(buf)
		eth := header.Ethernet(view)
		vv := buffer.NewVectorisedView(len(view), []buffer.View{view})

		e.camLock.Lock()
		e.cam[eth.SourceAddress()] = id
		e.camLock.Unlock()

		if eth.DestinationAddress() != e.gateway.LinkAddress() {
			if err := e.tx(eth.SourceAddress(), eth.DestinationAddress(), stack.NewPacketBuffer(stack.PacketBufferOptions{
				Data: vv,
			})); err != nil {
				log.Error(err)
			}
		}
		if eth.DestinationAddress() == e.gateway.LinkAddress() || eth.DestinationAddress() == header.EthernetBroadcastAddress {
			vv.TrimFront(header.EthernetMinimumSize)
			e.gateway.DeliverNetworkPacket(
				eth.SourceAddress(),
				eth.DestinationAddress(),
				eth.Type(),
				stack.NewPacketBuffer(stack.PacketBufferOptions{
					Data: vv,
				}),
			)
		}

		atomic.AddUint64(&e.Received, uint64(size))
	}
	return nil
}

func protocolImplementation(protocol types.Protocol) protocol {
	if protocol == types.QemuProtocol {
		return &qemuProtocol{}
	}
	return &hyperkitProtocol{}
}
