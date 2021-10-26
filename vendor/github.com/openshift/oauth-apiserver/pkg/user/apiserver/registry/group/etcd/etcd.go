package etcd

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/generic/registry"

	"github.com/openshift/api/user"
	"github.com/openshift/oauth-apiserver/pkg/printers"
	"github.com/openshift/oauth-apiserver/pkg/printerstorage"

	userapi "github.com/openshift/oauth-apiserver/pkg/user/apis/user"
	"github.com/openshift/oauth-apiserver/pkg/user/apiserver/registry/group"
	userprinters "github.com/openshift/oauth-apiserver/pkg/user/printers/internalversion"
)

// REST implements a RESTStorage for groups against etcd
type REST struct {
	*registry.Store
}

// NewREST returns a RESTStorage object that will work against groups
func NewREST(optsGetter generic.RESTOptionsGetter) (*REST, error) {
	store := &registry.Store{
		NewFunc:                  func() runtime.Object { return &userapi.Group{} },
		NewListFunc:              func() runtime.Object { return &userapi.GroupList{} },
		DefaultQualifiedResource: user.Resource("groups"),

		TableConvertor: printerstorage.TableConvertor{TableGenerator: printers.NewTableGenerator().With(userprinters.AddUserOpenShiftHandler)},

		CreateStrategy: group.Strategy,
		UpdateStrategy: group.Strategy,
		DeleteStrategy: group.Strategy,
	}

	options := &generic.StoreOptions{RESTOptions: optsGetter}
	if err := store.CompleteWithOptions(options); err != nil {
		return nil, err
	}

	return &REST{store}, nil
}
