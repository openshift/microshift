package importer

import (
	"context"
	"net/http"
	"sync"

	"github.com/docker/distribution"
	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/docker/api/types"
	dockerregistry "github.com/docker/docker/registry"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/credentialprovider"
	"k8s.io/kubernetes/pkg/credentialprovider/secrets"

	"github.com/openshift/library-go/pkg/image/reference"
	"github.com/openshift/library-go/pkg/image/registryclient"
)

const (
	// Location from where to read mounted node credentials.
	nodeCredentialsDir = "/var/lib/kubelet/"
)

func NewStaticCredentialsContext(
	transport http.RoundTripper,
	insecureTransport http.RoundTripper,
	secrets []corev1.Secret,
) *StaticCredentialsContext {
	return &StaticCredentialsContext{
		transport:         transport,
		insecureTransport: insecureTransport,
		secrets:           secrets,
	}
}

type StaticCredentialsContext struct {
	transport         http.RoundTripper
	insecureTransport http.RoundTripper
	secrets           []corev1.Secret
	contexts          sync.Map
}

// Repository retrieves ref docker repository.
//
// Kubernetes Secrets and node pull credentials are merged, the first has
// higher priority. In case of failure reading node pull credentials only
// kubernetes secrets are taken into account and a log entry is created.
func (s *StaticCredentialsContext) Repository(
	ctx context.Context,
	ref reference.DockerImageReference,
	insecure bool,
) (distribution.Repository, error) {
	defRef := ref.DockerClientDefaults()
	repo := defRef.AsRepository().Exact()
	if ctxIf, ok := s.contexts.Load(repo); ok {
		importCtx := ctxIf.(*registryclient.Context)
		return importCtx.Repository(
			ctx, defRef.RegistryURL(), defRef.RepositoryName(), insecure,
		)
	}

	nodeKeyring := &credentialprovider.BasicDockerKeyring{}
	if config, err := credentialprovider.ReadDockerConfigJSONFile(
		[]string{nodeCredentialsDir},
	); err != nil {
		klog.V(5).Infof("proceeding without node credentials: %v", err)
	} else {
		nodeKeyring.Add(config)
	}

	keyring, err := secrets.MakeDockerKeyring(s.secrets, nodeKeyring)
	if err != nil {
		return nil, err
	}

	var cred auth.CredentialStore = registryclient.NoCredentials
	if auths, found := keyring.Lookup(defRef.String()); found {
		cred = dockerregistry.NewStaticCredentialStore(&types.AuthConfig{
			Username: auths[0].Username,
			Password: auths[0].Password,
		})
	}

	importCtx := registryclient.NewContext(
		s.transport, s.insecureTransport,
	).WithCredentials(cred)
	s.contexts.Store(repo, importCtx)

	return importCtx.Repository(
		ctx, defRef.RegistryURL(), defRef.RepositoryName(), insecure,
	)
}
