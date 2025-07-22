package csr

import (
	"crypto/x509/pkix"
	"fmt"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"

	"github.com/openshift/library-go/pkg/controller/factory"
	"github.com/openshift/library-go/pkg/operator/events"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewSimpleClientCertificateController creates a controller that keeps a secret up to date with a client-cert
// valid against the kube-apiserver. This version only works in a single cluster.  The base library allows
// the secret in one cluster and the CSR in another.
func NewSimpleClientCertificateController(
	secretNamespace, secretName string,
	commonName string, groups []string,
	kubeInformers informers.SharedInformerFactory,
	kubeClient kubernetes.Interface,
	recorder events.Recorder,
) (factory.Controller, error) {
	certOptions := ClientCertOption{
		SecretNamespace: secretNamespace,
		SecretName:      secretName,
	}
	csrOptions := CSROption{
		ObjectMeta: metav1.ObjectMeta{},
		Subject: &pkix.Name{
			Organization: groups,
			CommonName:   commonName,
		},
		SignerName:      "kubernetes.io/kube-apiserver-client",
		EventFilterFunc: nil,
	}
	controllerName := fmt.Sprintf("client-cert-%s[%s]", secretName, secretNamespace)

	return NewClientCertificateController(
		certOptions,
		csrOptions,
		kubeInformers.Certificates().V1().CertificateSigningRequests(),
		kubeClient.CertificatesV1().CertificateSigningRequests(),
		kubeInformers.Core().V1().Secrets(),
		kubeClient.CoreV1(),
		recorder,
		controllerName,
	)
}
