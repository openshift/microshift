package certrotation

import (
	"github.com/openshift/api/annotations"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewTLSArtifactObjectMeta(name, namespace, jiraComponent, description string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Namespace: namespace,
		Name:      name,
		Annotations: map[string]string{
			annotations.OpenShiftComponent:   jiraComponent,
			annotations.OpenShiftDescription: description,
		},
	}
}

// EnsureTLSMetadataUpdate mutates objectMeta setting necessary annotations if unset
func EnsureTLSMetadataUpdate(meta *metav1.ObjectMeta, jiraComponent, description string) bool {
	modified := false
	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}
	if len(jiraComponent) > 0 && meta.Annotations[annotations.OpenShiftComponent] != jiraComponent {
		meta.Annotations[annotations.OpenShiftComponent] = jiraComponent
		modified = true
	}
	if len(description) > 0 && meta.Annotations[annotations.OpenShiftDescription] != description {
		meta.Annotations[annotations.OpenShiftDescription] = description
		modified = true
	}
	return modified
}
