package certchains

import (
	"fmt"

	"github.com/openshift/library-go/pkg/crypto"

	"github.com/openshift/microshift/pkg/util/cryptomaterial"
)

type SignerInfo interface {
	Name() string
	Directory() string
	ValidityDays() int
}

type CertificateSignerBuilder interface {
	SignerInfo

	WithSubCAs(subCAsInfo ...CertificateSignerBuilder) CertificateSignerBuilder
	WithClientCertificates(signInfos ...*ClientCertificateSigningRequestInfo) CertificateSignerBuilder
	WithServingCertificates(signInfos ...*ServingCertificateSigningRequestInfo) CertificateSignerBuilder
	WithPeerCertificiates(signInfos ...*PeerCertificateSigningRequestInfo) CertificateSignerBuilder
	Complete() (*CertificateSigner, error)
}

type certificateSigner struct {
	signerName         string
	signerDir          string
	signerValidityDays int

	// signerConfig should only be used in case this is a sub-ca signer
	// It should be populated during CertificateSigner.SignSubCA()
	signerConfig             *crypto.CA
	subCAs                   []CertificateSignerBuilder
	clientCertificatesToSign []*ClientCertificateSigningRequestInfo
	serverCertificatesToSign []*ServingCertificateSigningRequestInfo
	peerCertificatesToSign   []*PeerCertificateSigningRequestInfo
}

// NewCertificateSigner returns a builder object for a certificate chain for the given signer
func NewCertificateSigner(signerName, signerDir string, validityDays int) CertificateSignerBuilder {
	return &certificateSigner{
		signerName:         signerName,
		signerDir:          signerDir,
		signerValidityDays: validityDays,
	}
}

func (s *certificateSigner) Name() string      { return s.signerName }
func (s *certificateSigner) Directory() string { return s.signerDir }
func (s *certificateSigner) ValidityDays() int { return s.signerValidityDays }

func (s *certificateSigner) WithClientCertificates(signInfos ...*ClientCertificateSigningRequestInfo) CertificateSignerBuilder {
	s.clientCertificatesToSign = append(s.clientCertificatesToSign, signInfos...)
	return s
}

func (s *certificateSigner) WithServingCertificates(signInfos ...*ServingCertificateSigningRequestInfo) CertificateSignerBuilder {
	s.serverCertificatesToSign = append(s.serverCertificatesToSign, signInfos...)
	return s
}

func (s *certificateSigner) WithPeerCertificiates(signInfos ...*PeerCertificateSigningRequestInfo) CertificateSignerBuilder {
	s.peerCertificatesToSign = append(s.peerCertificatesToSign, signInfos...)
	return s
}

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
			s.signerValidityDays,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to generate %s CA certificate: %w", s.signerName, err)
		}
	}

	signerCompleted := &CertificateSigner{
		signerName:         s.signerName,
		signerDir:          s.signerDir,
		signerValidityDays: s.signerValidityDays,
		signerConfig:       signerConfig,

		subCAs:             make(map[string]*CertificateSigner),
		signedCertificates: make(map[string]*signedCertificateInfo),
	}

	for _, subCA := range s.subCAs {
		subCA := subCA
		if err := signerCompleted.SignSubCA(subCA); err != nil {
			return nil, err
		}
	}

	for _, si := range s.clientCertificatesToSign {
		si := si
		if err := signerCompleted.SignClientCertificate(si); err != nil {
			return nil, err
		}
	}

	for _, si := range s.serverCertificatesToSign {
		si := si
		if err := signerCompleted.SignServingCertificate(si); err != nil {
			return nil, err
		}
	}

	for _, si := range s.peerCertificatesToSign {
		si := si
		if err := signerCompleted.SignPeerCertificate(si); err != nil {
			return nil, err
		}
	}

	return signerCompleted, nil
}
