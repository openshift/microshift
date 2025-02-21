package certrotation

import (
	"github.com/google/go-cmp/cmp"
	"github.com/openshift/api/annotations"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

const (
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
}

func (a AdditionalAnnotations) EnsureTLSMetadataUpdate(meta *metav1.ObjectMeta) bool {
	modified := false
	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}
	if len(a.JiraComponent) > 0 && meta.Annotations[annotations.OpenShiftComponent] != a.JiraComponent {
		diff := cmp.Diff(meta.Annotations[annotations.OpenShiftComponent], a.JiraComponent)
		klog.V(2).Infof("Updating %q annotation for %s/%s, diff: %s", annotations.OpenShiftComponent, meta.Name, meta.Namespace, diff)
		meta.Annotations[annotations.OpenShiftComponent] = a.JiraComponent
		modified = true
	}
	if len(a.Description) > 0 && meta.Annotations[annotations.OpenShiftDescription] != a.Description {
		diff := cmp.Diff(meta.Annotations[annotations.OpenShiftDescription], a.Description)
		klog.V(2).Infof("Updating %q annotation for %s/%s, diff: %s", annotations.OpenShiftDescription, meta.Name, meta.Namespace, diff)
		meta.Annotations[annotations.OpenShiftDescription] = a.Description
		modified = true
	}
	if len(a.AutoRegenerateAfterOfflineExpiry) > 0 && meta.Annotations[AutoRegenerateAfterOfflineExpiryAnnotation] != a.AutoRegenerateAfterOfflineExpiry {
		diff := cmp.Diff(meta.Annotations[AutoRegenerateAfterOfflineExpiryAnnotation], a.AutoRegenerateAfterOfflineExpiry)
		klog.V(2).Infof("Updating %q annotation for %s/%s, diff: %s", AutoRegenerateAfterOfflineExpiryAnnotation, meta.Name, meta.Namespace, diff)
		meta.Annotations[AutoRegenerateAfterOfflineExpiryAnnotation] = a.AutoRegenerateAfterOfflineExpiry
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
