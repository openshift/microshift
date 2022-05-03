package podsecuritypolicyselfsubjectreview

import (
	"context"
	"fmt"

	"github.com/openshift/apiserver-library-go/pkg/securitycontextconstraints/sccmatching"
	securityapi "github.com/openshift/openshift-apiserver/pkg/security/apis/security"
	securityvalidation "github.com/openshift/openshift-apiserver/pkg/security/apis/security/validation"
	podsecuritypolicysubjectreview "github.com/openshift/openshift-apiserver/pkg/security/apiserver/registry/podsecuritypolicysubjectreview"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kutilerrors "k8s.io/apimachinery/pkg/util/errors"
	authserviceaccount "k8s.io/apiserver/pkg/authentication/serviceaccount"
	"k8s.io/apiserver/pkg/authentication/user"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
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

// New creates a new PodSecurityPolicySelfSubjectReview object
func (r *REST) New() runtime.Object {
	return &securityapi.PodSecurityPolicySelfSubjectReview{}
}

func (s *REST) NamespaceScoped() bool {
	return true
}

// Create registers a given new pspssr instance to r.registry.
func (r *REST) Create(ctx context.Context, obj runtime.Object, _ rest.ValidateObjectFunc, options *metav1.CreateOptions) (runtime.Object, error) {
	pspssr, ok := obj.(*securityapi.PodSecurityPolicySelfSubjectReview)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("not a PodSecurityPolicySelfSubjectReview: %#v", obj))
	}
	if errs := securityvalidation.ValidatePodSecurityPolicySelfSubjectReview(pspssr); len(errs) > 0 {
		return nil, apierrors.NewInvalid(coreapi.Kind("PodSecurityPolicySelfSubjectReview"), "", errs)
	}
	userInfo, ok := apirequest.UserFrom(ctx)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("no user data associated with context"))
	}
	ns, ok := apirequest.NamespaceFrom(ctx)
	if !ok {
		return nil, apierrors.NewBadRequest("namespace parameter required.")
	}

	users := []user.Info{userInfo}
	saName := pspssr.Spec.Template.Spec.ServiceAccountName
	if len(saName) > 0 {
		users = append(users, authserviceaccount.UserInfo(ns, saName, ""))
	}

	matchedConstraints, err := r.sccMatcher.FindApplicableSCCs(ctx, ns, users...)
	if err != nil {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("unable to find SecurityContextConstraints: %v", err))
	}

	providers, errs := sccmatching.CreateProvidersFromConstraints(ctx, ns, matchedConstraints, r.client)
	if len(errs) > 0 {
		return nil, kutilerrors.NewAggregate(errs)
	}

	for _, provider := range providers {
		filled, err := podsecuritypolicysubjectreview.FillPodSecurityPolicySubjectReviewStatus(&pspssr.Status, provider, pspssr.Spec.Template.Spec, provider.GetSCC())
		if err != nil {
			klog.Errorf("unable to fill PodSecurityPolicySelfSubjectReview from constraint %v", err)
			continue
		}
		if filled {
			return pspssr, nil
		}
	}
	return pspssr, nil
}
