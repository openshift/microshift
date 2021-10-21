package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	kapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	informers "k8s.io/client-go/informers/core/v1"
	kclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/credentialprovider"
	"k8s.io/kubernetes/pkg/registry/core/secret"

	"github.com/openshift/library-go/pkg/build/naming"
)

const (
	ServiceAccountTokenSecretNameKey = "openshift.io/token-secret.name"
	MaxRetriesBeforeResync           = 5

	// ServiceAccountTokenValueAnnotation stores the actual value of the token so that a dockercfg secret can be
	// made without having a value dockerURL
	ServiceAccountTokenValueAnnotation = "openshift.io/token-secret.value"

	// CreateDockercfgSecretsController is the name of this controller that should be
	// attached to all token secrets this controller create
	CreateDockercfgSecretsController = "openshift.io/create-dockercfg-secrets"

	// PendingTokenAnnotation contains the name of the token secret that is waiting for the
	// token data population
	PendingTokenAnnotation = "openshift.io/create-dockercfg-secrets.pending-token"

	// DeprecatedKubeCreatedByAnnotation was removed by https://github.com/kubernetes/kubernetes/pull/54445 (liggitt approved).
	DeprecatedKubeCreatedByAnnotation = "kubernetes.io/created-by"

	// These constants are here to create a name that is short enough to survive chopping by generate name
	maxNameLength             = 63
	randomLength              = 5
	maxSecretPrefixNameLength = maxNameLength - randomLength
)

// DockercfgControllerOptions contains options for the DockercfgController
type DockercfgControllerOptions struct {
	// Resync is the time.Duration at which to fully re-list service accounts.
	// If zero, re-list will be delayed as long as possible
	Resync time.Duration

	// DockerURLsInitialized is used to send a signal to this controller that it has the correct set of docker urls
	// This is normally signaled from the DockerRegistryServiceController which watches for updates to the internal
	// container image registry service.
	DockerURLsInitialized chan struct{}
}

// NewDockercfgController returns a new *DockercfgController.
func NewDockercfgController(serviceAccounts informers.ServiceAccountInformer, secrets informers.SecretInformer, cl kclientset.Interface, options DockercfgControllerOptions) *DockercfgController {
	e := &DockercfgController{
		client:                cl,
		queue:                 workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "serviceaccount-create-dockercfg"),
		dockerURLsInitialized: options.DockerURLsInitialized,
	}

	serviceAccountCache := serviceAccounts.Informer().GetStore()
	e.serviceAccountController = serviceAccounts.Informer().GetController()
	serviceAccounts.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				serviceAccount := obj.(*v1.ServiceAccount)
				klog.V(5).Infof("Adding service account %s", serviceAccount.Name)
				e.enqueueServiceAccount(serviceAccount)
			},
			UpdateFunc: func(old, cur interface{}) {
				serviceAccount := cur.(*v1.ServiceAccount)
				klog.V(5).Infof("Updating service account %s", serviceAccount.Name)
				// Resync on service object relist.
				e.enqueueServiceAccount(serviceAccount)
			},
		},
	)
	e.serviceAccountCache = NewEtcdMutationCache(serviceAccountCache)

	e.secretCache = secrets.Informer().GetIndexer()
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
				AddFunc:    func(cur interface{}) { e.handleTokenSecretUpdate(nil, cur) },
				UpdateFunc: func(old, cur interface{}) { e.handleTokenSecretUpdate(old, cur) },
				DeleteFunc: e.handleTokenSecretDelete,
			},
		},
	)
	e.syncHandler = e.syncServiceAccount

	return e
}

// DockercfgController manages dockercfg secrets for ServiceAccount objects
type DockercfgController struct {
	client kclientset.Interface

	dockerURLLock         sync.Mutex
	dockerURLs            []string
	dockerURLsInitialized chan struct{}

	serviceAccountCache      MutationCache
	serviceAccountController cache.Controller
	secretCache              cache.Store
	secretController         cache.Controller

	queue workqueue.RateLimitingInterface

	// syncHandler does the work. It's factored out for unit testing
	syncHandler func(serviceKey string) error
}

