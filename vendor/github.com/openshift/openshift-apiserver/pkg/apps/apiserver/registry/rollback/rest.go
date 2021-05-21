package rollback

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/kubernetes"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"

	v1 "github.com/openshift/openshift-apiserver/pkg/apps/apis/apps/v1"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"

	"github.com/openshift/api/apps"
	appsclient "github.com/openshift/client-go/apps/clientset/versioned"
	appsclienttyped "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"

	"github.com/openshift/library-go/pkg/apps/appsserialization"
	"github.com/openshift/library-go/pkg/apps/appsutil"

	appsapi "github.com/openshift/openshift-apiserver/pkg/apps/apis/apps"
	"github.com/openshift/openshift-apiserver/pkg/apps/apis/apps/validation"
)

// REST provides a rollback generation endpoint. Only the Create method is implemented.
type REST struct {
	generator RollbackGenerator
	dn        appsclienttyped.DeploymentConfigsGetter
	rn        corev1client.ReplicationControllersGetter
}

var _ rest.Creater = &REST{}

// NewREST safely creates a new REST.
func NewREST(appsclient appsclient.Interface, kc kubernetes.Interface) *REST {
	return &REST{
		generator: NewRollbackGenerator(),
		dn:        appsclient.AppsV1(),
		rn:        kc.CoreV1(),
	}
}

// New creates an empty DeploymentConfigRollback resource
func (r *REST) New() runtime.Object {
	return &appsapi.DeploymentConfigRollback{}
}

// Create generates a new DeploymentConfig representing a rollback.
func (r *REST) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, options *metav1.CreateOptions) (runtime.Object, error) {
	namespace, ok := apirequest.NamespaceFrom(ctx)
	if !ok {
		return nil, apierrors.NewBadRequest("namespace parameter required.")
	}
	rollback, ok := obj.(*appsapi.DeploymentConfigRollback)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("not a rollback spec: %#v", obj))
	}

	if errs := validation.ValidateDeploymentConfigRollback(rollback); len(errs) > 0 {
		return nil, apierrors.NewInvalid(apps.Kind("DeploymentConfigRollback"), rollback.Name, errs)
	}
	if err := createValidation(ctx, obj); err != nil {
		return nil, err
	}

	from, err := r.dn.DeploymentConfigs(namespace).Get(ctx, rollback.Name, metav1.GetOptions{})
	if err != nil {
		return nil, newInvalidError(rollback, fmt.Sprintf("cannot get deployment config %q: %v", rollback.Name, err))
	}

	switch from.Status.LatestVersion {
	case 0:
		return nil, newInvalidError(rollback, "cannot rollback an undeployed config")
	case 1:
		return nil, newInvalidError(rollback, fmt.Sprintf("no previous deployment exists for %q", appsutil.LabelForDeploymentConfig(from)))
	case rollback.Spec.Revision:
		return nil, newInvalidError(rollback, fmt.Sprintf("version %d is already the latest", rollback.Spec.Revision))
	}

	revision := from.Status.LatestVersion - 1
	if rollback.Spec.Revision > 0 {
		revision = rollback.Spec.Revision
	}

	// Find the target deployment and decode its config.
	name := appsutil.DeploymentNameForConfigVersion(from.Name, revision)
	targetDeployment, err := r.rn.ReplicationControllers(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, newInvalidError(rollback, err.Error())
	}

	fromConfig, err := appsserialization.DecodeDeploymentConfig(targetDeployment)
	if err != nil {
		return nil, newInvalidError(rollback, fmt.Sprintf("couldn't decode deployment config from deployment: %v", err))
	}

	toConfig := &appsapi.DeploymentConfig{}
	if err := v1.Convert_v1_DeploymentConfig_To_apps_DeploymentConfig(fromConfig, toConfig, nil); err != nil {
		return nil, apierrors.NewInternalError(err)
	}

	if from.Annotations == nil && len(rollback.UpdatedAnnotations) > 0 {
		from.Annotations = make(map[string]string)
	}
	for key, value := range rollback.UpdatedAnnotations {
		from.Annotations[key] = value
	}

	fromInternal := &appsapi.DeploymentConfig{}
	if err := v1.Convert_v1_DeploymentConfig_To_apps_DeploymentConfig(from, fromInternal, nil); err != nil {
		return nil, apierrors.NewInternalError(err)
	}

	return r.generator.GenerateRollback(fromInternal, toConfig, &rollback.Spec)
}

func newInvalidError(rollback *appsapi.DeploymentConfigRollback, reason string) error {
	err := field.Invalid(field.NewPath("name"), rollback.Name, reason)
	return apierrors.NewInvalid(apps.Kind("DeploymentConfigRollback"), rollback.Name, field.ErrorList{err})
}
