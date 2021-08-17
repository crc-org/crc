package forwarder

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/google/tcpproxy"
	log "github.com/sirupsen/logrus"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

type PortsForwarder struct {
	stack *stack.Stack

	proxiesLock sync.Mutex
	proxies     map[string]proxy
}

type proxy struct {
	Local      string `json:"local"`
	Remote     string `json:"remote"`
	underlying *tcpproxy.Proxy
}

func NewPortsForwarder(s *stack.Stack) *PortsForwarder {
	return &PortsForwarder{
		stack:   s,
		proxies: make(map[string]proxy),
	}
}

func (f *PortsForwarder) Expose(local, remote string) error {
	f.proxiesLock.Lock()
	defer f.proxiesLock.Unlock()
	if _, ok := f.proxies[local]; ok {
		return errors.New("proxy already running")
	}

	split := strings.Split(remote, ":")
	if len(split) != 2 {
		return errors.New("invalid remote addr")
	}
	port, err := strconv.Atoi(split[1])
	if err != nil {
		return err
	}
	var p tcpproxy.Proxy
	p.AddRoute(local, &tcpproxy.DialProxy{
		Addr: remote,
		DialContext: func(ctx context.Context, network, addr string) (conn net.Conn, e error) {
			return gonet.DialContextTCP(ctx, f.stack, tcpip.FullAddress{
				NIC:  1,
				Addr: tcpip.Address(net.ParseIP(split[0]).To4()),
				Port: uint16(port),
			}, ipv4.ProtocolNumber)
		},
	})
	if err := p.Start(); err != nil {
		return err
	}
	go func() {
		if err := p.Wait(); err != nil {
			log.Error(err)
		}
	}()
	f.proxies[local] = proxy{
		Local:      local,
		Remote:     remote,
		underlying: &p,
	}
	return nil
}

func (f *PortsForwarder) Unexpose(local string) error {
	f.proxiesLock.Lock()
	defer f.proxiesLock.Unlock()
	proxy, ok := f.proxies[local]
	if !ok {
		return errors.New("proxy not found")
	}
	delete(f.proxies, local)
	return proxy.underlying.Close()
}

func (f *PortsForwarder) Mux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/all", func(w http.ResponseWriter, r *http.Request) {
		ret := make([]proxy, 0)
		for _, proxy := range f.proxies {
			ret = append(ret, proxy)
		}
		_ = json.NewEncoder(w).Encode(ret)
	})
	mux.HandleFunc("/expose", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "post only", http.StatusBadRequest)
			return
		}
		var req types.ExposeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := f.Expose(req.Local, req.Remote); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/unexpose", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "post only", http.StatusBadRequest)
			return
		}
		var req types.UnexposeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := f.Unexpose(req.Local); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	return mux
}