// handleTokenSecretUpdate checks if the service account token secret is populated with
// token data and triggers re-sync of service account when the data are observed.
func (e *DockercfgController) handleTokenSecretUpdate(oldObj, newObj interface{}) {
	secret := newObj.(*v1.Secret)
	if secret.Annotations[DeprecatedKubeCreatedByAnnotation] != CreateDockercfgSecretsController {
		return
	}
	isPopulated := len(secret.Data[v1.ServiceAccountTokenKey]) > 0

	wasPopulated := false
	if oldObj != nil {
		oldSecret := oldObj.(*v1.Secret)
		wasPopulated = len(oldSecret.Data[v1.ServiceAccountTokenKey]) > 0
		klog.V(5).Infof("Updating token secret %s/%s", secret.Namespace, secret.Name)
	} else {
		klog.V(5).Infof("Adding token secret %s/%s", secret.Namespace, secret.Name)
	}

	if !wasPopulated && isPopulated {
		e.enqueueServiceAccountForToken(secret)
	}
}

// handleTokenSecretDelete handles token secrets deletion and re-sync the service account
// which will cause a token to be re-created.
func (e *DockercfgController) handleTokenSecretDelete(obj interface{}) {
	secret, isSecret := obj.(*v1.Secret)
	if !isSecret {
		tombstone, objIsTombstone := obj.(cache.DeletedFinalStateUnknown)
		if !objIsTombstone {
			klog.V(2).Infof("Expected tombstone object when deleting token, got %v", obj)
			return
		}
		secret, isSecret = tombstone.Obj.(*v1.Secret)
		if !isSecret {
			klog.V(2).Infof("Expected tombstone object to contain secret, got: %v", obj)
			return
		}
	}
	if secret.Annotations[DeprecatedKubeCreatedByAnnotation] != CreateDockercfgSecretsController {
		return
	}
	if len(secret.Data[v1.ServiceAccountTokenKey]) > 0 {
		// Let deleted_token_secrets handle deletion of populated tokens
		return
	}
	e.enqueueServiceAccountForToken(secret)
}

func (e *DockercfgController) enqueueServiceAccountForToken(tokenSecret *v1.Secret) {
	serviceAccount := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tokenSecret.Annotations[v1.ServiceAccountNameKey],
			Namespace: tokenSecret.Namespace,
			UID:       types.UID(tokenSecret.Annotations[v1.ServiceAccountUIDKey]),
		},
	}
	key, err := controller.KeyFunc(serviceAccount)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("error syncing token secret %s/%s: %v", tokenSecret.Namespace, tokenSecret.Name, err))
		return
	}
	e.queue.Add(key)
}

func (e *DockercfgController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer e.queue.ShutDown()

	klog.Infof("Starting DockercfgController controller")
	defer klog.Infof("Shutting down DockercfgController controller")

	// Wait for the store to sync before starting any work in this controller.
	ready := make(chan struct{})
	go e.waitForDockerURLs(ready, stopCh)
	select {
	case <-ready:
	case <-stopCh:
		return
	}
	klog.V(1).Infof("urls found")

	// Wait for the stores to fill
	if !cache.WaitForCacheSync(stopCh, e.serviceAccountController.HasSynced, e.secretController.HasSynced) {
		return
	}
	klog.V(1).Infof("caches synced")

	for i := 0; i < workers; i++ {
		go wait.Until(e.worker, time.Second, stopCh)
	}
	<-stopCh
}

func (c *DockercfgController) waitForDockerURLs(ready chan<- struct{}, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()

	// wait for the initialization to complete to be informed of a stop
	select {
	case <-c.dockerURLsInitialized:
	case <-stopCh:
		return
	}

	close(ready)
}

func (e *DockercfgController) enqueueServiceAccount(serviceAccount *v1.ServiceAccount) {
	if !needsDockercfgSecret(serviceAccount) {
		return
	}

	key, err := controller.KeyFunc(serviceAccount)
	if err != nil {
		klog.Errorf("Couldn't get key for object %+v: %v", serviceAccount, err)
		return
	}

	e.queue.Add(key)
}

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
// It enforces that the syncHandler is never invoked concurrently with the same key.
func (e *DockercfgController) worker() {
	for {
		if !e.work() {
			return
		}
	}
}

