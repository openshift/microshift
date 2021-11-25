package etcd

import (
	"context"
	"fmt"

	kauthenticationv1 "k8s.io/api/authentication/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/registry/rest"
	kauthinternal "k8s.io/kubernetes/pkg/apis/authentication"
	kauthv1internal "k8s.io/kubernetes/pkg/apis/authentication/v1"
	"k8s.io/kubernetes/pkg/registry/authentication/tokenreview"
)

// REST object wraps the kube TokenReviews REST so that we can use it with our own API path
type REST struct {
	wrapped *tokenreview.REST
}

var _ rest.Storage = &REST{}
var _ rest.Creater = &REST{}

func (r *REST) New() runtime.Object {
	return &kauthenticationv1.TokenReview{}
}

func NewREST(tokenAuthenticator authenticator.Request) (*REST, error) {
	return &REST{wrapped: tokenreview.NewREST(tokenAuthenticator, []string{})}, nil
}

func (r *REST) GroupVersionKind(containingGV schema.GroupVersion) schema.GroupVersionKind {
	return kauthenticationv1.SchemeGroupVersion.WithKind("TokenReview")
}

func (r *REST) NamespaceScoped() bool {
	return false
}

func (r *REST) Create(ctx context.Context, obj runtime.Object, validateObj rest.ValidateObjectFunc, createOptions *metav1.CreateOptions) (runtime.Object, error) {
	tokenReview, ok := obj.(*kauthenticationv1.TokenReview)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("not a TokenReview: %#v", obj))
	}

	// convert to internal types here to avoid "cannot find model definition for %v" errors when using internal types in New()
	tokenReviewInternal := &kauthinternal.TokenReview{}
	if err := kauthv1internal.Convert_v1_TokenReview_To_authentication_TokenReview(tokenReview, tokenReviewInternal, nil); err != nil {
		return nil, apierrors.NewInternalError(fmt.Errorf("failed to convert %#v to internal TokenReview: %v", obj, err))
	}
	return r.wrapped.Create(ctx, tokenReviewInternal, validateObj, createOptions)
}
