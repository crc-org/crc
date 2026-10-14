package dns

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/containers/gvisor-tap-vsock/pkg/apilog"
	"github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

type upstreamResolver interface {
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
	LookupCNAME(ctx context.Context, host string) (string, error)
	LookupMX(ctx context.Context, name string) ([]*net.MX, error)
	LookupNS(ctx context.Context, name string) ([]*net.NS, error)
	LookupSRV(ctx context.Context, service, proto, name string) (string, []*net.SRV, error)
	LookupTXT(ctx context.Context, name string) ([]string, error)
}

var validDNSChars = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

type dnsHandler struct {
	zones     []types.Zone
	zonesLock sync.RWMutex
	upstream  upstreamResolver
	// nameservers are the host's upstream DNS servers ("ip:port") used to
	// forward raw queries for record types that Go's net.Resolver cannot
	// look up directly (SOA, PTR, AAAA, CAA, ...).
	nameservers []string
	client      *dns.Client
}

func (h *dnsHandler) handle(w dns.ResponseWriter, r *dns.Msg, responseMessageSize int) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.RecursionAvailable = true
	h.addAnswers(m)
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
	h.handle(w, r, dns.MaxMsgSize)
}

func (h *dnsHandler) handleUDP(w dns.ResponseWriter, r *dns.Msg) {
	h.handle(w, r, dns.MinMsgSize)
}

func (h *dnsHandler) addLocalAnswers(m *dns.Msg, q dns.Question) bool {
	h.zonesLock.RLock()
	defer h.zonesLock.RUnlock()

	for _, zone := range h.zones {
		zoneSuffix := fmt.Sprintf(".%s", zone.Name)
		lowerName := strings.ToLower(q.Name)
		lowerSuffix := strings.ToLower(zoneSuffix)
		if strings.HasSuffix(lowerName, lowerSuffix) {
			withoutZone := strings.TrimSuffix(lowerName, lowerSuffix)
			var matchedIP net.IP
			for _, record := range zone.Records {
				if (record.Name != "" && strings.EqualFold(record.Name, withoutZone)) ||
					(record.Regexp != nil && record.Regexp.MatchString(withoutZone)) {
					matchedIP = record.IP
					break
				}
			}
			if matchedIP == nil && !zone.DefaultIP.Equal(net.IP("")) {
				matchedIP = zone.DefaultIP
			}

			if matchedIP == nil {
				// No record and no DefaultIP catch-all: the name truly
				// does not exist in this zone.
				m.Rcode = dns.RcodeNameError
				return true
			}
			if q.Qtype != dns.TypeA {
				// Name exists in this zone, but the zone only ever
				// serves A records. Answer NOERROR/NODATA locally
				// instead of forwarding to the real upstream resolver.
				return true
			}
			m.Answer = append(m.Answer, &dns.A{
				Hdr: dns.RR_Header{
					Name:   q.Name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    0,
				},
				A: matchedIP,
			})
			return true
		}
	}
	return false
}

func splitTxt(s string) []string {
	const k = 255
	var c []string

	if len(s) <= k {
		return []string{s}
	}

	for len(s) > k {
		c = append(c, s[:k])
		s = s[k:]
	}

	if len(s) > 0 {
		c = append(c, s)
	}

	return c
}
func (h *dnsHandler) addAnswers(m *dns.Msg) {
	for _, q := range m.Question {
		if done := h.addLocalAnswers(m, q); done {
			return
		}

		resolver := h.upstream
		switch q.Qtype {
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
		case dns.TypeCNAME:
			cname, err := resolver.LookupCNAME(context.TODO(), q.Name)
			if err != nil {
				m.Rcode = dns.RcodeNameError
				return
			}
			m.Answer = append(m.Answer, &dns.CNAME{
				Hdr: dns.RR_Header{
					Name:   q.Name,
					Rrtype: dns.TypeCNAME,
					Class:  dns.ClassINET,
					Ttl:    0,
				},
				Target: cname,
			})
		case dns.TypeMX:
			records, err := resolver.LookupMX(context.TODO(), q.Name)
			if err != nil {
				m.Rcode = dns.RcodeNameError
				return
			}
			for _, mx := range records {
				m.Answer = append(m.Answer, &dns.MX{
					Hdr: dns.RR_Header{
						Name:   q.Name,
						Rrtype: dns.TypeMX,
						Class:  dns.ClassINET,
						Ttl:    0,
					},
					Mx:         mx.Host,
					Preference: mx.Pref,
				})
			}
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
		case dns.TypeSRV:
			_, records, err := resolver.LookupSRV(context.TODO(), "", "", q.Name)
			if err != nil {
				m.Rcode = dns.RcodeNameError
				return
			}
			for _, srv := range records {
				m.Answer = append(m.Answer, &dns.SRV{
					Hdr: dns.RR_Header{
						Name:   q.Name,
						Rrtype: dns.TypeSRV,
						Class:  dns.ClassINET,
						Ttl:    0,
					},
					Port:     srv.Port,
					Priority: srv.Priority,
					Target:   srv.Target,
					Weight:   srv.Weight,
				})
			}
		case dns.TypeTXT:
			txts, err := resolver.LookupTXT(context.TODO(), q.Name)
			if err != nil {
				m.Rcode = dns.RcodeNameError
				return
			}

			for _, txt := range txts {
				m.Answer = append(m.Answer, &dns.TXT{
					Hdr: dns.RR_Header{
						Name:   q.Name,
						Rrtype: dns.TypeTXT,
						Class:  dns.ClassINET,
						Ttl:    0,
					},
					Txt: splitTxt(txt),
				})
			}
		default:
			// Record types not handled above (SOA, PTR, AAAA, CAA, ...) are
			// forwarded verbatim to the host's upstream nameservers, since
			// Go's net.Resolver has no generic query for them.
			h.addAnswersFromNameservers(m, q)
		}
	}
}

