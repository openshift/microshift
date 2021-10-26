package etcd

import (
	"context"
	"errors"
	"strings"

	kerrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"

	usergroup "github.com/openshift/api/user"
	"github.com/openshift/apiserver-library-go/pkg/apivalidation"
	"github.com/openshift/oauth-apiserver/pkg/printers"
	"github.com/openshift/oauth-apiserver/pkg/printerstorage"
	userapi "github.com/openshift/oauth-apiserver/pkg/user/apis/user"
	"github.com/openshift/oauth-apiserver/pkg/user/apiserver/registry/user"
	userprinters "github.com/openshift/oauth-apiserver/pkg/user/printers/internalversion"
)

// rest implements a RESTStorage for users against etcd
type REST struct {
	*registry.Store
}

var _ rest.StandardStorage = &REST{}

// NewREST returns a RESTStorage object that will work against users
func NewREST(optsGetter generic.RESTOptionsGetter) (*REST, error) {
	store := &registry.Store{
		NewFunc:                  func() runtime.Object { return &userapi.User{} },
		NewListFunc:              func() runtime.Object { return &userapi.UserList{} },
		DefaultQualifiedResource: usergroup.Resource("users"),

		TableConvertor: printerstorage.TableConvertor{TableGenerator: printers.NewTableGenerator().With(userprinters.AddUserOpenShiftHandler)},

		CreateStrategy: user.Strategy,
		UpdateStrategy: user.Strategy,
		DeleteStrategy: user.Strategy,
	}

	options := &generic.StoreOptions{RESTOptions: optsGetter}
	if err := store.CompleteWithOptions(options); err != nil {
		return nil, err
	}

	return &REST{store}, nil
}

// Get retrieves the item from etcd.
func (r *REST) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	// "~" means the currently authenticated user
	if name == "~" {
		user, ok := apirequest.UserFrom(ctx)
		if !ok || user.GetName() == "" {
			return nil, kerrs.NewForbidden(usergroup.Resource("user"), "~", errors.New("requests to ~ must be authenticated"))
		}
		name = user.GetName()
		contextGroups := sets.NewString(user.GetGroups()...).List() // sort and deduplicate

		// build a virtual user object using the context data
		virtualUser := &userapi.User{ObjectMeta: metav1.ObjectMeta{Name: name, UID: types.UID(user.GetUID())}, Groups: contextGroups}

		if reasons := apivalidation.ValidateUserName(name, false); len(reasons) != 0 {
			// The user the authentication layer has identified cannot be a valid persisted user
			// Return an API representation of the virtual user
			return virtualUser, nil
		}

		// see if the context user exists in storage
		obj, err := r.Store.Get(ctx, name, options)

		// valid persisted user
		if err == nil {
			// copy persisted user
			persistedUser := obj.(*userapi.User).DeepCopy()
			// and mutate it to include the complete list of groups from the request context
			persistedUser.Groups = virtualUser.Groups
			// favor the UID on the request since that is what we actually base decisions on
			if len(virtualUser.UID) != 0 {
				persistedUser.UID = virtualUser.UID
			}
			return persistedUser, nil
		}

		// server is broken
		if !kerrs.IsNotFound(err) {
			return nil, kerrs.NewInternalError(err)
		}

		// impersonation, remote token authn, etc
		return virtualUser, nil
	}

	// do not bother looking up users that cannot be persisted
	// make sure we return a status error otherwise the API server will complain
	if reasons := apivalidation.ValidateUserName(name, false); len(reasons) != 0 {
		err := field.Invalid(field.NewPath("metadata", "name"), name, strings.Join(reasons, ", "))
		return nil, kerrs.NewInvalid(usergroup.Kind("User"), name, field.ErrorList{err})
	}

	return r.Store.Get(ctx, name, options)
}
