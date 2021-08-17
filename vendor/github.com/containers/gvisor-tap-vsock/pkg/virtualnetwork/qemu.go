package virtualnetwork

import (
	"context"
	"net"
)

func (n *VirtualNetwork) AcceptQemu(ctx context.Context, conn net.Conn) error {
	return n.networkSwitch.Accept(ctx, conn)
}
