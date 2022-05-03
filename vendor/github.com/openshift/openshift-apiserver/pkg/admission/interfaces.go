package admission

import (
	configinformers "github.com/openshift/client-go/config/informers/externalversions"
	routeinformers "github.com/openshift/client-go/route/informers/externalversions"
)

// WantsOpenShiftConfigInformers interface should be implemented by admission plugins
// that want to have an openshift config informer factory injected.
type WantsOpenShiftConfigInformers interface {
	SetOpenShiftConfigInformers(informers configinformers.SharedInformerFactory)
}

// WantsOpenShiftRouteInformers interface should be implemented by admission plugins
// that want to have an openshift route informer factory injected.
type WantsOpenShiftRouteInformers interface {
	SetOpenShiftRouteInformers(informers routeinformers.SharedInformerFactory)
}
