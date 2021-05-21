package internalversion

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kapi "k8s.io/kubernetes/pkg/apis/core"
	kprinters "k8s.io/kubernetes/pkg/printers"

	routev1 "github.com/openshift/api/route/v1"
	routeapi "github.com/openshift/openshift-apiserver/pkg/route/apis/route"
)

func AddRouteOpenShiftHandlers(h kprinters.PrintHandler) {
	routeColumnDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Host/Port", Type: "string", Description: routev1.RouteSpec{}.SwaggerDoc()["host"]},
		{Name: "Path", Type: "string", Description: routev1.RouteSpec{}.SwaggerDoc()["path"]},
		{Name: "Services", Type: "string", Description: "Primary and alternate route backends."},
		{Name: "Port", Type: "string", Description: routev1.RouteSpec{}.SwaggerDoc()["port"]},
		{Name: "Termination", Type: "string", Description: "Indicates termination type."},
		{Name: "Wildcard", Type: "string", Description: routev1.RouteSpec{}.SwaggerDoc()["wildcardPolicy"]},
	}
	if err := h.TableHandler(routeColumnDefinitions, printRouteList); err != nil {
		panic(err)
	}
	if err := h.TableHandler(routeColumnDefinitions, printRoute); err != nil {
		panic(err)
	}
}

func printRoute(route *routeapi.Route, options kprinters.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: route},
	}

	tlsTerm := ""
	insecurePolicy := ""
	if route.Spec.TLS != nil {
		tlsTerm = string(route.Spec.TLS.Termination)
		insecurePolicy = string(route.Spec.TLS.InsecureEdgeTerminationPolicy)
	}

	name := route.Name

	var (
		matchedHost bool
		reason      string
		host        = route.Spec.Host

		admitted, errors = 0, 0
	)
	for _, ingress := range route.Status.Ingress {
		switch status, condition := ingressConditionStatus(&ingress, routeapi.RouteAdmitted); status {
		case kapi.ConditionTrue:
			admitted++
			if !matchedHost {
				matchedHost = ingress.Host == route.Spec.Host
				host = ingress.Host
			}
		case kapi.ConditionFalse:
			reason = condition.Reason
			errors++
		}
	}
	switch {
	case route.Status.Ingress == nil:
		// this is the legacy case, we should continue to show the host when talking to servers
		// that have not set status ingress, since we can't distinguish this condition from there
		// being no routers.
	case admitted == 0 && errors > 0:
		host = reason
	case errors > 0:
		host = fmt.Sprintf("%s ... %d rejected", host, errors)
	case admitted == 0:
		host = "Pending"
	case admitted > 1:
		host = fmt.Sprintf("%s ... %d more", host, admitted-1)
	}
	var policy string
	switch {
	case len(tlsTerm) != 0 && len(insecurePolicy) != 0:
		policy = fmt.Sprintf("%s/%s", tlsTerm, insecurePolicy)
	case len(tlsTerm) != 0:
		policy = tlsTerm
	case len(insecurePolicy) != 0:
		policy = fmt.Sprintf("default/%s", insecurePolicy)
	default:
		policy = ""
	}

	backends := append([]routeapi.RouteTargetReference{route.Spec.To}, route.Spec.AlternateBackends...)
	totalWeight := int32(0)
	for _, backend := range backends {
		if backend.Weight != nil {
			totalWeight += *backend.Weight
		}
	}
	var backendInfo []string
	for _, backend := range backends {
		switch {
		case backend.Weight == nil, len(backends) == 1 && totalWeight != 0:
			backendInfo = append(backendInfo, backend.Name)
		case totalWeight == 0:
			backendInfo = append(backendInfo, fmt.Sprintf("%s(0%%)", backend.Name))
		default:
			backendInfo = append(backendInfo, fmt.Sprintf("%s(%d%%)", backend.Name, *backend.Weight*100/totalWeight))
		}
	}

	var port string
	if route.Spec.Port != nil {
		port = route.Spec.Port.TargetPort.String()
	} else {
		port = "<all>"
	}

	row.Cells = append(row.Cells, name, host, route.Spec.Path, strings.Join(backendInfo, ","), port, policy, string(route.Spec.WildcardPolicy))

	return []metav1.TableRow{row}, nil
}

func printRouteList(list *routeapi.RouteList, options kprinters.GenerateOptions) ([]metav1.TableRow, error) {
	rows := make([]metav1.TableRow, 0, len(list.Items))
	for i := range list.Items {
		r, err := printRoute(&list.Items[i], options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func ingressConditionStatus(ingress *routeapi.RouteIngress, t routeapi.RouteIngressConditionType) (kapi.ConditionStatus, routeapi.RouteIngressCondition) {
	for _, condition := range ingress.Conditions {
		if t != condition.Type {
			continue
		}
		return condition.Status, condition
	}
	return kapi.ConditionUnknown, routeapi.RouteIngressCondition{}
}

// formatResourceName receives a resource kind, name, and boolean specifying
// whether or not to update the current name to "kind/name"
func formatResourceName(kind schema.GroupKind, name string, withKind bool) string {
	if !withKind || kind.Empty() {
		return name
	}

	return strings.ToLower(kind.String()) + "/" + name
}