// work returns true if the worker thread should continue
func (e *DockercfgController) work() bool {
	key, quit := e.queue.Get()
	if quit {
		return false
	}
	defer e.queue.Done(key)

	if err := e.syncHandler(key.(string)); err == nil {
		// this means the request was successfully handled.  We should "forget" the item so that any retry
		// later on is reset
		e.queue.Forget(key)

	} else {
		// if we had an error it means that we didn't handle it, which means that we want to requeue the work
		if e.queue.NumRequeues(key) > MaxRetriesBeforeResync {
			utilruntime.HandleError(fmt.Errorf("error syncing service, it will be tried again on a resync %v: %v", key, err))
			e.queue.Forget(key)
		} else {
			klog.V(4).Infof("error syncing service, it will be retried %v: %v", key, err)
			e.queue.AddRateLimited(key)
		}
	}

	return true
}

func (e *DockercfgController) SetDockerURLs(newDockerURLs ...string) {
	e.dockerURLLock.Lock()
	defer e.dockerURLLock.Unlock()

	e.dockerURLs = newDockerURLs
}

func needsDockercfgSecret(serviceAccount *v1.ServiceAccount) bool {
	mountableDockercfgSecrets, imageDockercfgPullSecrets := getGeneratedDockercfgSecretNames(serviceAccount)

	// look for an ImagePullSecret in the form
	if len(imageDockercfgPullSecrets) > 0 && len(mountableDockercfgSecrets) > 0 {
		return false
	}

	return true
}

func (e *DockercfgController) syncServiceAccount(key string) error {
	obj, exists, err := e.serviceAccountCache.GetByKey(key)
	if err != nil {
		klog.V(4).Infof("Unable to retrieve service account %v from store: %v", key, err)
		return err
	}
	if !exists {
		klog.V(4).Infof("Service account has been deleted %v", key)
		return nil
	}
	if !needsDockercfgSecret(obj.(*v1.ServiceAccount)) {
		return e.syncDockercfgOwnerRefs(obj.(*v1.ServiceAccount))
	}

	serviceAccount := obj.(*v1.ServiceAccount).DeepCopyObject().(*v1.ServiceAccount)

	mountableDockercfgSecrets, imageDockercfgPullSecrets := getGeneratedDockercfgSecretNames(serviceAccount)

	// If we have a pull secret in one list, use it for the other.  It must only be in one list because
	// otherwise we wouldn't "needsDockercfgSecret"
	foundPullSecret := len(imageDockercfgPullSecrets) > 0
	foundMountableSecret := len(mountableDockercfgSecrets) > 0
	if foundPullSecret || foundMountableSecret {
		switch {
		case foundPullSecret:
			serviceAccount.Secrets = append(serviceAccount.Secrets, v1.ObjectReference{Name: imageDockercfgPullSecrets.List()[0]})
		case foundMountableSecret:
			serviceAccount.ImagePullSecrets = append(serviceAccount.ImagePullSecrets, v1.LocalObjectReference{Name: mountableDockercfgSecrets.List()[0]})
		}
		// Clear the pending token annotation when updating
		delete(serviceAccount.Annotations, PendingTokenAnnotation)

		updatedSA, err := e.client.CoreV1().ServiceAccounts(serviceAccount.Namespace).Update(context.TODO(), serviceAccount, metav1.UpdateOptions{})
		if err == nil {
			e.serviceAccountCache.Mutation(updatedSA)
		}
		return err
	}

	dockercfgSecret, created, err := e.createDockerPullSecret(serviceAccount)
	if err != nil {
		return err
	}
	if !created {
		klog.V(5).Infof("The dockercfg secret was not created for service account %s/%s, will retry", serviceAccount.Namespace, serviceAccount.Name)
		return nil
	}

	first := true
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		if !first {
			obj, exists, err := e.serviceAccountCache.GetByKey(key)
			if err != nil {
				return err
			}
			if !exists || !needsDockercfgSecret(obj.(*v1.ServiceAccount)) || serviceAccount.UID != obj.(*v1.ServiceAccount).UID {
				// somehow a dockercfg secret appeared or the SA disappeared.  cleanup the secret we made and return
				klog.V(2).Infof("Deleting secret because the work is already done %s/%s", dockercfgSecret.Namespace, dockercfgSecret.Name)
				e.client.CoreV1().Secrets(dockercfgSecret.Namespace).Delete(context.TODO(), dockercfgSecret.Name, metav1.DeleteOptions{})
				return nil
			}

			serviceAccount = obj.(*v1.ServiceAccount).DeepCopyObject().(*v1.ServiceAccount)
		}
		first = false

		serviceAccount.Secrets = append(serviceAccount.Secrets, v1.ObjectReference{Name: dockercfgSecret.Name})
		serviceAccount.ImagePullSecrets = append(serviceAccount.ImagePullSecrets, v1.LocalObjectReference{Name: dockercfgSecret.Name})
		// Clear the pending token annotation when updating
		delete(serviceAccount.Annotations, PendingTokenAnnotation)

		updatedSA, err := e.client.CoreV1().ServiceAccounts(serviceAccount.Namespace).Update(context.TODO(), serviceAccount, metav1.UpdateOptions{})
		if err == nil {
			e.serviceAccountCache.Mutation(updatedSA)
		}
		return err
	})

	if err != nil {
		// nothing to do.  Our choice was stale or we got a conflict.  Either way that means that the service account was updated.  We simply need to return because we'll get an update notification later
		// we do need to clean up our dockercfgSecret.  token secrets are cleaned up by the controller handling service account dockercfg secret deletes
		klog.V(2).Infof("Deleting secret %s/%s (err=%v)", dockercfgSecret.Namespace, dockercfgSecret.Name, err)
		e.client.CoreV1().Secrets(dockercfgSecret.Namespace).Delete(context.TODO(), dockercfgSecret.Name, metav1.DeleteOptions{})
	}
	return err
}

