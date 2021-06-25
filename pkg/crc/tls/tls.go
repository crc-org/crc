package tls

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"math"
	"math/big"
	"net"
	"time"

	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/pkg/errors"
)

// This file is taken from openshift/installer repo and adds more function
// https://github.com/openshift/installer/blob/master/pkg/asset/tls/tls.go
// Importing installer just for this file adds lots of different dependencies
// and also increase the binary size.
const (
	keySize = 2048

	// ValidityOneDay sets the validity of a cert to 24 hours.
	ValidityOneDay = time.Hour * 24

	// ValidityOneYear sets the validity of a cert to 1 year.
	ValidityOneYear = ValidityOneDay * 365

	// ValidityTenYears sets the validity of a cert to 10 years.
	ValidityTenYears = ValidityOneYear * 10
)

// CertCfg contains all needed fields to configure a new certificate
type CertCfg struct {
	DNSNames     []string
	ExtKeyUsages []x509.ExtKeyUsage
	IPAddresses  []net.IP
	KeyUsages    x509.KeyUsage
	Subject      pkix.Name
	Validity     time.Duration
	IsCA         bool
}

// rsaPublicKey reflects the ASN.1 structure of a PKCS#1 public key.
type rsaPublicKey struct {
	N *big.Int
	E int
}

// PrivateKey generates an RSA Private key and returns the value
func PrivateKey() (*rsa.PrivateKey, error) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, errors.Wrap(err, "error generating RSA private key")
	}

	return rsaKey, nil
}

// SelfSignedCertificate creates a self signed certificate
func SelfSignedCertificate(cfg *CertCfg, key *rsa.PrivateKey) (*x509.Certificate, error) {
	serial, err := rand.Int(rand.Reader, new(big.Int).SetInt64(math.MaxInt64))
	if err != nil {
		return nil, err
	}
	cert := x509.Certificate{
		BasicConstraintsValid: true,
		IsCA:                  cfg.IsCA,
		KeyUsage:              cfg.KeyUsages,
		NotAfter:              time.Now().Add(cfg.Validity),
		NotBefore:             time.Now(),
		SerialNumber:          serial,
		Subject:               cfg.Subject,
	}
	// verifies that the CN and/or OU for the cert is set
	if len(cfg.Subject.CommonName) == 0 || len(cfg.Subject.OrganizationalUnit) == 0 {
		return nil, errors.Errorf("certification's subject is not set, or invalid")
	}
	pub := key.Public()
	cert.SubjectKeyId, err = generateSubjectKeyID(pub)
	if err != nil {
		return nil, errors.Wrap(err, "failed to set subject key identifier")
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, &cert, &cert, key.Public(), key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create certificate")
	}
	return x509.ParseCertificate(certBytes)
}

// SignedCertificate creates a new X.509 certificate based on a template.
func SignedCertificate(
	cfg *CertCfg,
	csr *x509.CertificateRequest,
	key *rsa.PrivateKey,
	caCert *x509.Certificate,
	caKey *rsa.PrivateKey,
) (*x509.Certificate, error) {
	serial, err := rand.Int(rand.Reader, new(big.Int).SetInt64(math.MaxInt64))
	if err != nil {
		return nil, err
	}

	certTmpl := x509.Certificate{
		DNSNames:              csr.DNSNames,
		ExtKeyUsage:           cfg.ExtKeyUsages,
		IPAddresses:           csr.IPAddresses,
		KeyUsage:              cfg.KeyUsages,
		NotAfter:              time.Now().Add(cfg.Validity),
		NotBefore:             caCert.NotBefore,
		SerialNumber:          serial,
		Subject:               csr.Subject,
		IsCA:                  cfg.IsCA,
		Version:               3,
		BasicConstraintsValid: true,
	}
	pub := key.Public()
	certTmpl.SubjectKeyId, err = generateSubjectKeyID(pub)
	if err != nil {
		return nil, errors.Wrap(err, "failed to set subject key identifier")
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, &certTmpl, caCert, key.Public(), caKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create x509 certificate")
	}
	return x509.ParseCertificate(certBytes)
}

