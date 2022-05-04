package podsecuritypolicysubjectreview

import (
	"context"
	"fmt"

	securityv1 "github.com/openshift/api/security/v1"
	"github.com/openshift/apiserver-library-go/pkg/securitycontextconstraints/sccmatching"
	securityapi "github.com/openshift/openshift-apiserver/pkg/security/apis/security"
	securityvalidation "github.com/openshift/openshift-apiserver/pkg/security/apis/security/validation"
	kapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kutilerrors "k8s.io/apimachinery/pkg/util/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	authserviceaccount "k8s.io/apiserver/pkg/authentication/serviceaccount"
	"k8s.io/apiserver/pkg/authentication/user"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/reference"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/apis/core"
	coreapi "k8s.io/kubernetes/pkg/apis/core"
)

// REST implements the RESTStorage interface in terms of an Registry.
type REST struct {
	sccMatcher sccmatching.SCCMatcher
	client     kubernetes.Interface
}

var _ rest.Creater = &REST{}
var _ rest.Scoper = &REST{}

// NewREST creates a new REST for policies..
func NewREST(m sccmatching.SCCMatcher, c kubernetes.Interface) *REST {
	return &REST{sccMatcher: m, client: c}
}

// New creates a new PodSecurityPolicySubjectReview object
func (r *REST) New() runtime.Object {
	return &securityapi.PodSecurityPolicySubjectReview{}
}

func (s *REST) NamespaceScoped() bool {
	return true
}

// Create registers a given new PodSecurityPolicySubjectReview instance to r.registry.
func (r *REST) Create(ctx context.Context, obj runtime.Object, _ rest.ValidateObjectFunc, options *metav1.CreateOptions) (runtime.Object, error) {
	pspsr, ok := obj.(*securityapi.PodSecurityPolicySubjectReview)
	if !ok {
		return nil, kapierrors.NewBadRequest(fmt.Sprintf("not a PodSecurityPolicySubjectReview: %#v", obj))
	}

	ns, ok := apirequest.NamespaceFrom(ctx)
	if !ok {
		return nil, kapierrors.NewBadRequest("namespace parameter required.")
	}

	if errs := securityvalidation.ValidatePodSecurityPolicySubjectReview(pspsr); len(errs) > 0 {
		return nil, kapierrors.NewInvalid(coreapi.Kind("PodSecurityPolicySubjectReview"), "", errs)
	}

	var users []user.Info

	specUser := &user.DefaultInfo{Name: pspsr.Spec.User, Groups: pspsr.Spec.Groups}
	if len(specUser.Name) > 0 || len(specUser.Groups) > 0 {
		users = append(users, specUser)
	}

	saName := pspsr.Spec.Template.Spec.ServiceAccountName
	if len(saName) > 0 {
		users = append(users, authserviceaccount.UserInfo(ns, saName, ""))
	}

	matchedConstraints, err := r.sccMatcher.FindApplicableSCCs(ctx, ns, users...)
	if err != nil {
		return nil, kapierrors.NewBadRequest(fmt.Sprintf("unable to find SecurityContextConstraints: %v", err))
	}

	providers, errs := sccmatching.CreateProvidersFromConstraints(ctx, ns, matchedConstraints, r.client)
	if len(errs) > 0 {
		return nil, kutilerrors.NewAggregate(errs)
	}

	for _, provider := range providers {
		filled, err := FillPodSecurityPolicySubjectReviewStatus(&pspsr.Status, provider, pspsr.Spec.Template.Spec, provider.GetSCC())
		if err != nil {
			klog.Errorf("unable to fill PodSecurityPolicySubjectReviewStatus from constraint %v", err)
			continue
		}
		if filled {
			return pspsr, nil
		}
	}
	return pspsr, nil
}

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(securityapi.Install(scheme))
	utilruntime.Must(securityv1.Install(scheme))
}

// FillPodSecurityPolicySubjectReviewStatus fills PodSecurityPolicySubjectReviewStatus assigning SecurityContectConstraint to the PodSpec
func FillPodSecurityPolicySubjectReviewStatus(s *securityapi.PodSecurityPolicySubjectReviewStatus, provider sccmatching.SecurityContextConstraintsProvider, spec coreapi.PodSpec, constraint *securityv1.SecurityContextConstraints) (bool, error) {
	pod := &coreapi.Pod{
		Spec: spec,
	}
	if errs := sccmatching.AssignSecurityContext(provider, pod, field.NewPath(fmt.Sprintf("provider %s: ", provider.GetSCCName()))); len(errs) > 0 {
		klog.Errorf("unable to assign SecurityContextConstraints provider: %v", errs)
		s.Reason = "CantAssignSecurityContextConstraintProvider"
		return false, fmt.Errorf("unable to assign SecurityContextConstraints provider: %v", errs.ToAggregate())
	}
	ref, err := reference.GetReference(scheme, constraint)
	if err != nil {
		s.Reason = "CantObtainReference"
		return false, fmt.Errorf("unable to get SecurityContextConstraints reference: %v", err)
	}
	if s.AllowedBy == nil {
		s.AllowedBy = &core.ObjectReference{}
	}
	err = scheme.Convert(ref, s.AllowedBy, nil)
	if err != nil {
		s.Reason = "CantObtainReference"
		return false, fmt.Errorf("unable to convert ObjectReference to an internal type: %v", err)
	}

	if len(spec.ServiceAccountName) > 0 {
		s.Template.Spec = pod.Spec
	}
	return true, nil
}
