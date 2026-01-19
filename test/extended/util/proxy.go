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
	"time"

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

// NewProxy creates and configures a new proxy server.
// Returns the server, a cleanup function to close the log file, and any error.
// The caller is responsible for:
//   - Starting the server with go server.ListenAndServe()
//   - Shutting it down with server.Shutdown()
//   - Calling the cleanup function after shutdown to close the log file
func NewProxy() (*http.Server, func(), error) {
	if err := setCA(); err != nil {
		return nil, nil, fmt.Errorf("error setting up the CA: %w", err)
	}

	// Create a DEDICATED logger for the proxy instead of using the global one
	proxyLogger := log.New()

	var logFile *os.File
	logfilePath := filepath.Join("out", "goproxylogfile.log")
	f, err := os.OpenFile(logfilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("error opening file: %v", err)
		proxyLogger.SetOutput(os.Stderr)
	} else {
		logFile = f
		proxyLogger.SetOutput(f)
	}

	// Cleanup function to close log file
	cleanup := func() {
		if logFile != nil {
			logFile.Close()
		}
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

	server := &http.Server{
		Addr:              addr,
		Handler:           proxy,
		ReadHeaderTimeout: 10 * time.Second,
	}

	proxyLogger.Infof("Proxy server configured on %s", addr)

	return server, cleanup, nil
}
