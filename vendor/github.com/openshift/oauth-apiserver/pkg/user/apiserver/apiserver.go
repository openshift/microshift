package apiserver

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"

	userapiv1 "github.com/openshift/api/user/v1"
	userclient "github.com/openshift/client-go/user/clientset/versioned"
	"github.com/openshift/oauth-apiserver/pkg/serverscheme"
	groupetcd "github.com/openshift/oauth-apiserver/pkg/user/apiserver/registry/group/etcd"
	identityetcd "github.com/openshift/oauth-apiserver/pkg/user/apiserver/registry/identity/etcd"
	useretcd "github.com/openshift/oauth-apiserver/pkg/user/apiserver/registry/user/etcd"
	"github.com/openshift/oauth-apiserver/pkg/user/apiserver/registry/useridentitymapping"
)

type ExtraConfig struct {
}

type UserConfig struct {
	GenericConfig *genericapiserver.RecommendedConfig
	ExtraConfig   ExtraConfig
}

type UserServer struct {
	GenericAPIServer *genericapiserver.GenericAPIServer
}

type completedConfig struct {
	GenericConfig genericapiserver.CompletedConfig
	ExtraConfig   *ExtraConfig
}

type CompletedConfig struct {
	// Embed a private pointer that cannot be instantiated outside of this package.
	*completedConfig
}

// Complete fills in any fields not set that are required to have valid data. It's mutating the receiver.
func (c *UserConfig) Complete() completedConfig {
	cfg := completedConfig{
		c.GenericConfig.Complete(),
		&c.ExtraConfig,
	}

	return cfg
}

// New returns a new instance of UserServer from the given config.
func (c completedConfig) New(delegationTarget genericapiserver.DelegationTarget) (*UserServer, error) {
	genericServer, err := c.GenericConfig.New("user.openshift.io-apiserver", delegationTarget)
	if err != nil {
		return nil, err
	}

	s := &UserServer{
		GenericAPIServer: genericServer,
	}

	v1Storage, err := c.newV1RESTStorage()
	if err != nil {
		return nil, err
	}

	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(userapiv1.GroupName, serverscheme.Scheme, metav1.ParameterCodec, serverscheme.Codecs)
	apiGroupInfo.VersionedResourcesStorageMap[userapiv1.SchemeGroupVersion.Version] = v1Storage
	if err := s.GenericAPIServer.InstallAPIGroup(&apiGroupInfo); err != nil {
		return nil, err
	}

	return s, nil
}

func (c *completedConfig) newV1RESTStorage() (map[string]rest.Storage, error) {
	userClient, err := userclient.NewForConfig(c.GenericConfig.LoopbackClientConfig)
	if err != nil {
		return nil, err
	}
	userStorage, err := useretcd.NewREST(c.GenericConfig.RESTOptionsGetter)
	if err != nil {
		return nil, err
	}
	identityStorage, err := identityetcd.NewREST(c.GenericConfig.RESTOptionsGetter)
	if err != nil {
		return nil, err
	}
	userIdentityMappingStorage := useridentitymapping.NewREST(userClient.UserV1().Users(), userClient.UserV1().Identities())
	groupStorage, err := groupetcd.NewREST(c.GenericConfig.RESTOptionsGetter)
	if err != nil {
		return nil, err
	}

	v1Storage := map[string]rest.Storage{}
	v1Storage["users"] = userStorage
	v1Storage["groups"] = groupStorage
	v1Storage["identities"] = identityStorage
	v1Storage["userIdentityMappings"] = userIdentityMappingStorage
	return v1Storage, nil
}
