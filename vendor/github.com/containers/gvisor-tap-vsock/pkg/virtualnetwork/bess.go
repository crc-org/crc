package virtualnetwork

import (
	"context"
	"net"
)

func (n *VirtualNetwork) AcceptBess(ctx context.Context, conn net.Conn) error {
	return n.networkSwitch.Accept(ctx, conn)
}