// generateSubjectKeyID generates a SHA-1 hash of the subject public key.
func generateSubjectKeyID(pub crypto.PublicKey) ([]byte, error) {
	var publicKeyBytes []byte
	var err error

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		publicKeyBytes, err = asn1.Marshal(rsaPublicKey{N: pub.N, E: pub.E})
		if err != nil {
			return nil, errors.Wrap(err, "failed to Marshal ans1 public key")
		}
	case *ecdsa.PublicKey:
		publicKeyBytes = elliptic.Marshal(pub.Curve, pub.X, pub.Y)
	default:
		return nil, errors.New("only RSA and ECDSA public keys supported")
	}

	hash := sha256.Sum256(publicKeyBytes)
	return hash[:], nil
}

// GenerateSignedCertificate generate a key and cert defined by CertCfg and signed by CA.
func GenerateSignedCertificate(caKey *rsa.PrivateKey, caCert *x509.Certificate,
	cfg *CertCfg) (*rsa.PrivateKey, *x509.Certificate, error) {

	// create a private key
	key, err := PrivateKey()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to generate private key")
	}

	// create a CSR
	csrTmpl := x509.CertificateRequest{Subject: cfg.Subject, DNSNames: cfg.DNSNames, IPAddresses: cfg.IPAddresses}
	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &csrTmpl, key)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create certificate request")
	}

	csr, err := x509.ParseCertificateRequest(csrBytes)
	if err != nil {
		logging.Debugf("Failed to parse x509 certificate request: %s", err)
		return nil, nil, errors.Wrap(err, "error parsing x509 certificate request")
	}

	// create a cert
	cert, err := SignedCertificate(cfg, csr, key, caCert, caKey)
	if err != nil {
		logging.Debugf("Failed to create a signed certificate: %s", err)
		return nil, nil, errors.Wrap(err, "failed to create a signed certificate")
	}
	return key, cert, nil
}

// GenerateSelfSignedCertificate generates a key/cert pair defined by CertCfg.
func GenerateSelfSignedCertificate(cfg *CertCfg) (*rsa.PrivateKey, *x509.Certificate, error) {
	key, err := PrivateKey()
	if err != nil {
		logging.Debugf("Failed to generate a private key: %s", err)
		return nil, nil, errors.Wrap(err, "failed to generate private key")
	}

	crt, err := SelfSignedCertificate(cfg, key)
	if err != nil {
		logging.Debugf("Failed to create self-signed certificate: %s", err)
		return nil, nil, errors.Wrap(err, "failed to create self-signed certificate")
	}
	return key, crt, nil
}

func GetSelfSignedCA() (*rsa.PrivateKey, *x509.Certificate, error) {
	rootCAConf := &CertCfg{
		Subject:   pkix.Name{CommonName: "admin-kubeconfig-signer-custom", OrganizationalUnit: []string{"openshift"}},
		KeyUsages: x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		Validity:  ValidityTenYears,
		IsCA:      true,
	}
	return GenerateSelfSignedCertificate(rootCAConf)
}

func GenerateClientCertificate(rootCAKey *rsa.PrivateKey, rootCACert *x509.Certificate) ([]byte, []byte, error) {
	adminUserConf := &CertCfg{
		Subject:      pkix.Name{CommonName: "system:admin", OrganizationalUnit: []string{"system:masters"}},
		KeyUsages:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		Validity:     ValidityTenYears,
	}
	clientKey, clientCrt, err := GenerateSignedCertificate(rootCAKey, rootCACert, adminUserConf)
	if err != nil {
		return nil, nil, err
	}
	return PrivateKeyToPem(clientKey), CertToPem(clientCrt), nil
}

// PrivateKeyToPem converts an rsa.PrivateKey object to pem string
func PrivateKeyToPem(key *rsa.PrivateKey) []byte {
	keyInBytes := x509.MarshalPKCS1PrivateKey(key)
	keyinPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: keyInBytes,
		},
	)
	return keyinPem
}

// CertToPem converts an x509.Certificate object to a pem string
func CertToPem(cert *x509.Certificate) []byte {
	certInPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		},
	)
	return certInPem
}

// VerifyCertificateAgainstRootCA  takes caPEM and certificatePEM as string
// to validate if given certificate is signed by given ca.
func VerifyCertificateAgainstRootCA(ca, certificate string) (bool, error) {
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(ca))
	if !ok {
		return false, errors.New("failed to parse root certificate")
	}

	block, _ := pem.Decode([]byte(certificate))
	if block == nil {
		return false, errors.New("failed to decode client PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, errors.Wrapf(err, "failed to parse client PEM")
	}

	opts := x509.VerifyOptions{
		Roots:     roots,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	if _, err := cert.Verify(opts); err != nil {
		return false, nil
	}
	return true, nil
}
