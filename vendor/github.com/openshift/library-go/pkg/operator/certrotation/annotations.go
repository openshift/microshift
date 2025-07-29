package certrotation

import (
	"github.com/google/go-cmp/cmp"
	"github.com/openshift/api/annotations"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

const (
	// CertificateNotBeforeAnnotation contains the certificate expiration date in RFC3339 format.
	CertificateNotBeforeAnnotation = "auth.openshift.io/certificate-not-before"
	// CertificateNotAfterAnnotation contains the certificate expiration date in RFC3339 format.
	CertificateNotAfterAnnotation = "auth.openshift.io/certificate-not-after"
	// CertificateIssuer contains the common name of the certificate that signed another certificate.
	CertificateIssuer = "auth.openshift.io/certificate-issuer"
	// CertificateHostnames contains the hostnames used by a signer.
	CertificateHostnames = "auth.openshift.io/certificate-hostnames"
	// AutoRegenerateAfterOfflineExpiryAnnotation contains a link to PR and an e2e test name which verifies
	// that TLS artifact is correctly regenerated after it has expired
	AutoRegenerateAfterOfflineExpiryAnnotation string = "certificates.openshift.io/auto-regenerate-after-offline-expiry"
)

type AdditionalAnnotations struct {
	// JiraComponent annotates tls artifacts so that owner could be easily found
	JiraComponent string
	// Description is a human-readable one sentence description of certificate purpose
	Description string
	// AutoRegenerateAfterOfflineExpiry contains a link to PR and an e2e test name which verifies
	// that TLS artifact is correctly regenerated after it has expired
	AutoRegenerateAfterOfflineExpiry string
	// NotBefore contains certificate the certificate creation date in RFC3339 format.
	NotBefore string
	// NotAfter contains certificate the certificate validity date in RFC3339 format.
	NotAfter string
}

func (a AdditionalAnnotations) EnsureTLSMetadataUpdate(meta *metav1.ObjectMeta) bool {
	modified := false
	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}
	if len(a.JiraComponent) > 0 && meta.Annotations[annotations.OpenShiftComponent] != a.JiraComponent {
		diff := cmp.Diff(meta.Annotations[annotations.OpenShiftComponent], a.JiraComponent)
		klog.V(2).Infof("Updating %q annotation for %s/%s, diff: %s", annotations.OpenShiftComponent, meta.Namespace, meta.Name, diff)
		meta.Annotations[annotations.OpenShiftComponent] = a.JiraComponent
		modified = true
	}
	if len(a.Description) > 0 && meta.Annotations[annotations.OpenShiftDescription] != a.Description {
		diff := cmp.Diff(meta.Annotations[annotations.OpenShiftDescription], a.Description)
		klog.V(2).Infof("Updating %q annotation for %s/%s, diff: %s", annotations.OpenShiftDescription, meta.Namespace, meta.Name, diff)
		meta.Annotations[annotations.OpenShiftDescription] = a.Description
		modified = true
	}
	if len(a.AutoRegenerateAfterOfflineExpiry) > 0 && meta.Annotations[AutoRegenerateAfterOfflineExpiryAnnotation] != a.AutoRegenerateAfterOfflineExpiry {
		diff := cmp.Diff(meta.Annotations[AutoRegenerateAfterOfflineExpiryAnnotation], a.AutoRegenerateAfterOfflineExpiry)
		klog.V(2).Infof("Updating %q annotation for %s/%s, diff: %s", AutoRegenerateAfterOfflineExpiryAnnotation, meta.Namespace, meta.Name, diff)
		meta.Annotations[AutoRegenerateAfterOfflineExpiryAnnotation] = a.AutoRegenerateAfterOfflineExpiry
		modified = true
	}
	if len(a.NotBefore) > 0 && meta.Annotations[CertificateNotBeforeAnnotation] != a.NotBefore {
		meta.Annotations[CertificateNotBeforeAnnotation] = a.NotBefore
		modified = true
	}
	if len(a.NotAfter) > 0 && meta.Annotations[CertificateNotAfterAnnotation] != a.NotAfter {
		meta.Annotations[CertificateNotAfterAnnotation] = a.NotAfter
		modified = true
	}
	return modified
}

func NewTLSArtifactObjectMeta(name, namespace string, annotations AdditionalAnnotations) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{
		Namespace: namespace,
		Name:      name,
	}
	_ = annotations.EnsureTLSMetadataUpdate(&meta)
	return meta
}
