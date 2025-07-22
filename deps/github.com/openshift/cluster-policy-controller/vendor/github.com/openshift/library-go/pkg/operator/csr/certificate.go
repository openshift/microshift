package csr

import (
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"time"

	certificatesv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	restclient "k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	certutil "k8s.io/client-go/util/cert"
)

// IsCertificateValid return true if
// 1) All certs in client certificate are not expired.
// 2) At least one cert matches the given subject if specified
func IsCertificateValid(certData []byte, subject *pkix.Name) error {
	certs, err := certutil.ParseCertsPEM(certData)
	if err != nil {
		return fmt.Errorf("unable to parse certificate: %w", err)
	}

	if len(certs) == 0 {
		return fmt.Errorf("No cert found in certificate: %w", err)
	}

	now := time.Now()
	// make sure no cert in the certificate chain expired
	for _, cert := range certs {
		if now.After(cert.NotAfter) {
			return fmt.Errorf("part of the certificate is expired: sub: %v, notAfter: %v", cert.Subject.String(), cert.NotAfter.String())
		}
	}

	if subject == nil {
		return nil
	}

	// check subject of certificates
	for _, cert := range certs {
		if cert.Subject.CommonName != subject.CommonName {
			continue
		}
		return nil
	}

	return fmt.Errorf("the certificate was not issued for subject (cn=%s)", subject.CommonName)
}

// getCertValidityPeriod returns the validity period of the client certificate in the secret
func getCertValidityPeriod(secret *corev1.Secret) (*time.Time, *time.Time, error) {
	if secret.Data == nil {
		return nil, nil, fmt.Errorf("no client certificate found in secret %q", secret.Namespace+"/"+secret.Name)
	}

	certData, ok := secret.Data[TLSCertFile]
	if !ok {
		return nil, nil, fmt.Errorf("no client certificate found in secret %q", secret.Namespace+"/"+secret.Name)
	}

	certs, err := certutil.ParseCertsPEM(certData)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse TLS certificates: %w", err)
	}

	if len(certs) == 0 {
		return nil, nil, errors.New("No cert found in certificate")
	}

	// find out the validity period for all certs in the certificate chain
	var notBefore, notAfter *time.Time
	for index, cert := range certs {
		if index == 0 {
			notBefore = &cert.NotBefore
			notAfter = &cert.NotAfter
			continue
		}

		if notBefore.Before(cert.NotBefore) {
			notBefore = &cert.NotBefore
		}

		if notAfter.After(cert.NotAfter) {
			notAfter = &cert.NotAfter
		}
	}

	return notBefore, notAfter, nil
}

// BuildKubeconfig builds a kubeconfig based on a rest config template with a cert/key pair
func BuildKubeconfig(clientConfig *restclient.Config, certPath, keyPath string) clientcmdapi.Config {
	// Build kubeconfig.
	kubeconfig := clientcmdapi.Config{
		// Define a cluster stanza based on the bootstrap kubeconfig.
		Clusters: map[string]*clientcmdapi.Cluster{"default-cluster": {
			Server:                   clientConfig.Host,
			InsecureSkipTLSVerify:    false,
			CertificateAuthorityData: clientConfig.CAData,
		}},
		// Define auth based on the obtained client cert.
		AuthInfos: map[string]*clientcmdapi.AuthInfo{"default-auth": {
			ClientCertificate: certPath,
			ClientKey:         keyPath,
		}},
		// Define a context that connects the auth info and cluster, and set it as the default
		Contexts: map[string]*clientcmdapi.Context{"default-context": {
			Cluster:   "default-cluster",
			AuthInfo:  "default-auth",
			Namespace: "configuration",
		}},
		CurrentContext: "default-context",
	}

	return kubeconfig
}

// isCSRApproved returns true if the given csr has been approved
func isCSRApproved(csr *certificatesv1.CertificateSigningRequest) bool {
	approved := false
	for _, condition := range csr.Status.Conditions {
		if condition.Type == certificatesv1.CertificateDenied {
			return false
		} else if condition.Type == certificatesv1.CertificateApproved {
			approved = true
		}
	}

	return approved
}
