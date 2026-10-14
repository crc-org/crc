package virtualnetwork

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strconv"

	"github.com/containers/gvisor-tap-vsock/pkg/apilog"
	"github.com/containers/gvisor-tap-vsock/pkg/tokenauth"
	"github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/inetaf/tcpproxy"
	log "github.com/sirupsen/logrus"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
)

// newServicesMux builds the raw ServeMux with service handlers.
// Use ServicesMux or Mux for the middleware-wrapped versions.
func (n *VirtualNetwork) newServicesMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/services/", http.StripPrefix("/services", n.servicesMux))
	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(statsAsJSON(n.networkSwitch.Sent, n.networkSwitch.Received, n.stack.Stats())); err != nil {
			apilog.SetError(r, err)
		}
	})
	mux.HandleFunc("/cam", func(w http.ResponseWriter, r *http.Request) {
		cam := n.networkSwitch.CAM()
		apilog.AddField(r, "entries", len(cam))
		if err := json.NewEncoder(w).Encode(cam); err != nil {
			apilog.SetError(r, err)
		}
	})
	mux.HandleFunc("/leases", func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(n.ipPool.Leases()); err != nil {
			apilog.SetError(r, err)
		}
	})
	mux.HandleFunc("/tunnel", func(w http.ResponseWriter, r *http.Request) {
		ip := r.URL.Query().Get("ip")
		apilog.AddField(r, "ip", ip)
		if ip == "" {
			http.Error(w, "ip is mandatory", http.StatusInternalServerError)
			return
		}
		port, err := strconv.ParseUint(r.URL.Query().Get("port"), 10, 16)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		port16 := uint16(port)

		apilog.AddField(r, "port", port16)

		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
			return
		}

		conn, bufrw, err := hj.Hijack()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		if err := bufrw.Flush(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := conn.Write([]byte(`OK`)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		remote := tcpproxy.DialProxy{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return gonet.DialContextTCP(ctx, n.stack, tcpip.FullAddress{
					NIC:  1,
					Addr: tcpip.AddrFrom4Slice(net.ParseIP(ip).To4()),
					Port: port16,
				}, ipv4.ProtocolNumber)
			},
			OnDialError: func(_ net.Conn, dstDialErr error) {
				log.Errorf("cannot dial: %v", dstDialErr)
				// The connection has already been hijacked by this point,
				// so the response body is no longer visible to Middleware;
				// record the error explicitly.
				apilog.SetError(r, dstDialErr)
			},
		}
		remote.HandleConn(conn)
	})
	return mux
}

// ServicesMux returns the services mux wrapped with audit logging middleware.
func (n *VirtualNetwork) ServicesMux() http.Handler {
	return apilog.Middleware(n.newServicesMux())
}

// Mux returns the full mux (services + connect) wrapped with audit logging middleware.
func (n *VirtualNetwork) Mux() http.Handler {
	mux := n.newServicesMux()
	mux.HandleFunc(types.ConnectPath, func(w http.ResponseWriter, r *http.Request) {
		apilog.AddField(r, "protocol", n.configuration.Protocol)
		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
			return
		}
		conn, bufrw, err := hj.Hijack()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		if err := bufrw.Flush(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// io.EOF indicates the VM closed the connection normally; only
		// record genuine failures, not a normal disconnect.
		if err := n.networkSwitch.Accept(context.Background(), conn, n.configuration.Protocol); err != nil && !errors.Is(err, io.EOF) {
			apilog.SetError(r, err)
		}
	})
	return apilog.Middleware(mux)
}

// GatewayMux returns a mux for the gateway endpoint (accessible from the VM)
// It only exposes the forwarder endpoints, optionally wrapped with
// token authentication if SetAPIToken() was called with a non-empty token
func (n *VirtualNetwork) GatewayMux() *http.ServeMux {
	handler := http.Handler(n.newServicesMux())

	// If a token is configured, wrap it with authentication middleware
	if n.apiToken != "" {
		handler = tokenauth.BearerAuthMiddleware(n.apiToken)(handler)
	}
	handler = apilog.Middleware(handler)

	// Only expose the forwarder endpoints
	gatewayMux := http.NewServeMux()
	gatewayMux.Handle("/services/forwarder/all", handler)
	gatewayMux.Handle("/services/forwarder/expose", handler)
	gatewayMux.Handle("/services/forwarder/unexpose", handler)

	return gatewayMux
}
