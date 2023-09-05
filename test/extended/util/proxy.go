package util

import (
	"crypto/tls"
	"crypto/x509"
	_ "embed" // blanks are good
	"flag"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

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

	proxy := goproxy.NewProxyHttpServer()
	proxy.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile("^.*$"))).HandleConnect(goproxy.AlwaysMitm)

	logfile := filepath.Join("out", "goproxylogfile.log")
	f, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	ipaddr := "127.0.0.1"        // user mode is default on windows and darwin
	if runtime.GOOS == "linux" { // system mode is default on linux
		ipaddr = "192.168.130.1"
	}

	verbose := flag.Bool("v", true, "should every proxy request be logged to stdout")
	addr := flag.String(ipaddr, ":8888", "proxy listen address") // using network-mode=user
	flag.Parse()
	proxy.Verbose = *verbose
	proxy.Logger = log.StandardLogger()
	err = http.ListenAndServe(*addr, proxy) // #nosec G114
	if err != nil {
		log.Printf("error running proxy: %s", err)
	}
}
