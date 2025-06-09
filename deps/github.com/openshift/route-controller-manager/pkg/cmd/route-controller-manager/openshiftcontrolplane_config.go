package openshift_controller_manager

import (
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	openshiftcontrolplanev1 "github.com/openshift/api/openshiftcontrolplane/v1"
	"github.com/openshift/library-go/pkg/config/configdefaults"
)

func asOpenshiftControllerManagerConfig(config *unstructured.Unstructured) (*openshiftcontrolplanev1.OpenShiftControllerManagerConfig, error) {
	result := &openshiftcontrolplanev1.OpenShiftControllerManagerConfig{}
	if config != nil {
		// make a copy we can mutate
		configCopy := config.DeepCopy()
		// force the config to our version to read it
		configCopy.SetGroupVersionKind(openshiftcontrolplanev1.GroupVersion.WithKind("OpenShiftControllerManagerConfig"))
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(configCopy.Object, result); err != nil {
			return nil, err
		}
	}

	setRecommendedOpenShiftControllerConfigDefaults(result)

	return result, nil
}

func setRecommendedOpenShiftControllerConfigDefaults(config *openshiftcontrolplanev1.OpenShiftControllerManagerConfig) {
	configdefaults.SetRecommendedKubeClientConfigDefaults(&config.KubeClientConfig)

	configdefaults.DefaultStringSlice(&config.Controllers, []string{"*"})

	if config.ResourceQuota.MinResyncPeriod.Duration == 0 {
		config.ResourceQuota.MinResyncPeriod.Duration = 5 * time.Minute
	}
	if config.ResourceQuota.SyncPeriod.Duration == 0 {
		config.ResourceQuota.SyncPeriod.Duration = 12 * time.Hour
	}
	if config.ResourceQuota.ConcurrentSyncs == 0 {
		config.ResourceQuota.ConcurrentSyncs = 5
	}
}
