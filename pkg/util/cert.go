/*
Copyright © 2021 Microshift Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package util

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"io/ioutil"
	"math"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

var (
	rootCA  *x509.Certificate
	rootKey *rsa.PrivateKey
)

const (
	defaultDurationDays = 365
	defaultDuration     = defaultDurationDays * 24 * time.Hour
	defaultHostname     = "localhost"

	keySize = 2048

	ValidityOneDay   = 24 * time.Hour
	ValidityOneYear  = 365 * ValidityOneDay
	ValidityTenYears = 10 * ValidityOneYear
)

func GetRootCA() *x509.Certificate {
	return rootCA
}

func GenCA(common string, svcName []string, duration time.Duration) (*rsa.PrivateKey, *x509.Certificate, error) {
	_, dns := IPAddressesDNSNames(svcName)
	cfg := &CertCfg{
		Validity:     duration,
		IsCA:         true,
		KeyUsages:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		Subject:      pkix.Name{CommonName: common, Organization: dns},
	}
	key, ca, err := cfg.GenerateSelfSignedCertificate()
	return key, ca, err
}

func LoadRootCA(dir, certFilename, keyFilename string) error {

	key, err := ioutil.ReadFile(filepath.Join(dir, keyFilename))
	if err != nil {
		return errors.Wrap(err, "error reading CA key")
	}

	if rootKey, err = PemToPrivateKey(key); err != nil {
		return errors.Wrap(err, "parsing CA key from PEM")
	}

	certPath := filepath.Join(dir, certFilename)
	cert, err := ioutil.ReadFile(certPath)
	if err != nil {
		return errors.Wrap(err, "reading CA certificate")
	}

	if rootCA, err = PemToCertificate(cert); err != nil {
		return errors.Wrap(err, "parsing CA certificate")
	}

	now := time.Now()

	if now.After(rootCA.NotAfter) {
		klog.Errorf("CA has expired: current time %s is after %s", now.Format(time.RFC3339), rootCA.NotAfter.Format(time.RFC3339), nil)
	}

	return nil
}

func StoreRootCA(common, dir, certFilename, keyFilename string, svcName []string) error {
	if rootCA == nil || rootKey == nil {
		var err error
		rootKey, rootCA, err = GenCA(common, svcName, defaultDuration)
		if err != nil {
			return err
		}
	}
	certBuff := CertToPem(rootCA)
	keyBuff := PrivateKeyToPem(rootKey)
	os.MkdirAll(dir, 0700)
	certPath := filepath.Join(dir, certFilename)
	keyPath := filepath.Join(dir, keyFilename)
	ioutil.WriteFile(certPath, certBuff, 0644)
	ioutil.WriteFile(keyPath, keyBuff, 0644)
	return nil
}

// GenCertsBuff create cert and key buff
func GenCertsBuff(common string, svcName []string) ([]byte, []byte, error) {
	ip, dns := IPAddressesDNSNames(svcName)
	cfg := &CertCfg{
		DNSNames:     dns,
		IPAddresses:  ip,
		Validity:     defaultDuration,
		IsCA:         false,
		KeyUsages:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		Subject:      pkix.Name{CommonName: common, Organization: dns},
	}
	key, ca, err := cfg.GenerateSignedCertificate(rootKey, rootCA)
	if err != nil {
		return nil, nil, err
	}
	certBuff := CertToPem(ca)
	keyBuff := PrivateKeyToPem(key)
	return certBuff, keyBuff, nil
}

// GenCerts creates certs and keys
// GenCerts("/var/lib/openshift/service-ca/key", "tls.crt", "tls.key", "example.com")
func GenCerts(common, dir, certFilename, keyFilename string, svcName []string) error {
	var err error
	if rootCA == nil || rootKey == nil {
		return err
	}
	os.MkdirAll(dir, 0700)
	certBuff, keyBuff, err := GenCertsBuff(common, svcName)
	if err != nil {
		return err
	}
	certPath := filepath.Join(dir, certFilename)
	keyPath := filepath.Join(dir, keyFilename)
	ioutil.WriteFile(certPath, certBuff, 0644)
	ioutil.WriteFile(keyPath, keyBuff, 0644)
	return err
}

// GenKeys generates and save rsa keys
func GenKeys(dir, pubFilename, keyFilename string) error {
	key, err := PrivateKey()
	if err != nil {
		return err
	}
	pub := &key.PublicKey
	pubBuff, err := PublicKeyToPem(pub)
	if err != nil {
		return err
	}
	keyBuff := PrivateKeyToPem(key)
	os.MkdirAll(dir, 0700)
	pubPath := filepath.Join(dir, pubFilename)
	keyPath := filepath.Join(dir, keyFilename)
	ioutil.WriteFile(pubPath, pubBuff, 0644)
	ioutil.WriteFile(keyPath, keyBuff, 0644)
	return err

}

// based on github.com/hypershift/certs/tls.go

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

// GenerateSelfSignedCertificate generates a key/cert pair defined by CertCfg.
func (cfg *CertCfg) GenerateSelfSignedCertificate() (*rsa.PrivateKey, *x509.Certificate, error) {
	key, err := PrivateKey()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to generate private key")
	}

	crt, err := cfg.SelfSignedCertificate(key)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create self-signed certificate")
	}
	return key, crt, nil
}

// GenerateSignedCertificate generate a key and cert defined by CertCfg and signed by CA.
func (cfg *CertCfg) GenerateSignedCertificate(caKey *rsa.PrivateKey, caCert *x509.Certificate) (*rsa.PrivateKey, *x509.Certificate, error) {

	if caCert == nil {
		return nil, nil, errors.New("Unable to GenerateSignedCertificate with (nil) caCert")
	}

	if caKey == nil {
		return nil, nil, errors.New("Unable to GenerateSignedCertificate with (nil) caKey")
	}

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
		return nil, nil, errors.Wrap(err, "error parsing x509 certificate request")
	}

	// create a cert
	cert, err := cfg.SignedCertificate(csr, key, caCert, caKey)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create a signed certificate")
	}
	return key, cert, nil
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
func (cfg *CertCfg) SelfSignedCertificate(key *rsa.PrivateKey) (*x509.Certificate, error) {
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
	if len(cfg.Subject.CommonName) == 0 {
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
func (cfg *CertCfg) SignedCertificate(
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
	pub := caCert.PublicKey.(*rsa.PublicKey)
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

	hash := sha1.Sum(publicKeyBytes)
	return hash[:], nil
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

// CSRToPem converts an x509.CertificateRequest to a pem string
func CSRToPem(cert *x509.CertificateRequest) []byte {
	certInPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE REQUEST",
			Bytes: cert.Raw,
		},
	)
	return certInPem
}

// PublicKeyToPem converts an rsa.PublicKey object to pem string
func PublicKeyToPem(key *rsa.PublicKey) ([]byte, error) {
	keyInBytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to MarshalPKIXPublicKey")
	}
	keyinPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: keyInBytes,
		},
	)
	return keyinPem, nil
}

// PemToPrivateKey converts a data block to rsa.PrivateKey.
func PemToPrivateKey(data []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.Errorf("could not find a PEM block in the private key")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// PemToCertificate converts a data block to x509.Certificate.
func PemToCertificate(data []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.Errorf("could not find a PEM block in the certificate")
	}
	return x509.ParseCertificate(block.Bytes)
}

// per https://github.com/openshift/library-go/blob/e555322cb70844f90b003cbdfe7ce002d7c61810/pkg/crypto/crypto.go#L989
func IPAddressesDNSNames(hosts []string) ([]net.IP, []string) {
	ips := []net.IP{}
	dns := []string{}
	for _, host := range hosts {
		if ip := net.ParseIP(host); ip != nil {
			ips = append(ips, ip)
		} else {
			dns = append(dns, host)
		}
	}

	// Include IP addresses as DNS subjectAltNames in the cert as well, for the sake of Python, Windows (< 10), and unnamed other libraries
	// Ensure these technically invalid DNS subjectAltNames occur after the valid ones, to avoid triggering cert errors in Firefox
	// See https://bugzilla.mozilla.org/show_bug.cgi?id=1148766
	for _, ip := range ips {
		dns = append(dns, ip.String())
	}

	return ips, dns
}
func Base64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
