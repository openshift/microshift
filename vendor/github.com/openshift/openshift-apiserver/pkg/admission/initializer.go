package admission

import (
	configinformers "github.com/openshift/client-go/config/informers/externalversions"
	routeinformers "github.com/openshift/client-go/route/informers/externalversions"
	"k8s.io/apiserver/pkg/admission"
)

// NewOpenShiftInformersInitializer returns an admission plugin initializer that injects
// openshift shared informer factories into admission plugins.
func NewOpenShiftInformersInitializer(
	configInformers configinformers.SharedInformerFactory,
	routeInformers routeinformers.SharedInformerFactory,
) *openshiftInformersInitializer {
	return &openshiftInformersInitializer{
		configInformers: configInformers,
		routeInformers:  routeInformers,
	}
}

type openshiftInformersInitializer struct {
	configInformers configinformers.SharedInformerFactory
	routeInformers  routeinformers.SharedInformerFactory
}

func (i *openshiftInformersInitializer) Initialize(plugin admission.Interface) {
	if wants, ok := plugin.(WantsOpenShiftConfigInformers); ok {
		wants.SetOpenShiftConfigInformers(i.configInformers)
	}
	if wants, ok := plugin.(WantsOpenShiftRouteInformers); ok {
		wants.SetOpenShiftRouteInformers(i.routeInformers)
	}
}
