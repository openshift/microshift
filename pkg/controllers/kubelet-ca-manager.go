/*
Copyright © 2025 MicroShift Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package controllers

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const (
	kubeletCAConfigMapName      = "kubelet-client-ca"
	kubeletCAConfigMapNamespace = "kube-system"
	kubeletCAFileName           = "ca.crt"
	defaultInformerResyncPeriod = 10 * time.Minute
)

type KubeletCAManager struct {
	cfg        *config.Config
	client     *kubernetes.Clientset
	queue      workqueue.TypedRateLimitingInterface[string]
	informer   cache.SharedIndexInformer
	caCertData map[string]string
}

func NewKubeletCAManager(cfg *config.Config) *KubeletCAManager {
	return &KubeletCAManager{
		cfg: cfg,
	}
}

func (s *KubeletCAManager) Name() string { return "kubelet-ca-manager" }
func (s *KubeletCAManager) Dependencies() []string {
	return []string{"kube-apiserver"}
}

func (s *KubeletCAManager) restConfig() (*rest.Config, error) {
	kubeConfigPath := s.cfg.KubeConfigPath(config.KubeAdmin)
	return clientcmd.BuildConfigFromFlags("", kubeConfigPath)
}

func (s *KubeletCAManager) loadCACertData() error {
	certsDir := cryptomaterial.CertsDirectory(config.DataDir)
	kubeletCAPath := cryptomaterial.KubeletClientCAPath(certsDir)

	caCertPEM, err := os.ReadFile(kubeletCAPath)
	if err != nil {
		return fmt.Errorf("failed to read kubelet client CA file %s: %v", kubeletCAPath, err)
	}

	s.caCertData = map[string]string{
		kubeletCAFileName: string(caCertPEM),
	}
	return nil
}

func (s *KubeletCAManager) ensureConfigMap(ctx context.Context) error {
	var cm = "core/kubelet-client-ca.yaml"
	kubeConfigPath := s.cfg.KubeConfigPath(config.KubeAdmin)

	if err := assets.ApplyConfigMapWithData(ctx, cm, s.caCertData, kubeConfigPath); err != nil {
		return fmt.Errorf("failed to apply configMap %v: %v", cm, err)
	}
	return nil
}

func (s *KubeletCAManager) processNextItem(ctx context.Context) bool {
	key, quit := s.queue.Get()
	if quit {
		return false
	}
	defer s.queue.Done(key)

	err := s.syncConfigMap(ctx, key)
	if err != nil {
		s.queue.AddRateLimited(key)
		klog.Errorf("failed to sync configmap %v: %v", key, err)
		return true
	}

	s.queue.Forget(key)
	return true
}

func (s *KubeletCAManager) syncConfigMap(ctx context.Context, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return fmt.Errorf("invalid resource key: %s", key)
	}

	configMap, err := s.client.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			klog.Infof("configmap %s/%s was deleted, recreating with default CA data", namespace, name)
			return s.ensureConfigMap(ctx)
		}
		return fmt.Errorf("failed to get ConfigMap %s/%s: %v", namespace, name, err)
	}

	if configMap.Data[kubeletCAFileName] != s.caCertData[kubeletCAFileName] {
		klog.Infof("configmap %s/%s data has been tampered with, restoring default CA data", namespace, name)
		return s.ensureConfigMap(ctx)
	}

	return nil
}

func (s *KubeletCAManager) runWorker(ctx context.Context) {
	for s.processNextItem(ctx) {
	}
}

func (s *KubeletCAManager) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)

	if err := s.loadCACertData(); err != nil {
		klog.Errorf("failed to load CA certificate data: %v", err)
		return err
	}

	restCfg, err := s.restConfig()
	if err != nil {
		return fmt.Errorf("failed to create rest config for kubelet CA manager: %w", err)
	}
	s.client, err = kubernetes.NewForConfig(restCfg)
	if err != nil {
		return fmt.Errorf("failed to create clientset for kubelet CA manager: %w", err)
	}

	if err := s.ensureConfigMap(ctx); err != nil {
		klog.Errorf("failed to create initial ConfigMap: %v", err)
		return err
	}
	klog.Infof("Successfully created initial kubelet client CA ConfigMap")

	stopCh := make(chan struct{})
	defer close(stopCh)

	factory := informers.NewSharedInformerFactoryWithOptions(
		s.client,
		defaultInformerResyncPeriod,
		informers.WithNamespace(kubeletCAConfigMapNamespace),
	)

	configMapInformer := factory.Core().V1().ConfigMaps()
	s.informer = configMapInformer.Informer()
	s.queue = workqueue.NewTypedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[string]())

	_, err = s.informer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			switch cm := obj.(type) {
			case *corev1.ConfigMap:
				return cm.Name == kubeletCAConfigMapName && cm.Namespace == kubeletCAConfigMapNamespace
			case cache.DeletedFinalStateUnknown:
				if deletedCM, ok := cm.Obj.(*corev1.ConfigMap); ok {
					return deletedCM.Name == kubeletCAConfigMapName && deletedCM.Namespace == kubeletCAConfigMapNamespace
				}
			}
			return false
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(obj)
				if err == nil {
					s.queue.Add(key)
				}
			},
			UpdateFunc: func(oldObj interface{}, newObj interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(newObj)
				if err == nil {
					s.queue.Add(key)
				}
			},
			DeleteFunc: func(obj interface{}) {
				key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
				if err == nil {
					s.queue.Add(key)
				}
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to initialize informer event handlers: %w", err)
	}

	factory.Start(stopCh)

	if !cache.WaitForCacheSync(stopCh, s.informer.HasSynced) {
		return fmt.Errorf("timed out waiting for caches to sync")
	}

	go func() {
		defer func() {
			s.queue.ShutDown()
		}()
		s.runWorker(ctx)
	}()

	close(ready)

	<-ctx.Done()

	return ctx.Err()
}
