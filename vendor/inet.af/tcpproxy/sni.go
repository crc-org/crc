// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tcpproxy

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"net"
)

// AddSNIRoute appends a route to the ipPort listener that routes to
// dest if the incoming TLS SNI server name is sni. If it doesn't
// match, rule processing continues for any additional routes on
// ipPort.
//
// The ipPort is any valid net.Listen TCP address.
func (p *Proxy) AddSNIRoute(ipPort, sni string, dest Target) {
	p.AddSNIMatchRoute(ipPort, equals(sni), dest)
}

// AddSNIMatchRoute appends a route to the ipPort listener that routes
// to dest if the incoming TLS SNI server name is accepted by
// matcher. If it doesn't match, rule processing continues for any
// additional routes on ipPort.
//
// The ipPort is any valid net.Listen TCP address.
func (p *Proxy) AddSNIMatchRoute(ipPort string, matcher Matcher, dest Target) {
	p.addRoute(ipPort, sniMatch{matcher: matcher, target: dest})
}

// SNITargetFunc is the func callback used by Proxy.AddSNIRouteFunc.
type SNITargetFunc func(ctx context.Context, sniName string) (t Target, ok bool)

// AddSNIRouteFunc adds a route to ipPort that matches an SNI request and calls
// fn to map its nap to a target.
func (p *Proxy) AddSNIRouteFunc(ipPort string, fn SNITargetFunc) {
	p.addRoute(ipPort, sniMatch{targetFunc: fn})
}

type sniMatch struct {
	matcher Matcher
	target  Target

	// Alternatively, if targetFunc is non-nil, it's used instead:
	targetFunc SNITargetFunc
}

func (m sniMatch) match(br *bufio.Reader) (Target, string) {
	sni := clientHelloServerName(br)
	if sni == "" {
		return nil, ""
	}
	if m.targetFunc != nil {
		if t, ok := m.targetFunc(context.TODO(), sni); ok {
			return t, sni
		}
		return nil, ""
	}
	if m.matcher(context.TODO(), sni) {
		return m.target, sni
	}
	return nil, ""
}

// clientHelloServerName returns the SNI server name inside the TLS ClientHello,
// without consuming any bytes from br.
// On any error, the empty string is returned.
func clientHelloServerName(br *bufio.Reader) (sni string) {
	const recordHeaderLen = 5
	hdr, err := br.Peek(recordHeaderLen)
	if err != nil {
		return ""
	}
	const recordTypeHandshake = 0x16
	if hdr[0] != recordTypeHandshake {
		return "" // Not TLS.
	}
	recLen := int(hdr[3])<<8 | int(hdr[4]) // ignoring version in hdr[1:3]
	helloBytes, err := br.Peek(recordHeaderLen + recLen)
	if err != nil {
		return ""
	}
	tls.Server(sniSniffConn{r: bytes.NewReader(helloBytes)}, &tls.Config{
		GetConfigForClient: func(hello *tls.ClientHelloInfo) (*tls.Config, error) {
			sni = hello.ServerName
			return nil, nil
		},
	}).Handshake()
	return
}

// sniSniffConn is a net.Conn that reads from r, fails on Writes,
// and crashes otherwise.
type sniSniffConn struct {
	r        io.Reader
	net.Conn // nil; crash on any unexpected use
}

func (c sniSniffConn) Read(p []byte) (int, error) { return c.r.Read(p) }
func (sniSniffConn) Write(p []byte) (int, error)  { return 0, io.EOF }