// addAnswersFromNameservers forwards the raw query q to the host's upstream
// nameservers and copies the response back into m. It is used as the fallback
// for record types that the typed net.Resolver cannot resolve.
func (h *dnsHandler) addAnswersFromNameservers(m *dns.Msg, q dns.Question) {
	if len(h.nameservers) == 0 {
		// No upstream discovered (e.g. Windows, no /etc/resolv.conf): preserve
		// the previous behavior (empty NOERROR) and leave the Rcode untouched.
		return
	}

	req := new(dns.Msg)
	req.SetQuestion(q.Name, q.Qtype)
	req.RecursionDesired = true

	for _, ns := range h.nameservers {
		resp, _, err := h.client.Exchange(req, ns)
		if err != nil || resp == nil {
			continue
		}
		// Copy the answer and authority sections. A NODATA response carries the
		// zone SOA in the authority (Ns) section. Extra is intentionally not
		// copied: it may hold an OPT/EDNS0 pseudo-record that would conflict
		// with the EDNS handling in handle().
		m.Answer = append(m.Answer, resp.Answer...)
		m.Ns = append(m.Ns, resp.Ns...)
		m.Rcode = resp.Rcode
		return
	}

	// The upstream servers are known but none could be reached.
	m.Rcode = dns.RcodeServerFailure
}

// hostNameservers returns the upstream DNS servers ("ip:port") configured on
// the host. /etc/resolv.conf exists on Linux and macOS; on other platforms
// (e.g. Windows) it is absent and an empty list is returned.
func hostNameservers() []string {
	conf, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		log.Warnf("cannot read /etc/resolv.conf, DNS queries for SOA/PTR/AAAA and other unhandled types will not be forwarded: %v", err)
		return nil
	}
	servers := make([]string, 0, len(conf.Servers))
	for _, s := range conf.Servers {
		servers = append(servers, net.JoinHostPort(s, conf.Port))
	}
	return servers
}

type Server struct {
	udpConn net.PacketConn
	tcpLn   net.Listener
	handler *dnsHandler
}

func New(udpConn net.PacketConn, tcpLn net.Listener, zones []types.Zone) (*Server, error) {
	upstream := &net.Resolver{
		PreferGo: false,
	}
	return NewWithUpstreamResolver(udpConn, tcpLn, zones, upstream)
}

func NewWithUpstreamResolver(udpConn net.PacketConn, tcpLn net.Listener, zones []types.Zone, upstream upstreamResolver) (*Server, error) {
	handler := &dnsHandler{
		zones:       zones,
		upstream:    upstream,
		nameservers: hostNameservers(),
		client:      &dns.Client{},
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
	mux.HandleFunc("/all", func(w http.ResponseWriter, r *http.Request) {
		s.handler.zonesLock.RLock()
		err := json.NewEncoder(w).Encode(s.handler.zones)
		s.handler.zonesLock.RUnlock()
		if err != nil {
			apilog.SetError(r, err)
		}
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

		apilog.AddField(r, "zone", req.Name)

		if err := s.validateZone(req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := s.addZone(req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	return mux
}

func (s *Server) validateZone(req types.Zone) error {
	if req.Name == "" {
		return fmt.Errorf("zone name is required")
	}

	if req.Name == "." {
		return fmt.Errorf("cannot add root zone")
	}

	if !dns.IsFqdn(req.Name) {
		return fmt.Errorf("zone name must be fully qualified (end with '.'): %s", req.Name)
	}

	if _, ok := dns.IsDomainName(req.Name); !ok {
		return fmt.Errorf("invalid DNS zone name: %s", req.Name)
	}

	if !isValidDNSName(req.Name) {
		return fmt.Errorf("zone name contains invalid characters: %s", req.Name)
	}

	if len(req.Records) == 0 && (req.DefaultIP == nil || req.DefaultIP.IsUnspecified()) {
		return fmt.Errorf("zone must have at least one record or a default IP")
	}

	for _, record := range req.Records {
		if record.Name == "" {
			return fmt.Errorf("record name is required")
		}
		if !isValidDNSName(record.Name) {
			return fmt.Errorf("record name contains invalid characters: %s", record.Name)
		}
		if record.IP == nil && record.Regexp == nil {
			return fmt.Errorf("record %s must have an IP or regexp", record.Name)
		}
	}

	return nil
}

func isValidDNSName(name string) bool {
	return validDNSChars.MatchString(name)
}

func (s *Server) addZone(req types.Zone) error {
	s.handler.zonesLock.Lock()
	defer s.handler.zonesLock.Unlock()
	for i, zone := range s.handler.zones {
		if strings.EqualFold(zone.Name, req.Name) {
			if zone.Protected {
				return fmt.Errorf("cannot modify protected zone: %s", req.Name)
			}
			req.Records = append(req.Records, zone.Records...)
			req.Protected = zone.Protected
			req.Name = zone.Name
			s.handler.zones[i] = req
			return nil
		}
	}
	// No existing zone for req.Name, add new one
	req.Protected = false
	s.handler.zones = append(s.handler.zones, req)
	return nil
}
