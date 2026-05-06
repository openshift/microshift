package controllers

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

type fileWatcherConfig struct {
	serviceName        string
	dependencies       []string
	file               string
	kubeconfig         string
	enabled            bool
	configMapNamespace string
	configMapName      string
	configMapDataKey   string
	labels             map[string]string
	annotations        map[string]string
	eventMask          fsnotify.Op
	reAddOnCreate      bool
	mergeAnnotations   bool
	deleteOnDisable    bool
}

type fileWatcher struct {
	cfg fileWatcherConfig
}

func (fw *fileWatcher) Name() string           { return fw.cfg.serviceName }
func (fw *fileWatcher) Dependencies() []string { return fw.cfg.dependencies }

func (fw *fileWatcher) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)

	if !fw.cfg.enabled {
		klog.Infof("%s is disabled (not configured)", fw.Name())
		fw.cleanupOnDisable(ctx)
		close(ready)
		return ctx.Err()
	}

	kubeClient, err := fw.createKubeClient()
	if err != nil {
		klog.Errorf("%s failed to create Kubernetes client: %v", fw.Name(), err)
		return err
	}

	if err := fw.updateConfigMap(ctx, kubeClient); err != nil {
		klog.Errorf("%s failed to create initial ConfigMap: %v", fw.Name(), err)
		return err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		klog.Errorf("%s failed to create file watcher: %v", fw.Name(), err)
		return err
	}
	defer func() {
		if cerr := watcher.Close(); cerr != nil {
			klog.Errorf("%s failed to close file watcher: %v", fw.Name(), cerr)
		}
	}()

	if err := fw.setupWatches(watcher); err != nil {
		return err
	}
	close(ready)

	lastHash, err := fw.getFileHash()
	if err != nil {
		klog.Warningf("%s failed to get initial file hash: %v", fw.Name(), err)
	}

	klog.Infof("%s is ready and watching %s", fw.Name(), fw.cfg.file)

	return fw.eventLoop(ctx, watcher, kubeClient, lastHash)
}

func (fw *fileWatcher) setupWatches(watcher *fsnotify.Watcher) error {
	filesToWatch := []string{
		fw.cfg.file,
		filepath.Dir(fw.cfg.file),
	}
	for i, file := range filesToWatch {
		if err := watcher.Add(file); err != nil {
			if i == 0 {
				klog.Errorf("%s failed to watch file %s: %v", fw.Name(), fw.cfg.file, err)
				return err
			}
			klog.Warningf("%s failed to watch directory %s: %v", fw.Name(), file, err)
		}
	}
	return nil
}

func (fw *fileWatcher) eventLoop(ctx context.Context, watcher *fsnotify.Watcher, kubeClient kubernetes.Interface, initHash string) error {
	lastHash := initHash

	for {
		select {
		case <-ctx.Done():
			klog.Infof("%s stopping", fw.Name())
			return ctx.Err()

		case event, ok := <-watcher.Events:
			if !ok {
				return fmt.Errorf("%s watcher channel closed", fw.Name())
			}
			if fw.isRelevantEvent(event) {
				if fw.cfg.reAddOnCreate && event.Op&fsnotify.Create == fsnotify.Create {
					_ = watcher.Add(fw.cfg.file)
				}
				updated, newHash, updateErr := fw.handleChange(ctx, kubeClient, lastHash)
				if updateErr != nil {
					klog.Errorf("%s failed to process file change: %v", fw.Name(), updateErr)
					continue
				}
				if updated {
					lastHash = newHash
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return fmt.Errorf("%s watcher error channel closed", fw.Name())
			}
			klog.Errorf("%s watcher error: %v", fw.Name(), err)
		}
	}
}

func (fw *fileWatcher) isRelevantEvent(event fsnotify.Event) bool {
	if event.Name != fw.cfg.file {
		return false
	}
	return event.Op&fw.cfg.eventMask != 0
}

func (fw *fileWatcher) handleChange(ctx context.Context, kubeClient kubernetes.Interface, lastHash string) (bool, string, error) {
	klog.Infof("%s detected change in %s", fw.Name(), fw.cfg.file)
	currentHash, err := fw.getFileHash()
	if err != nil {
		klog.Warningf("%s failed to get file hash after change: %v", fw.Name(), err)
		return false, lastHash, err
	}
	if currentHash == lastHash {
		klog.V(2).Infof("%s file hash unchanged, skipping update", fw.Name())
		return false, lastHash, nil
	}
	if err := fw.updateConfigMap(ctx, kubeClient); err != nil {
		klog.Errorf("%s failed to update ConfigMap: %v", fw.Name(), err)
		return false, lastHash, err
	}
	klog.Infof("%s successfully updated ConfigMap %s/%s", fw.Name(), fw.cfg.configMapNamespace, fw.cfg.configMapName)
	return true, currentHash, nil
}

func (fw *fileWatcher) createKubeClient() (*kubernetes.Clientset, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", fw.cfg.kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}
	return client, nil
}

func (fw *fileWatcher) updateConfigMap(ctx context.Context, client kubernetes.Interface) error {
	content, err := os.ReadFile(fw.cfg.file)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", fw.cfg.file, err)
	}
	return fw.createOrUpdateConfigMap(ctx, client, string(content))
}

