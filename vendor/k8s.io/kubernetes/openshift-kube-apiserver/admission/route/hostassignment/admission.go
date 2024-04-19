package hostassignment

import (
	"context"
	"fmt"
	"io"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/admission/initializer"
	"k8s.io/client-go/kubernetes"
	authorizationv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/component-base/featuregate"

	configv1 "github.com/openshift/api/config/v1"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/openshift/library-go/pkg/config/helpers"
	"github.com/openshift/library-go/pkg/route"
	"github.com/openshift/library-go/pkg/route/hostassignment"
	hostassignmentapi "k8s.io/kubernetes/openshift-kube-apiserver/admission/route/apis/hostassignment"
	hostassignmentv1 "k8s.io/kubernetes/openshift-kube-apiserver/admission/route/apis/hostassignment/v1"
)

const PluginName = "route.openshift.io/RouteHostAssignment"

func Register(plugins *admission.Plugins) {
	plugins.Register(PluginName, func(config io.Reader) (admission.Interface, error) {
		pluginConfig, err := readConfig(config)
		if err != nil {
			return nil, err
		}
		return newHostAssignment(pluginConfig)
	})
}

type hostAssignment struct {
	*admission.Handler

	hostnameGenerator         hostassignment.HostnameGenerator
	sarClient                 authorizationv1.SubjectAccessReviewInterface
	allowExternalCertificates bool
	inspectedFeatureGates     bool
}

func readConfig(reader io.Reader) (*hostassignmentapi.HostAssignmentAdmissionConfig, error) {
	obj, err := helpers.ReadYAMLToInternal(reader, hostassignmentapi.Install, hostassignmentv1.Install)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		scheme := runtime.NewScheme()
		hostassignmentapi.Install(scheme)
		hostassignmentv1.Install(scheme)
		external := &hostassignmentv1.HostAssignmentAdmissionConfig{}
		scheme.Default(external)
		internal := &hostassignmentapi.HostAssignmentAdmissionConfig{}
		if err := scheme.Convert(external, internal, nil); err != nil {
			return nil, fmt.Errorf("failed to produce default config: %w", err)
		}
		obj = internal
	}
	config, ok := obj.(*hostassignmentapi.HostAssignmentAdmissionConfig)
	if !ok {
		return nil, fmt.Errorf("unexpected config object: %#v", obj)
	}
	return config, nil
}

func newHostAssignment(config *hostassignmentapi.HostAssignmentAdmissionConfig) (*hostAssignment, error) {
	hostnameGenerator, err := hostassignment.NewSimpleAllocationPlugin(config.Domain)
	if err != nil {
		return nil, fmt.Errorf("configuration failed: %w", err)
	}
	return &hostAssignment{
		Handler:           admission.NewHandler(admission.Create, admission.Update),
		hostnameGenerator: hostnameGenerator,
	}, nil
}

func toRoute(uncastObj runtime.Object) (*routev1.Route, runtime.Unstructured, field.ErrorList) {
	u, ok := uncastObj.(runtime.Unstructured)
	if !ok {
		return nil, nil, field.ErrorList{
			field.NotSupported(field.NewPath("kind"), fmt.Sprintf("%T", uncastObj), []string{"Route"}),
			field.NotSupported(field.NewPath("apiVersion"), fmt.Sprintf("%T", uncastObj), []string{routev1.GroupVersion.String()}),
		}
	}

	var out routev1.Route
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), &out); err != nil {
		return nil, nil, field.ErrorList{
			field.NotSupported(field.NewPath("kind"), fmt.Sprintf("%T", uncastObj), []string{"Route"}),
			field.NotSupported(field.NewPath("apiVersion"), fmt.Sprintf("%T", uncastObj), []string{routev1.GroupVersion.String()}),
		}
	}

	return &out, u, nil
}

var _ admission.MutationInterface = &hostAssignment{}

func (a *hostAssignment) Admit(ctx context.Context, attributes admission.Attributes, o admission.ObjectInterfaces) error {
	if attributes.GetResource().GroupResource() != (schema.GroupResource{Group: "route.openshift.io", Resource: "routes"}) {
		return nil
	}
	// if a subresource is specified, skip it
	if len(attributes.GetSubresource()) > 0 {
		return nil
	}

	switch attributes.GetOperation() {
	case admission.Create:
		r, u, errs := toRoute(attributes.GetObject())
		if len(errs) > 0 {
			return errors.NewInvalid(attributes.GetKind().GroupKind(), attributes.GetName(), errs)
		}
		errs = hostassignment.AllocateHost(ctx, r, a.sarClient, a.hostnameGenerator, route.RouteValidationOptions{AllowExternalCertificates: a.allowExternalCertificates})
		if len(errs) > 0 {
			return errors.NewInvalid(attributes.GetKind().GroupKind(), attributes.GetName(), errs)
		}
		content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(r)
		if err != nil {
			return errors.NewInvalid(attributes.GetKind().GroupKind(), attributes.GetName(), field.ErrorList{
				field.InternalError(field.NewPath(""), err),
			})
		}
		u.SetUnstructuredContent(content)
	case admission.Update:
		r, _, errs := toRoute(attributes.GetObject())
		if len(errs) > 0 {
			return errors.NewInvalid(attributes.GetKind().GroupKind(), attributes.GetName(), errs)
		}
		old, _, errs := toRoute(attributes.GetOldObject())
		if len(errs) > 0 {
			return errors.NewInvalid(attributes.GetKind().GroupKind(), attributes.GetName(), errs)
		}
		errs = hostassignment.ValidateHostUpdate(ctx, r, old, a.sarClient, route.RouteValidationOptions{AllowExternalCertificates: a.allowExternalCertificates})
		if len(errs) > 0 {
			return errors.NewInvalid(attributes.GetKind().GroupKind(), attributes.GetName(), errs)
		}
	default:
		return admission.NewForbidden(attributes, fmt.Errorf("unhandled operation: %v", attributes.GetOperation()))
	}

	return nil
}

var _ initializer.WantsExternalKubeClientSet = &hostAssignment{}
var _ initializer.WantsFeatures = &hostAssignment{}

func (a *hostAssignment) SetExternalKubeClientSet(clientset kubernetes.Interface) {
	a.sarClient = clientset.AuthorizationV1().SubjectAccessReviews()
}

func (a *hostAssignment) InspectFeatureGates(fgs featuregate.FeatureGate) {
	a.allowExternalCertificates = fgs.Enabled(featuregate.Feature(configv1.FeatureGateRouteExternalCertificate))
	a.inspectedFeatureGates = true
}

func (a *hostAssignment) ValidateInitialization() error {
	if !a.inspectedFeatureGates {
		return fmt.Errorf("did not inspect feature gates")
	}
	if a.sarClient == nil {
		return fmt.Errorf("missing SubjectAccessReview client")
	}
	return nil
}
