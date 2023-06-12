package dns

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

type dnsHandler struct {
	zones     []types.Zone
	zonesLock sync.RWMutex
}

func (h *dnsHandler) handle(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.RecursionAvailable = true
	m.Compress = true
	h.addAnswers(m)
	if err := w.WriteMsg(m); err != nil {
		log.Error(err)
	}
}

func (h *dnsHandler) addAnswers(m *dns.Msg) {
	h.zonesLock.RLock()
	defer h.zonesLock.RUnlock()
	for _, q := range m.Question {
		for _, zone := range h.zones {
			zoneSuffix := fmt.Sprintf(".%s", zone.Name)
			if strings.HasSuffix(q.Name, zoneSuffix) {
				if q.Qtype != dns.TypeA {
					return
				}
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
						return
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
					return
				}
				m.Rcode = dns.RcodeNameError
				return
			}
		}

		resolver := net.Resolver{
			PreferGo: false,
		}
		switch q.Qtype {
		case dns.TypeNS:
			records, err := resolver.LookupNS(context.TODO(), q.Name)
			if err != nil {
				m.Rcode = dns.RcodeNameError
				return
			}
			for _, ns := range records {
				m.Answer = append(m.Answer, &dns.NS{
					Hdr: dns.RR_Header{
						Name:   q.Name,
						Rrtype: dns.TypeNS,
						Class:  dns.ClassINET,
						Ttl:    0,
					},
					Ns: ns.Host,
				})
			}
		case dns.TypeA:
			ips, err := resolver.LookupIPAddr(context.TODO(), q.Name)
			if err != nil {
				m.Rcode = dns.RcodeNameError
				return
			}
			for _, ip := range ips {
				if len(ip.IP.To4()) != net.IPv4len {
					continue
				}
				m.Answer = append(m.Answer, &dns.A{
					Hdr: dns.RR_Header{
						Name:   q.Name,
						Rrtype: dns.TypeA,
						Class:  dns.ClassINET,
						Ttl:    0,
					},
					A: ip.IP.To4(),
				})
			}
		}
	}
}

type Server struct {
	udpConn net.PacketConn
	tcpLn   net.Listener
	handler *dnsHandler
}

func New(udpConn net.PacketConn, tcpLn net.Listener, zones []types.Zone) (*Server, error) {
	handler := &dnsHandler{zones: zones}
	return &Server{udpConn: udpConn, tcpLn: tcpLn, handler: handler}, nil
}

func (s *Server) Serve() error {
	mux := dns.NewServeMux()
	mux.HandleFunc(".", s.handler.handle)
	srv := &dns.Server{
		PacketConn: s.udpConn,
		Handler:    mux,
	}
	return srv.ActivateAndServe()
}

func (s *Server) ServeTCP() error {
	mux := dns.NewServeMux()
	mux.HandleFunc(".", s.handler.handle)
	tcpSrv := &dns.Server{
		Listener: s.tcpLn,
		Handler:  mux,
	}
	return tcpSrv.ActivateAndServe()
}

func (s *Server) Mux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/all", func(w http.ResponseWriter, r *http.Request) {
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

		s.handler.zonesLock.Lock()
		s.handler.zones = append([]types.Zone{req}, s.handler.zones...)
		s.handler.zonesLock.Unlock()
		w.WriteHeader(http.StatusOK)
	})
	return mux
}
