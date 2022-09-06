package route

var ControllerManagerInitialization = map[string]InitFunc{
	"openshift.io/ingress-ip":       RunIngressIPController,
	"openshift.io/ingress-to-route": RunIngressToRouteController,
}

const (
	infraServiceIngressIPControllerServiceAccountName = "service-ingress-ip-controller"
	InfraIngressToRouteControllerServiceAccountName   = "ingress-to-route-controller"

	defaultOpenShiftInfraNamespace = "openshift-infra"
)
