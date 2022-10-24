package certchains

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog/v2"

	"github.com/openshift/library-go/pkg/crypto"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
)

type CertificateSigner struct {
	signerName         string
	signerConfig       *crypto.CA
	signerDir          string
	signerValidityDays int

	subCAs             map[string]*CertificateSigner
	signedCertificates map[string]*signedCertificateInfo
}

type signedCertificateInfo struct {
	certDir   string
	tlsConfig *crypto.TLSCertificateConfig
}

type ClientCertificateSigningRequestInfo struct {
	CertificateSigningRequestInfo

	UserInfo user.Info
}

type ServingCertificateSigningRequestInfo struct {
	CertificateSigningRequestInfo

	Hostnames []string
}

type PeerCertificateSigningRequestInfo struct {
	CertificateSigningRequestInfo

	UserInfo  user.Info
	Hostnames []string
}

type CertificateSigningRequestInfo struct {
	Name         string
	ValidityDays int
}

func (s *CertificateSigner) GetSignerCertPEM() ([]byte, error) {
	certPem, _, err := s.signerConfig.Config.GetPEMBytes()
	return certPem, err
}

func (s *CertificateSigner) SignSubCA(subSignerInfo CertificateSignerBuilder) error {
	subSignerName := subSignerInfo.Name()
	subSignerDir := subSignerInfo.Directory()

	subCA, _, err := libraryGoEnsureSubCA(
		s.signerConfig,
		cryptomaterial.CABundlePath(subSignerDir),
		cryptomaterial.CAKeyPath(subSignerDir),
		cryptomaterial.CASerialsPath(subSignerDir),
		subSignerName,
		subSignerInfo.ValidityDays(),
	)
	if err != nil {
		return fmt.Errorf("failed to generate sub-CA %q: %w", subSignerName, err)
	}

	// the library code above writes the whole cert chain in files but some of
	// the kube code requires a single cert per signer cert file
	subCACertPath := cryptomaterial.CACertPath(subSignerDir)
	if _, err := os.Stat(subCACertPath); err == nil || os.IsNotExist(err) {
		certPEM, err := crypto.EncodeCertificates(subCA.Config.Certs[0])
		if err != nil {
			return fmt.Errorf("failed to encode sub-CA %q certs to pem: %w", subSignerName, err)
		}

		if err := os.WriteFile(subCACertPath, certPEM, os.FileMode(0644)); err != nil {
			return fmt.Errorf("failed to write certificate for sub-CA %q: %w", subSignerName, err)
		}
	}

	// create the internal representation of a signer to inject the subCA CA config
	subSignerInfoInternal := &certificateSigner{
		signerName:         subSignerName,
		signerDir:          subSignerDir,
		signerValidityDays: subSignerInfo.ValidityDays(),

		signerConfig: subCA,
	}

	subCertSigner, err := subSignerInfoInternal.Complete()
	if err != nil {
		return err
	}

	s.subCAs[subCertSigner.signerName] = subCertSigner
	return nil
}

func (s *CertificateSigner) SignClientCertificate(signInfo *ClientCertificateSigningRequestInfo) error {
	certDir := filepath.Join(s.signerDir, signInfo.Name)

	tlsConfig, _, err := s.signerConfig.EnsureClientCertificate(
		cryptomaterial.ClientCertPath(certDir),
		cryptomaterial.ClientKeyPath(certDir),
		signInfo.UserInfo,
		signInfo.ValidityDays,
	)

	if err != nil {
		return fmt.Errorf("failed to generate client certificate for %q: %w", signInfo.Name, err)
	}

	s.signedCertificates[signInfo.Name] = &signedCertificateInfo{
		certDir:   certDir,
		tlsConfig: tlsConfig,
	}
	return nil
}

func (s *CertificateSigner) SignServingCertificate(signInfo *ServingCertificateSigningRequestInfo) error {
	certDir := filepath.Join(s.signerDir, signInfo.Name)

	tlsConfig, _, err := s.signerConfig.EnsureServerCert(
		cryptomaterial.ServingCertPath(certDir),
		cryptomaterial.ServingKeyPath(certDir),
		sets.NewString(signInfo.Hostnames...),
		signInfo.ValidityDays,
	)

	if err != nil {
		return fmt.Errorf("failed to generate serving certificate for %q: %w", signInfo.Name, err)
	}

	s.signedCertificates[signInfo.Name] = &signedCertificateInfo{
		certDir:   certDir,
		tlsConfig: tlsConfig,
	}
	return nil
}

func (s *CertificateSigner) SignPeerCertificate(signInfo *PeerCertificateSigningRequestInfo) error {
	certDir := filepath.Join(s.signerDir, signInfo.Name)

	hostnameSet := sets.NewString(signInfo.Hostnames...)
	if _, err := crypto.GetServerCert(
		cryptomaterial.PeerCertPath(certDir),
		cryptomaterial.PeerKeyPath(certDir),
		hostnameSet,
	); err == nil {
		return nil
	}

	tlsConfig, err := s.signerConfig.MakeServerCertForDuration(
		hostnameSet,
		time.Duration(signInfo.ValidityDays)*24*time.Hour,
		func(certTemplate *x509.Certificate) error {
			certTemplate.Subject = userToSubject(signInfo.UserInfo)
			certTemplate.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}

			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("failed to generate peer certificate for %q: %w", signInfo.Name, err)
	}

	if err := tlsConfig.WriteCertConfigFile(
		cryptomaterial.PeerCertPath(certDir),
		cryptomaterial.PeerKeyPath(certDir),
	); err != nil {
		return fmt.Errorf("failed to write peer certificate for %q: %w", signInfo.Name, err)
	}

	s.signedCertificates[signInfo.Name] = &signedCertificateInfo{
		certDir:   certDir,
		tlsConfig: tlsConfig,
	}

	return nil
}

