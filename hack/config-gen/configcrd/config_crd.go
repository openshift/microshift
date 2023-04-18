// +kubebuilder:object:generate=true
// +groupName=config.microshift.openshift.io
// +versionName=v1

// This file only exists to help translate our config struct
// to an OpenAPIV3 spec via controller-gen, it should not be exported
package configcrd

import (
	"github.com/openshift/microshift/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//nolint:unused
type configSpec struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Config config.Config `json:"config"`
}
