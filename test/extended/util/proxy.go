package util

import (
	"crypto/tls"
	"crypto/x509"
	_ "embed" // blanks are good
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/elazarl/goproxy"
	log "github.com/sirupsen/logrus"
)

//go:embed rootCA.crt
var caCert []byte

//go:embed rootCA.key
var caKey []byte

var CACertTempLocation string

func setCA() error {

	certLocation, err := WriteTempFile(string(caCert), "rootCA.crt")
	if err != nil {
		return err
	}
	CACertTempLocation = certLocation

	myCa, err := tls.X509KeyPair(caCert, caKey)
	if err != nil {
		return err
	}
	if myCa.Leaf, err = x509.ParseCertificate(myCa.Certificate[0]); err != nil {
		return err
	}
	goproxy.GoproxyCa = myCa
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: goproxy.TLSConfigFromCA(&myCa)}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: goproxy.TLSConfigFromCA(&myCa)}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: goproxy.TLSConfigFromCA(&myCa)}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: goproxy.TLSConfigFromCA(&myCa)}
	return nil
}

func RunProxy() {

	err := setCA()
	if err != nil {
		log.Fatalf("error setting up the CA: %s", err)
	}

	// Create a DEDICATED logger for the proxy instead of using the global one
	proxyLogger := log.New()

	logfile := filepath.Join("out", "goproxylogfile.log")
	f, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("error opening file: %v", err)
		proxyLogger.SetOutput(os.Stderr)
	} else {
		defer f.Close()
		proxyLogger.SetOutput(f)
	}

	proxy := goproxy.NewProxyHttpServer()

	// Log and handle all CONNECT requests (HTTPS tunnels)
	proxy.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile("^.*$"))).
		HandleConnectFunc(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
			proxyLogger.Infof("PROXY CONNECT: host=%s remoteAddr=%s", host, ctx.Req.RemoteAddr)
			return goproxy.MitmConnect, host
		})

	// Log all HTTP requests passing through the proxy
	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		proxyLogger.Infof("PROXY REQUEST: %s %s %s", req.Method, req.URL.String(), req.RemoteAddr)
		return req, nil
	})

	// Log all HTTP responses passing through the proxy
	proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if resp != nil {
			proxyLogger.Infof("PROXY RESPONSE: %d %s", resp.StatusCode, ctx.Req.URL.String())
		}
		return resp
	})

	ipaddr := "127.0.0.1" // user mode is default on windows, darwin and linux
	addr := fmt.Sprintf("%s:8888", ipaddr)

	proxy.Verbose = true
	proxy.Logger = proxyLogger

	proxyLogger.Infof("Starting goproxy on %s", addr)

	err = http.ListenAndServe(addr, proxy) // #nosec G114
	if err != nil {
		proxyLogger.Errorf("error running proxy: %s", err)
	}
}
