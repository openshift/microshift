package certchains

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/openshift/library-go/pkg/crypto"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
)

type SignerInfo interface {
	Name() string
	Directory() string
	Validity() time.Duration
}

type CertificateSignerBuilder interface {
	SignerInfo

	WithSignerConfig(config *crypto.CA) CertificateSignerBuilder
	WithSubCAs(subCAsInfo ...CertificateSignerBuilder) CertificateSignerBuilder
	WithClientCertificates(signInfos ...*ClientCertificateSigningRequestInfo) CertificateSignerBuilder
	WithServingCertificates(signInfos ...*ServingCertificateSigningRequestInfo) CertificateSignerBuilder
	WithPeerCertificiates(signInfos ...*PeerCertificateSigningRequestInfo) CertificateSignerBuilder
	WithCABundlePaths(bundlePath ...string) CertificateSignerBuilder
	Complete() (*CertificateSigner, error)
}

type certificateSigner struct {
	signerName     string
	signerDir      string
	signerValidity time.Duration

	// signerConfig should only be used in case this is a sub-ca signer
	// It should be populated during CertificateSigner.SignSubCA()
	signerConfig       *crypto.CA
	subCAs             []CertificateSignerBuilder
	certificatesToSign []CSRInfo

	// locations of bundles where this signer appears
	caBundlePaths []string
}

// NewCertificateSigner returns a builder object for a certificate chain for the given signer
func NewCertificateSigner(signerName, signerDir string, validity time.Duration) CertificateSignerBuilder {
	return &certificateSigner{
		signerName:     signerName,
		signerDir:      signerDir,
		signerValidity: validity,
	}
}

func (s *certificateSigner) Name() string            { return s.signerName }
func (s *certificateSigner) Directory() string       { return s.signerDir }
func (s *certificateSigner) Validity() time.Duration { return s.signerValidity }

// WithSignerConfig uses the provided configuration in `config` to sign its
// direct certificates.
// This is useful when creating intermediate signers.
func (s *certificateSigner) WithSignerConfig(config *crypto.CA) CertificateSignerBuilder {
	s.signerConfig = config
	return s
}

func (s *certificateSigner) WithCABundlePaths(bundlePaths ...string) CertificateSignerBuilder {
	s.caBundlePaths = append(s.caBundlePaths, bundlePaths...)
	return s
}

func (s *certificateSigner) WithClientCertificates(signInfos ...*ClientCertificateSigningRequestInfo) CertificateSignerBuilder {
	for _, signInfo := range signInfos {
		s.certificatesToSign = append(s.certificatesToSign, signInfo)
	}
	return s
}

func (s *certificateSigner) WithServingCertificates(signInfos ...*ServingCertificateSigningRequestInfo) CertificateSignerBuilder {
	for _, signInfo := range signInfos {
		s.certificatesToSign = append(s.certificatesToSign, signInfo)
	}
	return s
}

func (s *certificateSigner) WithPeerCertificiates(signInfos ...*PeerCertificateSigningRequestInfo) CertificateSignerBuilder {
	for _, signInfo := range signInfos {
		s.certificatesToSign = append(s.certificatesToSign, signInfo)
	}
	return s
}

//nolint:ireturn
func (s *certificateSigner) WithSubCAs(subCAsInfo ...CertificateSignerBuilder) CertificateSignerBuilder {
	s.subCAs = append(s.subCAs, subCAsInfo...)
	return s
}

func (s *certificateSigner) Complete() (*CertificateSigner, error) {
	// in case this is a sub-ca, it's already going to have the signer-config populated
	signerConfig := s.signerConfig
	if signerConfig == nil {
		var err error
		signerConfig, _, err = crypto.EnsureCA(
			cryptomaterial.CACertPath(s.signerDir),
			cryptomaterial.CAKeyPath(s.signerDir),
			cryptomaterial.CASerialsPath(s.signerDir),
			s.signerName,
			s.signerValidity,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to generate %s CA certificate: %w", s.signerName, err)
		}
	}

	signerCompleted := &CertificateSigner{
		signerName:     s.signerName,
		signerDir:      s.signerDir,
		signerValidity: s.signerValidity,
		signerConfig:   signerConfig,

		subCAs:             make(map[string]*CertificateSigner),
		signedCertificates: make(map[string]*signedCertificateInfo),

		caBundlePaths: sets.New[string](),
	}

	for _, subCA := range s.subCAs {
		if err := signerCompleted.SignSubCA(subCA); err != nil {
			return nil, err
		}
	}

	for _, si := range s.certificatesToSign {
		if err := signerCompleted.SignCertificate(si); err != nil {
			return nil, err
		}
	}

	if err := signerCompleted.AddToBundles(s.caBundlePaths...); err != nil {
		return nil, err
	}

	return signerCompleted, nil
}
