package controllers

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	informers "k8s.io/client-go/informers/core/v1"
	kclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// DockercfgTokenDeletedControllerOptions contains options for the DockercfgTokenDeletedController
type DockercfgTokenDeletedControllerOptions struct {
	// Resync is the time.Duration at which to fully re-list secrets.
	// If zero, re-list will be delayed as long as possible
	Resync time.Duration
}

// NewDockercfgTokenDeletedController returns a new *DockercfgTokenDeletedController.
func NewDockercfgTokenDeletedController(secrets informers.SecretInformer, cl kclientset.Interface, options DockercfgTokenDeletedControllerOptions) *DockercfgTokenDeletedController {
	e := &DockercfgTokenDeletedController{
		client: cl,
	}

	e.secretController = secrets.Informer().GetController()
	secrets.Informer().AddEventHandler(
		cache.FilteringResourceEventHandler{
			FilterFunc: func(obj interface{}) bool {
				switch t := obj.(type) {
				case *v1.Secret:
					return t.Type == v1.SecretTypeServiceAccountToken
				default:
					utilruntime.HandleError(fmt.Errorf("object passed to %T that is not expected: %T", e, obj))
					return false
				}
			},
			Handler: cache.ResourceEventHandlerFuncs{
				DeleteFunc: e.secretDeleted,
			},
		},
	)

	return e
}

// The DockercfgTokenDeletedController watches for service account tokens to be deleted.
// On delete, it removes the associated dockercfg secret if it exists.
type DockercfgTokenDeletedController struct {
	client           kclientset.Interface
	secretController cache.Controller
}

// Runs controller loops and returns on shutdown
func (e *DockercfgTokenDeletedController) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	klog.Infof("Starting DockercfgTokenDeletedController controller")
	defer klog.Infof("Shutting down DockercfgTokenDeletedController controller")

	// Wait for the stores to fill
	if !cache.WaitForCacheSync(stopCh, e.secretController.HasSynced) {
		return
	}
	klog.V(1).Infof("caches synced")

	<-stopCh
}

// secretDeleted reacts to a token secret being deleted by looking for a corresponding dockercfg secret and deleting it if it exists
func (e *DockercfgTokenDeletedController) secretDeleted(obj interface{}) {
	tokenSecret, ok := obj.(*v1.Secret)
	if !ok {
		return
	}
	dockercfgSecrets, err := findDockercfgSecrets(e.client, tokenSecret)
	if err != nil {
		klog.Error(err)
		return
	}
	if len(dockercfgSecrets) == 0 {
		return
	}
	// remove the reference token secrets
	for _, dockercfgSecret := range dockercfgSecrets {
		if metav1.IsControlledBy(dockercfgSecret, tokenSecret) {
			// If the docker pull secret is owned by its associated token, let garbage collection take care of it.
			klog.V(5).Infof("Ignoring deletion of pull secret %s/%s because it should be removed via garbage collection", dockercfgSecret.Namespace, dockercfgSecret.Name)
			continue
		}
		klog.V(4).Infof("Deleting pull secret %s/%s because its associated token %s/%s has been deleted", dockercfgSecret.Namespace, dockercfgSecret.Name, tokenSecret.Namespace, tokenSecret.Name)
		if err := e.client.CoreV1().Secrets(dockercfgSecret.Namespace).Delete(context.TODO(), dockercfgSecret.Name, metav1.DeleteOptions{}); (err != nil) && !apierrors.IsNotFound(err) {
			utilruntime.HandleError(err)
		}
	}
}
