package forwarder

import (
	"context"
	"fmt"
	"net"

	"github.com/google/tcpproxy"
	log "github.com/sirupsen/logrus"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/waiter"
)

func TCP(s *stack.Stack) *tcp.Forwarder {
	return tcp.NewForwarder(s, 30000, 10, func(r *tcp.ForwarderRequest) {
		outbound, err := net.Dial("tcp", fmt.Sprintf("%s:%d", r.ID().LocalAddress, r.ID().LocalPort))
		if err != nil {
			log.Errorf("net.Dial() = %v", err)
			r.Complete(true)
			return
		}

		var wq waiter.Queue
		ep, tcpErr := r.CreateEndpoint(&wq)
		if tcpErr != nil {
			log.Errorf("r.CreateEndpoint() = %v", tcpErr)
			return
		}
		r.Complete(false)

		remote := tcpproxy.DialProxy{
			DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				return outbound, nil
			},
		}
		remote.HandleConn(gonet.NewTCPConn(&wq, ep))
	})
}
