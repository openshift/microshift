package controllers

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/openshift/microshift/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

const (
	trustedCABundlePath  = "/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem"
	trustedCABundleLabel = "config.openshift.io/inject-trusted-cabundle"
	trustedCABundleKey   = "ca-bundle.crt"
	resyncInterval       = 30 * time.Second
)

type TrustedCABundleManager struct {
	kubeconfig string
}

func NewTrustedCABundleManager(cfg *config.Config) *TrustedCABundleManager {
	return &TrustedCABundleManager{
		kubeconfig: cfg.KubeConfigPath(config.KubeAdmin),
	}
}

func (s *TrustedCABundleManager) Name() string { return "trusted-ca-bundle-manager" }
func (s *TrustedCABundleManager) Dependencies() []string {
	return []string{"kube-apiserver", "infrastructure-services-manager"}
}

func (s *TrustedCABundleManager) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)

	restCfg, err := clientcmd.BuildConfigFromFlags("", s.kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to build kubeconfig: %w", err)
	}
	client, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	if err := s.syncTrustedCABundle(ctx, client); err != nil {
		klog.Warningf("%s failed initial trusted CA bundle sync: %v", s.Name(), err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	defer func() {
		if cerr := watcher.Close(); cerr != nil {
			klog.Errorf("%s failed to close file watcher: %v", s.Name(), cerr)
		}
	}()

	for _, path := range []string{trustedCABundlePath, filepath.Dir(trustedCABundlePath)} {
		if err := watcher.Add(path); err != nil {
			klog.Warningf("%s failed to watch %s: %v", s.Name(), path, err)
		}
	}

	close(ready)
	klog.Infof("%s is ready, watching %s", s.Name(), trustedCABundlePath)

	lastHash, _ := s.getFileHash()
	resyncTicker := time.NewTicker(resyncInterval)
	defer resyncTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-resyncTicker.C:
			if err := s.syncTrustedCABundle(ctx, client); err != nil {
				klog.V(4).Infof("%s periodic sync: %v", s.Name(), err)
			}

		case event, ok := <-watcher.Events:
			if !ok {
				return fmt.Errorf("watcher channel closed")
			}
			if event.Name != trustedCABundlePath {
				continue
			}
			if event.Op&fsnotify.Write == 0 && event.Op&fsnotify.Create == 0 {
				continue
			}
			currentHash, err := s.getFileHash()
			if err != nil {
				klog.Warningf("%s failed to hash CA bundle: %v", s.Name(), err)
				continue
			}
			if currentHash == lastHash {
				continue
			}
			klog.Infof("%s detected CA bundle change, syncing", s.Name())
			if err := s.syncTrustedCABundle(ctx, client); err != nil {
				klog.Errorf("%s failed to sync CA bundle: %v", s.Name(), err)
				continue
			}
			lastHash = currentHash

		case err, ok := <-watcher.Errors:
			if !ok {
				return fmt.Errorf("watcher error channel closed")
			}
			klog.Errorf("%s watcher error: %v", s.Name(), err)
		}
	}
}

func (s *TrustedCABundleManager) syncTrustedCABundle(ctx context.Context, client kubernetes.Interface) error {
	caBundle, err := os.ReadFile(trustedCABundlePath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", trustedCABundlePath, err)
	}
	if len(caBundle) == 0 {
		return fmt.Errorf("%s is empty", trustedCABundlePath)
	}

	configMaps, err := client.CoreV1().ConfigMaps("").List(ctx, metav1.ListOptions{
		LabelSelector: trustedCABundleLabel + "=true",
	})
	if err != nil {
		return fmt.Errorf("failed to list ConfigMaps with label %s: %w", trustedCABundleLabel, err)
	}

	caBundleStr := string(caBundle)
	for i := range configMaps.Items {
		cm := &configMaps.Items[i]
		if cm.Data == nil {
			cm.Data = map[string]string{}
		}
		if cm.Data[trustedCABundleKey] == caBundleStr {
			klog.V(4).Infof("%s ConfigMap %s/%s already has current CA bundle", s.Name(), cm.Namespace, cm.Name)
			continue
		}
		cm.Data[trustedCABundleKey] = caBundleStr
		if _, err := client.CoreV1().ConfigMaps(cm.Namespace).Update(ctx, cm, metav1.UpdateOptions{}); err != nil {
			klog.Errorf("%s failed to update ConfigMap %s/%s: %v", s.Name(), cm.Namespace, cm.Name, err)
			continue
		}
		klog.Infof("%s injected trusted CA bundle into ConfigMap %s/%s", s.Name(), cm.Namespace, cm.Name)
	}

	return nil
}

func (s *TrustedCABundleManager) getFileHash() (string, error) {
	content, err := os.ReadFile(trustedCABundlePath)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(content)
	return fmt.Sprintf("%x", hash), nil
}
