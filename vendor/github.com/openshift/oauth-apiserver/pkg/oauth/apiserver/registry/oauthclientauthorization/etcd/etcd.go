package etcd

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/apiserver/pkg/storage"

	"github.com/openshift/api/oauth"
	oauthapi "github.com/openshift/oauth-apiserver/pkg/oauth/apis/oauth"
	"github.com/openshift/oauth-apiserver/pkg/oauth/apiserver/registry/oauthclient"
	"github.com/openshift/oauth-apiserver/pkg/oauth/apiserver/registry/oauthclientauthorization"
	oauthprinters "github.com/openshift/oauth-apiserver/pkg/oauth/printers/internalversion"
	"github.com/openshift/oauth-apiserver/pkg/printers"
	"github.com/openshift/oauth-apiserver/pkg/printerstorage"
)

// rest implements a RESTStorage for oauth client authorizations against etcd
type REST struct {
	*registry.Store
}

var _ rest.StandardStorage = &REST{}

// NewREST returns a RESTStorage object that will work against oauth clients
func NewREST(optsGetter generic.RESTOptionsGetter, clientGetter oauthclient.Getter) (*REST, error) {
	strategy := oauthclientauthorization.NewStrategy(clientGetter)

	store := &registry.Store{
		NewFunc:                  func() runtime.Object { return &oauthapi.OAuthClientAuthorization{} },
		NewListFunc:              func() runtime.Object { return &oauthapi.OAuthClientAuthorizationList{} },
		DefaultQualifiedResource: oauth.Resource("oauthclientauthorizations"),

		TableConvertor: printerstorage.TableConvertor{TableGenerator: printers.NewTableGenerator().With(oauthprinters.AddOAuthOpenShiftHandler)},

		CreateStrategy: strategy,
		UpdateStrategy: strategy,
		DeleteStrategy: strategy,
	}

	options := &generic.StoreOptions{
		RESTOptions: optsGetter,
		AttrFunc:    storage.AttrFunc(storage.DefaultNamespaceScopedAttr).WithFieldMutation(oauthapi.OAuthClientAuthorizationFieldSelector),
	}
	if err := store.CompleteWithOptions(options); err != nil {
		return nil, err
	}

	return &REST{store}, nil
}