// createTokenSecret creates a token secret for a given service account.  Returns the name of the token
func (e *DockercfgController) createTokenSecret(serviceAccount *v1.ServiceAccount) (*v1.Secret, bool, error) {
	pendingTokenName := serviceAccount.Annotations[PendingTokenAnnotation]

	// If this service account has no record of a pending token name, record one
	if len(pendingTokenName) == 0 {
		pendingTokenName = secret.Strategy.GenerateName(getTokenSecretNamePrefix(serviceAccount.Name))
		if serviceAccount.Annotations == nil {
			serviceAccount.Annotations = map[string]string{}
		}
		serviceAccount.Annotations[PendingTokenAnnotation] = pendingTokenName
		updatedServiceAccount, err := e.client.CoreV1().ServiceAccounts(serviceAccount.Namespace).Update(context.TODO(), serviceAccount, metav1.UpdateOptions{})
		// Conflicts mean we'll get called to sync this service account again
		if kapierrors.IsConflict(err) {
			return nil, false, nil
		}
		if err != nil {
			return nil, false, err
		}
		serviceAccount = updatedServiceAccount
	}

	// Return the token from cache
	existingTokenSecretObj, exists, err := e.secretCache.GetByKey(serviceAccount.Namespace + "/" + pendingTokenName)
	if err != nil {
		return nil, false, err
	}
	if exists {
		existingTokenSecret := existingTokenSecretObj.(*v1.Secret)
		return existingTokenSecret, len(existingTokenSecret.Data[v1.ServiceAccountTokenKey]) > 0, nil
	}

	// Try to create the named pending token
	tokenSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pendingTokenName,
			Namespace: serviceAccount.Namespace,
			Annotations: map[string]string{
				v1.ServiceAccountNameKey:          serviceAccount.Name,
				v1.ServiceAccountUIDKey:           string(serviceAccount.UID),
				DeprecatedKubeCreatedByAnnotation: CreateDockercfgSecretsController,
			},
		},
		Type: v1.SecretTypeServiceAccountToken,
		Data: map[string][]byte{},
	}

	klog.V(4).Infof("Creating token secret %q for service account %s/%s", tokenSecret.Name, serviceAccount.Namespace, serviceAccount.Name)
	token, err := e.client.CoreV1().Secrets(tokenSecret.Namespace).Create(context.TODO(), tokenSecret, metav1.CreateOptions{})
	// Already exists but not in cache means we'll get an add watch event and resync
	if kapierrors.IsAlreadyExists(err) {
		return nil, false, nil
	}
	// If we cannot create this secret because the namespace it is being terminated isn't a thing we should fail and requeue a retry.
	// Instead, we know that when a new namespace gets created, the serviceaccount will be recreated and we'll get a second shot at
	// processing the serviceaccount.
	if kapierrors.HasStatusCause(err, v1.NamespaceTerminatingCause) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return token, len(token.Data[v1.ServiceAccountTokenKey]) > 0, nil
}

