package openshift_apiserver

import (
	"github.com/openshift/library-go/pkg/serviceability"
	"k8s.io/klog/v2"

	genericapiserver "k8s.io/apiserver/pkg/server"
	genericapiserveroptions "k8s.io/apiserver/pkg/server/options"
	"k8s.io/client-go/pkg/version"
	"k8s.io/kubernetes/pkg/capabilities"
	kubelettypes "k8s.io/kubernetes/pkg/kubelet/types"

	openshiftcontrolplanev1 "github.com/openshift/api/openshiftcontrolplane/v1"
	"github.com/openshift/openshift-apiserver/pkg/cmd/openshift-apiserver/openshiftapiserver"
	_ "k8s.io/component-base/metrics/prometheus/clientgo"
)

func RunOpenShiftAPIServer(serverConfig *openshiftcontrolplanev1.OpenShiftAPIServerConfig, authenticationOptions *genericapiserveroptions.DelegatingAuthenticationOptions, authorizationOptions *genericapiserveroptions.DelegatingAuthorizationOptions, stopCh <-chan struct{}) error {
	serviceability.InitLogrusFromKlog()
	// Allow privileged containers
	capabilities.Initialize(capabilities.Capabilities{
		AllowPrivileged: true,
		PrivilegedSources: capabilities.PrivilegedSources{
			HostNetworkSources: []string{kubelettypes.ApiserverSource, kubelettypes.FileSource},
			HostPIDSources:     []string{kubelettypes.ApiserverSource, kubelettypes.FileSource},
			HostIPCSources:     []string{kubelettypes.ApiserverSource, kubelettypes.FileSource},
		},
	})

	openshiftAPIServerRuntimeConfig, err := openshiftapiserver.NewOpenshiftAPIConfig(serverConfig, authenticationOptions, authorizationOptions)
	if err != nil {
		return err
	}

	completedOpenshiftAPIServer := openshiftAPIServerRuntimeConfig.Complete()
	openshiftAPIServer, err := completedOpenshiftAPIServer.New(genericapiserver.NewEmptyDelegate())
	if err != nil {
		return err
	}
	preparedOpenshiftAPIServer := openshiftAPIServer.GenericAPIServer.PrepareRun()

	// this **must** be done after PrepareRun() as it sets up the openapi endpoints
	if err := completedOpenshiftAPIServer.WithOpenAPIAggregationController(preparedOpenshiftAPIServer.GenericAPIServer); err != nil {
		return err
	}

	klog.Infof("Starting master on %s (%s)", serverConfig.ServingInfo.BindAddress, version.Get().String())

	return preparedOpenshiftAPIServer.Run(stopCh)
}
