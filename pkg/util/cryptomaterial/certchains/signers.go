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

type CSRInfo interface{ GetMeta() CSRMeta }

type CSRMeta struct {
	Name     string
	Validity time.Duration
}

type ClientCertificateSigningRequestInfo struct {
	CSRMeta

	UserInfo user.Info
}

func (i *ClientCertificateSigningRequestInfo) GetMeta() CSRMeta { return i.CSRMeta }

type ServingCertificateSigningRequestInfo struct {
	CSRMeta

	Hostnames []string
}

func (i *ServingCertificateSigningRequestInfo) GetMeta() CSRMeta { return i.CSRMeta }

type PeerCertificateSigningRequestInfo struct {
	CSRMeta

	UserInfo  user.Info
	Hostnames []string
}

func (i *PeerCertificateSigningRequestInfo) GetMeta() CSRMeta { return i.CSRMeta }

type CertificateSigner struct {
	signerName     string
	signerConfig   *crypto.CA
	signerDir      string
	signerValidity time.Duration

	subCAs             map[string]*CertificateSigner
	signedCertificates map[string]*signedCertificateInfo

	caBundlePaths sets.Set[string]
}

type signedCertificateInfo struct {
	CSRInfo
	tlsConfig *crypto.TLSCertificateConfig
}

func (s *CertificateSigner) GetSignerCertPEM() ([]byte, error) {
	certPem, _, err := s.signerConfig.Config.GetPEMBytes()
	return certPem, err
}

func (s *CertificateSigner) Regenerate(certPath ...string) error {
	switch len(certPath) {
	case 0: // renew ourselves and all our sub-certs
		if len(s.signerConfig.Config.Certs) == 1 {
			// this is a root CA, not an intermediary, regen the TLS config
			if err := s.regenerateSelf(); err != nil {
				return fmt.Errorf("failed to regenerate CA %q: %v", s.signerName, err)
			}
		}

		for _, subCAName := range s.GetSubCANames() {
			if err := s.regenerateSubCA(subCAName); err != nil {
				return fmt.Errorf("failed to regenerate subCA %q: %v", subCAName, err)
			}
		}

		for _, certName := range s.GetCertNames() {
			if err := s.regenerateCertificate(certName); err != nil {
				return err
			}
		}
		return nil

	case 1: // renew a direct sub-cert
		if err := s.regenerateSubCA(certPath[0]); !IsSignerNotFoundError(err) {
			// either an error or everything went well
			return err
		}
		return s.regenerateCertificate(certPath[0])

	default: // forward the request to another sub-signer
		if subCA := s.GetSubCA(certPath[0]); subCA != nil {
			return subCA.Regenerate(certPath[1:]...)
		}
		return fmt.Errorf("%q is not an intermediary signer", certPath[0])
	}
}

func (s *CertificateSigner) regenerateSelf() error {
	if err := os.RemoveAll(s.signerDir); err != nil {
		return fmt.Errorf("failed to regenerate CA %q: %v", s.signerName, err)
	}

	signerConfig, _, err := crypto.EnsureCA(
		cryptomaterial.CACertPath(s.signerDir),
		cryptomaterial.CAKeyPath(s.signerDir),
		cryptomaterial.CASerialsPath(s.signerDir),
		s.signerName,
		s.signerValidity,
	)

	if err != nil {
		return fmt.Errorf("failed to regenerate %s CA certificate: %w", s.signerName, err)
	}

	s.signerConfig = signerConfig

	return s.AddToBundles(sets.List[string](s.caBundlePaths)...)
}

func (s *CertificateSigner) regenerateSubCA(subCAName string) error {
	subCA := s.GetSubCA(subCAName)
	if subCA == nil {
		return NewSignerNotFound(subCAName)
	}

	if err := os.RemoveAll(subCA.signerDir); err != nil {
		return fmt.Errorf("failed to remove subCA dir %q: %v", subCA.signerDir, err)
	}

	if err := s.SignSubCA(subCA.toBuilder()); err != nil {
		return fmt.Errorf("failed to regenerate subCA %q certificate: %v", subCA.signerName, err)
	}

	return nil
}

func (s *CertificateSigner) regenerateCertificate(certName string) error {
	certInfo := s.signedCertificates[certName]
	if certInfo == nil {
		return fmt.Errorf("no certificate with name %q was found", certName)
	}

	certDir := filepath.Join(s.signerDir, certInfo.GetMeta().Name)
	if err := os.RemoveAll(certDir); err != nil {
		return fmt.Errorf("failed to remove cert dir %q: %v", certDir, err)
	}

	if err := s.SignCertificate(certInfo.CSRInfo); err != nil {
		return fmt.Errorf("failed to regenerate cert %q: %v", certInfo.GetMeta().Name, err)
	}

	return nil
}