func (s *CertificateSigner) GetCertNames() []string {
	return signedCertificateInfoMapKeysOrdered(s.signedCertificates)
}

func (s *CertificateSigner) GetCertKey(subjectName string) ([]byte, []byte, error) {
	certConfig, exists := s.signedCertificates[subjectName]
	if !exists {
		return nil, nil, fmt.Errorf("no certificate with name %q was found", subjectName)
	}

	return certConfig.tlsConfig.GetPEMBytes()
}

func (s *CertificateSigner) GetSubCANames() []string {
	return certificateSignersMapKeysOrdered(s.subCAs)
}

func (s *CertificateSigner) GetSubCA(signerName string) *CertificateSigner {
	return s.subCAs[signerName]
}

// TODO: merge the two below functions with generics once we've got go 1.18
func certificateSignersMapKeysOrdered(stringMap map[string]*CertificateSigner) []string {
	keys := make(sort.StringSlice, 0, len(stringMap))
	for k := range stringMap {
		keys = append(keys, k)
	}

	keys.Sort()
	return keys
}

func signedCertificateInfoMapKeysOrdered(stringMap map[string]*signedCertificateInfo) []string {
	keys := make(sort.StringSlice, 0, len(stringMap))
	for k := range stringMap {
		keys = append(keys, k)
	}

	keys.Sort()
	return keys
}

// libraryGoEnsureSubCA comes from lib-go 4.12, use (ca *CA) EnsureSubCA from there once we get the updated lib-go
func libraryGoEnsureSubCA(ca *crypto.CA, certFile, keyFile, serialFile, name string, expireDays int) (*crypto.CA, bool, error) {
	if subCA, err := crypto.GetCA(certFile, keyFile, serialFile); err == nil {
		return subCA, false, err
	}
	subCA, err := libraryGoMakeAndWriteSubCA(ca, certFile, keyFile, serialFile, name, expireDays)
	return subCA, true, err
}

// lilibraryGoMakeAndWriteSubCA comes from lib-go 4.12, use (ca *CA) MakeAndWriteSubCA from there once we get the updated lib-go
func libraryGoMakeAndWriteSubCA(ca *crypto.CA, certFile, keyFile, serialFile, name string, expireDays int) (*crypto.CA, error) {
	klog.V(4).Infof("Generating sub-CA certificate in %s, key in %s, serial in %s", certFile, keyFile, serialFile)

	subCAConfig, err := crypto.MakeCAConfigForDuration(name, time.Duration(expireDays)*time.Hour*24, ca)
	if err != nil {
		return nil, err
	}

	if err := subCAConfig.WriteCertConfigFile(certFile, keyFile); err != nil {
		return nil, err
	}

	var serialGenerator crypto.SerialGenerator
	if len(serialFile) > 0 {
		// create / overwrite the serial file with a zero padded hex value (ending in a newline to have a valid file)
		if err := os.WriteFile(serialFile, []byte("00\n"), 0644); err != nil {
			return nil, err
		}

		serialGenerator, err = crypto.NewSerialFileGenerator(serialFile)
		if err != nil {
			return nil, err
		}
	} else {
		serialGenerator = &crypto.RandomSerialGenerator{}
	}

	subCA := &crypto.CA{
		Config:          subCAConfig,
		SerialGenerator: serialGenerator,
	}

	return subCA, nil
}

type sortedForDER []string

func (s sortedForDER) Len() int {
	return len(s)
}
func (s sortedForDER) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s sortedForDER) Less(i, j int) bool {
	l1 := len(s[i])
	l2 := len(s[j])
	if l1 == l2 {
		return s[i] < s[j]
	}
	return l1 < l2
}

func userToSubject(u user.Info) pkix.Name {
	// Ok we are going to order groups in a peculiar way here to workaround a
	// 2 bugs, 1 in golang (https://github.com/golang/go/issues/24254) which
	// incorrectly encodes Multivalued RDNs and another in GNUTLS clients
	// which are too picky (https://gitlab.com/gnutls/gnutls/issues/403)
	// and try to "correct" this issue when reading client certs.
	//
	// This workaround should be killed once Golang's pkix module is fixed to
	// generate a correct DER encoding.
	//
	// The workaround relies on the fact that the first octect that differs
	// between the encoding of two group RDNs will end up being the encoded
	// length which is directly related to the group name's length. So we'll
	// sort such that shortest names come first.
	ugroups := u.GetGroups()
	groups := make([]string, len(ugroups))
	copy(groups, ugroups)
	sort.Sort(sortedForDER(groups))

	return pkix.Name{
		CommonName:   u.GetName(),
		SerialNumber: u.GetUID(),
		Organization: groups,
	}
}