// createDockerPullSecret creates a dockercfg secret based on the token secret
func (e *DockercfgController) createDockerPullSecret(serviceAccount *v1.ServiceAccount) (*v1.Secret, bool, error) {
	tokenSecret, isPopulated, err := e.createTokenSecret(serviceAccount)
	if err != nil {
		return nil, false, err
	}
	if !isPopulated {
		klog.V(5).Infof("Token secret for service account %s/%s is not populated yet", serviceAccount.Namespace, serviceAccount.Name)
		return nil, false, nil
	}

	dockercfgSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secret.Strategy.GenerateName(getDockercfgSecretNamePrefix(serviceAccount.Name)),
			Namespace: tokenSecret.Namespace,
			Annotations: map[string]string{
				v1.ServiceAccountNameKey:           serviceAccount.Name,
				v1.ServiceAccountUIDKey:            string(serviceAccount.UID),
				ServiceAccountTokenSecretNameKey:   string(tokenSecret.Name),
				ServiceAccountTokenValueAnnotation: string(tokenSecret.Data[v1.ServiceAccountTokenKey]),
			},
		},
		Type: v1.SecretTypeDockercfg,
		Data: map[string][]byte{},
	}
	klog.V(4).Infof("Creating dockercfg secret %q for service account %s/%s", dockercfgSecret.Name, serviceAccount.Namespace, serviceAccount.Name)

	// prevent updating the DockerURL until we've created the secret
	e.dockerURLLock.Lock()
	defer e.dockerURLLock.Unlock()

	dockercfg := credentialprovider.DockerConfig{}
	for _, dockerURL := range e.dockerURLs {
		dockercfg[dockerURL] = credentialprovider.DockerConfigEntry{
			Username: "serviceaccount",
			Password: string(tokenSecret.Data[v1.ServiceAccountTokenKey]),
			Email:    "serviceaccount@example.org",
		}
	}
	dockercfgContent, err := json.Marshal(&dockercfg)
	if err != nil {
		return nil, false, err
	}
	dockercfgSecret.Data[v1.DockerConfigKey] = dockercfgContent
	blockDeletion := false
	ownerRef := metav1.NewControllerRef(tokenSecret, v1.SchemeGroupVersion.WithKind("Secret"))
	ownerRef.BlockOwnerDeletion = &blockDeletion
	dockercfgSecret.SetOwnerReferences([]metav1.OwnerReference{*ownerRef})

	// Save the secret
	createdSecret, err := e.client.CoreV1().Secrets(tokenSecret.Namespace).Create(context.TODO(), dockercfgSecret, metav1.CreateOptions{})
	// If we cannot create this secret because the namespace it is being terminated isn't a thing we should fail and requeue a retry.
	// Instead, we know that when a new namespace gets created, the serviceaccount will be recreated and we'll get a second shot at
	// processing the serviceaccount.
	if kapierrors.HasStatusCause(err, v1.NamespaceTerminatingCause) {
		return nil, false, nil
	}
	return createdSecret, err == nil, err
}