func (s *CertificateSigner) AddToBundles(bundlePaths ...string) error {
	cert := s.signerConfig.Config.Certs[0]

	for _, bundlePath := range bundlePaths {
		bundlePEMs, err := os.ReadFile(bundlePath)
		if err != nil && !os.IsNotExist(err) {
			return err
		}

		certs := []*x509.Certificate{}
		if len(bundlePEMs) > 0 {
			certs, err = crypto.CertsFromPEM(bundlePEMs)
			if err != nil {
				return err
			}
		}

		var certsChanged, certFound bool
		for i, c := range certs {
			if c.Subject.String() == cert.Subject.String() && c.Issuer.String() == cert.Issuer.String() {
				certFound = true
				if c.SerialNumber != cert.SerialNumber {
					certs[i] = cert
					certsChanged = true
				}
				break
			}
		}

		if certFound {
			if !certsChanged {
				continue
			}
		} else {
			certs = append(certs, cert)
		}

		// make sure the parent directory exists
		if err := os.MkdirAll(filepath.Dir(bundlePath), os.FileMode(0755)); err != nil {
			return err
		}

		certFileWriter, err := os.OpenFile(bundlePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		defer func() { _ = certFileWriter.Close() }()

		bytes, err := crypto.EncodeCertificates(certs...)
		if err != nil {
			return err
		}
		if _, err := certFileWriter.Write(bytes); err != nil {
			return err
		}

		s.caBundlePaths.Insert(bundlePath)
	}

	return nil
}

func (s *CertificateSigner) toBuilder() CertificateSignerBuilder { //nolint:ireturn
	signer := NewCertificateSigner(s.signerName, s.signerDir, s.signerValidity)

	for _, subCA := range s.subCAs {
		signer = signer.WithSubCAs(subCA.toBuilder())
	}

	for _, cert := range s.signedCertificates {
		switch csrInfo := cert.CSRInfo.(type) {
		case *ClientCertificateSigningRequestInfo:
			signer = signer.WithClientCertificates(csrInfo)
		case *ServingCertificateSigningRequestInfo:
			signer = signer.WithServingCertificates(csrInfo)
		case *PeerCertificateSigningRequestInfo:
			signer = signer.WithPeerCertificiates(csrInfo)
		default:
			panic("failed to handle type %T as a CSRInfo")
		}
	}

	signer = signer.WithCABundlePaths(sets.List[string](s.caBundlePaths)...)

	return signer
}

func (s *CertificateSigner) SignCertificate(csrInfo CSRInfo) error {
	switch csrInfo := csrInfo.(type) {
	case *ClientCertificateSigningRequestInfo:
		return s.SignClientCertificate(csrInfo)
	case *ServingCertificateSigningRequestInfo:
		return s.SignServingCertificate(csrInfo)
	case *PeerCertificateSigningRequestInfo:
		return s.SignPeerCertificate(csrInfo)
	default:
		return fmt.Errorf("unknown CSR info type: %T", csrInfo)
	}
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
		subSignerInfo.Validity(),
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

	subCertSigner, err := subSignerInfo.
		WithSignerConfig(subCA).
		Complete()
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
		signInfo.Validity,
	)

	if err != nil {
		return fmt.Errorf("failed to generate client certificate for %q: %w", signInfo.Name, err)
	}

	s.signedCertificates[signInfo.Name] = &signedCertificateInfo{
		CSRInfo:   signInfo,
		tlsConfig: tlsConfig,
	}
	return nil
}

func (s *CertificateSigner) SignServingCertificate(signInfo *ServingCertificateSigningRequestInfo) error {
	certDir := filepath.Join(s.signerDir, signInfo.Name)

	tlsConfig, _, err := s.signerConfig.EnsureServerCert(
		cryptomaterial.ServingCertPath(certDir),
		cryptomaterial.ServingKeyPath(certDir),
		sets.New[string](signInfo.Hostnames...),
		signInfo.Validity,
	)

	if err != nil {
		return fmt.Errorf("failed to generate serving certificate for %q: %w", signInfo.Name, err)
	}

	s.signedCertificates[signInfo.Name] = &signedCertificateInfo{
		CSRInfo:   signInfo,
		tlsConfig: tlsConfig,
	}
	return nil
}

func (s *CertificateSigner) SignPeerCertificate(signInfo *PeerCertificateSigningRequestInfo) error {
	certDir := filepath.Join(s.signerDir, signInfo.Name)

	hostnameSet := sets.New[string](signInfo.Hostnames...)
	if _, err := crypto.GetServerCert(
		cryptomaterial.PeerCertPath(certDir),
		cryptomaterial.PeerKeyPath(certDir),
		hostnameSet,
	); err == nil {
		return nil
	}

	tlsConfig, err := s.signerConfig.MakeServerCertForDuration(
		hostnameSet,
		signInfo.Validity,
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
		CSRInfo:   signInfo,
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
func libraryGoEnsureSubCA(ca *crypto.CA, certFile, keyFile, serialFile, name string, expire time.Duration) (*crypto.CA, bool, error) {
	if subCA, err := crypto.GetCA(certFile, keyFile, serialFile); err == nil {
		return subCA, false, nil
	}
	subCA, err := libraryGoMakeAndWriteSubCA(ca, certFile, keyFile, serialFile, name, expire)
	return subCA, true, err
}

// lilibraryGoMakeAndWriteSubCA comes from lib-go 4.12, use (ca *CA) MakeAndWriteSubCA from there once we get the updated lib-go
func libraryGoMakeAndWriteSubCA(ca *crypto.CA, certFile, keyFile, serialFile, name string, expire time.Duration) (*crypto.CA, error) {
	klog.V(4).Infof("Generating sub-CA certificate in %s, key in %s, serial in %s", certFile, keyFile, serialFile)

	subCAConfig, err := crypto.MakeCAConfigForDuration(name, expire, ca)
	if err != nil {
		return nil, err
	}

	if err := subCAConfig.WriteCertConfigFile(certFile, keyFile); err != nil {
		return nil, err
	}

	var serialGenerator crypto.SerialGenerator
	if len(serialFile) > 0 {
		// create / overwrite the serial file with a zero padded hex value (ending in a newline to have a valid file)
		if err := os.WriteFile(serialFile, []byte("00\n"), 0600); err != nil {
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
