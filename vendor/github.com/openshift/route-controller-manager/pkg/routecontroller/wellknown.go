package routecontroller

import (
	routev1 "github.com/openshift/api/route/v1"
)

// This file contains the well-known vars and constants used by route-controller-manager
// Annotations
var (
	// DestinationCACertificateAnnotationKey provides the secret with the content
	// that should be filled on the spec.tls.destinationCACertificate
	// of the Route being created.
	DestinationCACertificateAnnotationKey = routev1.GroupName + "/destination-ca-certificate-secret"
	// TerminationPolicyAnnotationKey defines what TLSTerminationType should be used
	// on the Route being created
	TerminationPolicyAnnotationKey = routev1.GroupName + "/termination"
	// PropagateIngressLabelFlag defines if the labels of the Ingress resource
	// should be used to replace the labels on the generated Route resource.
	// In case this feature/annotation is enabled, any existing label on the
	// underlying route resource will be replaced by the labels from the parent
	// ingress resource
	PropagateIngressLabelFlag = routev1.GroupName + "/reconcile-labels"
	// IngressClassAnnotation is the legacy annotation used to define which
	// controller/class should reconcile an ingress resource.
	// In case of a conversion from an ingress to route, if the ingress specifies
	// an ingressclass and the ingressclass does not specify openshift.io/ingress-to-route
	// as its controller, the ingress resource will be ignored and not converted
	// to a Route.
	IngressClassAnnotation = "kubernetes.io/ingress.class"
)

// Metrics
const (
	MetricOpenshiftBuildInfo = "openshift_build_info"
	MetricRouteController    = "openshift_ingress_to_route_controller"
	// MetricRouteWithUnmanagedOwner report the number of routes owned by unmanaged ingresses.
	MetricRouteWithUnmanagedOwner = MetricRouteController + "_route_with_unmanaged_owner"
	// MetricIngressWithoutClassName report the number of ingresses that do not specify ingressClassName
	MetricIngressWithoutClassName = MetricRouteController + "_ingress_without_class_name"
)