func (e *DockercfgController) syncDockercfgOwnerRefs(serviceAccount *v1.ServiceAccount) error {
	for _, secretRef := range serviceAccount.Secrets {
		secret, exists, err := e.secretCache.GetByKey(fmt.Sprintf("%s/%s", secretRef.Namespace, secretRef.Name))
		if err != nil {
			return err
		}
		if !exists {
			continue
		}
		err = e.syncDockercfgOwner(secret.(*v1.Secret))
		if err != nil {
			return err
		}
	}
	for _, secretRef := range serviceAccount.ImagePullSecrets {
		secret, exists, err := e.secretCache.GetByKey(fmt.Sprintf("%s/%s", serviceAccount.Namespace, secretRef.Name))
		if err != nil {
			return err
		}
		if !exists {
			continue
		}
		err = e.syncDockercfgOwner(secret.(*v1.Secret))
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *DockercfgController) syncDockercfgOwner(pullSecret *v1.Secret) error {
	if pullSecret.Type != v1.SecretTypeDockercfg {
		return nil
	}
	tokenName := pullSecret.Annotations[ServiceAccountTokenSecretNameKey]
	// If there is no token name, this pull secret was likely linked by a user.
	// No further work is needed.
	if len(tokenName) == 0 {
		return nil
	}
	// Make sure the token exists
	tokenSecret, exists, err := e.secretCache.GetByKey(fmt.Sprintf("%s/%s", pullSecret.Namespace, tokenName))
	if err != nil {
		return err
	}
	if !exists {
		// If the pull secret exists, and the token does not, delete the pull secret.
		klog.V(4).Infof("Deleting pull secret %s/%s because its associated token %s/%s is missing", pullSecret.Namespace, pullSecret.Name, pullSecret.Namespace, tokenName)
		return e.client.CoreV1().Secrets(pullSecret.Namespace).Delete(context.TODO(), pullSecret.Name, metav1.DeleteOptions{})
	}
	tokenSecretObj := tokenSecret.(*v1.Secret)
	// If the pull secret has an owner reference to its associated token, no further work is needed.
	if metav1.IsControlledBy(pullSecret, tokenSecretObj) {
		return nil
	}
	pullSecret = pullSecret.DeepCopy()
	blockDeletion := false
	ownerRef := metav1.NewControllerRef(tokenSecretObj, v1.SchemeGroupVersion.WithKind("Secret"))
	ownerRef.BlockOwnerDeletion = &blockDeletion
	pullSecret.SetOwnerReferences([]metav1.OwnerReference{*ownerRef})
	klog.V(4).Infof("Adding token %s/%s as the owner of pull secret %s/%s", pullSecret.Namespace, tokenName, pullSecret.Namespace, pullSecret.Name)
	_, err = e.client.CoreV1().Secrets(pullSecret.Namespace).Update(context.TODO(), pullSecret, metav1.UpdateOptions{})
	return err
}

func getGeneratedDockercfgSecretNames(serviceAccount *v1.ServiceAccount) (sets.String, sets.String) {
	mountableDockercfgSecrets := sets.String{}
	imageDockercfgPullSecrets := sets.String{}

	secretNamePrefix := getDockercfgSecretNamePrefix(serviceAccount.Name)

	for _, s := range serviceAccount.Secrets {
		if strings.HasPrefix(s.Name, secretNamePrefix) {
			mountableDockercfgSecrets.Insert(s.Name)
		}
	}
	for _, s := range serviceAccount.ImagePullSecrets {
		if strings.HasPrefix(s.Name, secretNamePrefix) {
			imageDockercfgPullSecrets.Insert(s.Name)
		}
	}
	return mountableDockercfgSecrets, imageDockercfgPullSecrets
}

func getDockercfgSecretNamePrefix(serviceAccountName string) string {
	return naming.GetName(serviceAccountName, "dockercfg-", maxSecretPrefixNameLength)
}

// getTokenSecretNamePrefix creates the prefix used for the generated SA token secret. This is compatible with kube up until
// long names, at which point we hash the SA name and leave the "token-" intact.  Upstream clips the value and generates a random
// string.
func getTokenSecretNamePrefix(serviceAccountName string) string {
	return naming.GetName(serviceAccountName, "token-", maxSecretPrefixNameLength)
}