func (fw *fileWatcher) createOrUpdateConfigMap(ctx context.Context, client kubernetes.Interface, content string) error {
	annotations := fw.buildAnnotations()
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fw.cfg.configMapName,
			Namespace:   fw.cfg.configMapNamespace,
			Labels:      fw.cfg.labels,
			Annotations: annotations,
		},
		Data: map[string]string{
			fw.cfg.configMapDataKey: content,
		},
	}

	configMapsClient := client.CoreV1().ConfigMaps(fw.cfg.configMapNamespace)

	existing, err := configMapsClient.Get(ctx, fw.cfg.configMapName, metav1.GetOptions{})
	switch {
	case apierrors.IsNotFound(err):
		if _, err = configMapsClient.Create(ctx, configMap, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("failed to create ConfigMap: %w", err)
		}
		klog.Infof("%s created ConfigMap %s/%s", fw.Name(), fw.cfg.configMapNamespace, fw.cfg.configMapName)
	case err != nil:
		return fmt.Errorf("failed to get ConfigMap: %w", err)
	default:
		existing.Data = configMap.Data
		if fw.cfg.mergeAnnotations {
			if existing.Annotations == nil {
				existing.Annotations = map[string]string{}
			}
			for k, v := range annotations {
				existing.Annotations[k] = v
			}
		} else {
			existing.Annotations = annotations
		}
		if _, err = configMapsClient.Update(ctx, existing, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("failed to update ConfigMap: %w", err)
		}
		klog.V(2).Infof("%s updated ConfigMap %s/%s", fw.Name(), fw.cfg.configMapNamespace, fw.cfg.configMapName)
	}

	return nil
}

func (fw *fileWatcher) buildAnnotations() map[string]string {
	annotations := make(map[string]string, len(fw.cfg.annotations)+1)
	for k, v := range fw.cfg.annotations {
		annotations[k] = v
	}
	annotations["microshift.io/last-updated"] = time.Now().Format(time.RFC3339)
	return annotations
}

func (fw *fileWatcher) cleanupOnDisable(ctx context.Context) {
	if !fw.cfg.deleteOnDisable {
		return
	}
	kubeClient, err := fw.createKubeClient()
	if err != nil {
		klog.Warningf("%s could not create Kubernetes client to delete ConfigMap: %v", fw.Name(), err)
		return
	}
	if err := fw.deleteConfigMap(ctx, kubeClient); err != nil {
		klog.Warningf("%s failed to delete ConfigMap when disabled: %v", fw.Name(), err)
	}
}

func (fw *fileWatcher) deleteConfigMap(ctx context.Context, client kubernetes.Interface) error {
	configMapsClient := client.CoreV1().ConfigMaps(fw.cfg.configMapNamespace)
	err := configMapsClient.Delete(ctx, fw.cfg.configMapName, metav1.DeleteOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			klog.V(2).Infof("%s ConfigMap %s/%s does not exist, nothing to delete", fw.Name(), fw.cfg.configMapNamespace, fw.cfg.configMapName)
			return nil
		}
		return fmt.Errorf("failed to delete ConfigMap: %w", err)
	}
	klog.Infof("%s deleted ConfigMap %s/%s", fw.Name(), fw.cfg.configMapNamespace, fw.cfg.configMapName)
	return nil
}

func (fw *fileWatcher) getFileHash() (string, error) {
	content, err := os.ReadFile(fw.cfg.file)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(content)
	return fmt.Sprintf("%x", hash), nil
}
