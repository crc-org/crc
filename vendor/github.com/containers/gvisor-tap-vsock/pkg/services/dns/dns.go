package dns

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/areYouLazy/libhosty"
	"github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

type dnsHandler struct {
	zones       []types.Zone
	zonesLock   sync.RWMutex
	udpClient   *dns.Client
	tcpClient   *dns.Client
	hostsFile   *HostsFile
	nameservers []string
}

func newDNSHandler(zones []types.Zone) (*dnsHandler, error) {

	nameservers, err := getDNSHostAndPort()
	if err != nil {
		return nil, err
	}

	hostsFile, err := NewHostsFile("")
	if err != nil {
		return nil, err
	}

	return &dnsHandler{
		zones:       zones,
		tcpClient:   &dns.Client{Net: "tcp"},
		udpClient:   &dns.Client{Net: "udp"},
		nameservers: nameservers,
		hostsFile:   hostsFile,
	}, nil

}

func (h *dnsHandler) handle(w dns.ResponseWriter, dnsClient *dns.Client, r *dns.Msg, responseMessageSize int) {
	m := h.addAnswers(dnsClient, r)
	edns0 := r.IsEdns0()
	if edns0 != nil {
		responseMessageSize = int(edns0.UDPSize())
	}
	m.Truncate(responseMessageSize)
	if err := w.WriteMsg(m); err != nil {
		log.Error(err)
	}
}

func (h *dnsHandler) handleTCP(w dns.ResponseWriter, r *dns.Msg) {
	h.handle(w, h.tcpClient, r, dns.MaxMsgSize)
}

func (h *dnsHandler) handleUDP(w dns.ResponseWriter, r *dns.Msg) {
	h.handle(w, h.udpClient, r, dns.MinMsgSize)
}

func (h *dnsHandler) addLocalAnswers(m *dns.Msg, q dns.Question) bool {
	// resolve only ipv4 requests
	if q.Qtype != dns.TypeA {
		return false
	}

	h.zonesLock.RLock()
	defer h.zonesLock.RUnlock()

	for _, zone := range h.zones {
		zoneSuffix := fmt.Sprintf(".%s", zone.Name)
		if strings.HasSuffix(q.Name, zoneSuffix) {
			for _, record := range zone.Records {
				withoutZone := strings.TrimSuffix(q.Name, zoneSuffix)
				if (record.Name != "" && record.Name == withoutZone) ||
					(record.Regexp != nil && record.Regexp.MatchString(withoutZone)) {
					m.Answer = append(m.Answer, &dns.A{
						Hdr: dns.RR_Header{
							Name:   q.Name,
							Rrtype: dns.TypeA,
							Class:  dns.ClassINET,
							Ttl:    0,
						},
						A: record.IP,
					})
					return true
				}
			}
			if !zone.DefaultIP.Equal(net.IP("")) {
				m.Answer = append(m.Answer, &dns.A{
					Hdr: dns.RR_Header{
						Name:   q.Name,
						Rrtype: dns.TypeA,
						Class:  dns.ClassINET,
						Ttl:    0,
					},
					A: zone.DefaultIP,
				})
				return true
			}
			m.Rcode = dns.RcodeNameError
			return true
		}
		ip, err := h.hostsFile.LookupByHostname(q.Name)
		if err != nil {
			// ignore only ErrHostnameNotFound error
			if !errors.Is(err, libhosty.ErrHostnameNotFound) {
				log.Errorf("Error during looking in hosts file records: %v", err)
			}
		} else {
			m.Answer = append(m.Answer, &dns.A{
				Hdr: dns.RR_Header{
					Name:   q.Name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    0,
				},
				A: ip,
			})
			return true
		}
	}
	return false
}

func (h *dnsHandler) addAnswers(dnsClient *dns.Client, r *dns.Msg) *dns.Msg {
	m := new(dns.Msg)
	m.SetReply(r)
	m.RecursionAvailable = true

	for _, q := range m.Question {
		if done := h.addLocalAnswers(m, q); done {
			return m

			// ignore IPv6 request, we support only IPv4 requests for now
		} else if q.Qtype == dns.TypeAAAA {
			return m
		}
	}
	for _, nameserver := range h.nameservers {
		msg := r.Copy()
		r, _, err := dnsClient.Exchange(msg, nameserver)
		// return first good answer
		if err == nil {
			return r
		}
		log.Debugf("Error during DNS Exchange: %s", err)
	}

	// return the error if none of configured nameservers has right answer
	m.Rcode = dns.RcodeNameError
	return m
}

type Server struct {
	udpConn net.PacketConn
	tcpLn   net.Listener
	handler *dnsHandler
}

func New(udpConn net.PacketConn, tcpLn net.Listener, zones []types.Zone) (*Server, error) {
	handler, err := newDNSHandler(zones)
	if err != nil {
		return nil, err
	}
	return &Server{udpConn: udpConn, tcpLn: tcpLn, handler: handler}, nil
}

func (s *Server) Serve() error {
	mux := dns.NewServeMux()
	mux.HandleFunc(".", s.handler.handleUDP)
	srv := &dns.Server{
		PacketConn: s.udpConn,
		Handler:    mux,
	}
	return srv.ActivateAndServe()
}

func (s *Server) ServeTCP() error {
	mux := dns.NewServeMux()
	mux.HandleFunc(".", s.handler.handleTCP)
	tcpSrv := &dns.Server{
		Listener: s.tcpLn,
		Handler:  mux,
	}
	return tcpSrv.ActivateAndServe()
}

func (s *Server) Mux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/all", func(w http.ResponseWriter, _ *http.Request) {
		s.handler.zonesLock.RLock()
		_ = json.NewEncoder(w).Encode(s.handler.zones)
		s.handler.zonesLock.RUnlock()
	})

	mux.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "post only", http.StatusBadRequest)
			return
		}
		var req types.Zone
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		s.addZone(req)
		w.WriteHeader(http.StatusOK)
	})
	return mux
}

func (s *Server) addZone(req types.Zone) {
	s.handler.zonesLock.Lock()
	defer s.handler.zonesLock.Unlock()
	for i, zone := range s.handler.zones {
		if zone.Name == req.Name {
			req.Records = append(req.Records, zone.Records...)
			s.handler.zones[i] = req
			return
		}
	}
	// No existing zone for req.Name, add new one
	s.handler.zones = append(s.handler.zones, req)
}
