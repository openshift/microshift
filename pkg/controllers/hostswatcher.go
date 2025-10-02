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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

var (
	targetNameSpace = "openshift-dns"
	configMapName   = "hosts-file"
)

type HostsWatcherManager struct {
	name         string
	dependencies []string
	cfg          *config.Config
}

func NewHostsWatcherManager(cfg *config.Config) *HostsWatcherManager {
	return &HostsWatcherManager{
		name:         "hosts-watcher-manager",
		dependencies: []string{"kube-apiserver"},
		cfg:          cfg,
	}
}

func (s *HostsWatcherManager) Name() string           { return s.name }
func (s *HostsWatcherManager) Dependencies() []string { return s.dependencies }

func (s *HostsWatcherManager) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)

	if s.cfg.DNS.Hosts.Status != config.HostsStatusEnabled {
		klog.Infof("%s is disabled (not configured)", s.Name())
		defer close(ready)
		return ctx.Err()
	}

	klog.Infof("%s starting to monitor hosts file: %s", s.Name(), s.cfg.DNS.Hosts.File)
	close(ready)

	// Create Kubernetes client
	kubeClient, err := s.createKubeClient()
	if err != nil {
		klog.Errorf("%s failed to create Kubernetes client: %v", s.Name(), err)
		return err
	}

	// Create initial ConfigMaps
	if err := s.updateConfigMaps(ctx, kubeClient); err != nil {
		klog.Errorf("%s failed to create initial ConfigMaps: %v", s.Name(), err)
		return err
	}

	// Set up file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		klog.Errorf("%s failed to create file watcher: %v", s.Name(), err)
		return err
	}
	defer watcher.Close()

	// Add the hosts file to the watcher
	err = watcher.Add(s.cfg.DNS.Hosts.File)
	if err != nil {
		klog.Errorf("%s failed to watch hosts file %s: %v", s.Name(), s.cfg.DNS.Hosts.File, err)
		return err
	}

	// Also watch the directory to catch file replacements (common with atomic writes)
	hostsDir := filepath.Dir(s.cfg.DNS.Hosts.File)
	err = watcher.Add(hostsDir)
	if err != nil {
		klog.Warningf("%s failed to watch hosts directory %s: %v", s.Name(), hostsDir, err)
	}

	klog.Infof("%s ready and watching for changes", s.Name())

	// Track last known content hash to avoid duplicate updates
	lastHash, err := s.getFileHash(s.cfg.DNS.Hosts.File)
	if err != nil {
		klog.Warningf("%s failed to get initial file hash: %v", s.Name(), err)
	}

	klog.Infof("%s is ready", s.Name())

	for {
		select {
		case <-ctx.Done():
			klog.Infof("%s stopping", s.Name())
			return nil

		case event, ok := <-watcher.Events:
			if !ok {
				return fmt.Errorf("%s watcher channel closed", s.Name())
			}

			// Check if this event is for our hosts file
			if event.Name != s.cfg.DNS.Hosts.File {
				continue
			}

			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				klog.V(2).Infof("%s detected change in hosts file: %s", s.Name(), event.Name)

				// Check if content actually changed
				currentHash, err := s.getFileHash(s.cfg.DNS.Hosts.File)
				if err != nil {
					klog.Warningf("%s failed to get file hash after change: %v", s.Name(), err)
					continue
				}

				if currentHash == lastHash {
					klog.V(2).Infof("%s file hash unchanged, skipping update", s.Name())
					continue
				}

				lastHash = currentHash

				// Update ConfigMaps in all target namespaces
				if err := s.updateConfigMaps(ctx, kubeClient); err != nil {
					klog.Errorf("%s failed to update ConfigMaps: %v", s.Name(), err)
				} else {
					klog.Infof("%s successfully updated ConfigMaps in namespaces: %v", s.Name(), targetNameSpace)
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return fmt.Errorf("%s watcher error channel closed", s.Name())
			}
			klog.Errorf("%s watcher error: %v", s.Name(), err)
		}
	}
}

func (s *HostsWatcherManager) createKubeClient() (kubernetes.Interface, error) {
	kubeConfigPath := s.cfg.KubeConfigPath(config.KubeAdmin)

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return client, nil
}

func (s *HostsWatcherManager) updateConfigMaps(ctx context.Context, client kubernetes.Interface) error {
	hostsContent, err := s.readHostsFile()
	if err != nil {
		return fmt.Errorf("failed to read hosts file: %w", err)
	}

	if err := s.createOrUpdateConfigMap(ctx, client, targetNameSpace, hostsContent); err != nil {
		klog.Errorf("%s failed to update ConfigMap in namespace %s: %v", s.Name(), targetNameSpace, err)
	}

	return nil
}

func (s *HostsWatcherManager) createOrUpdateConfigMap(ctx context.Context, client kubernetes.Interface, namespace string, hostsContent string) error {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "microshift-hosts-watcher",
				"app.kubernetes.io/component":  "hosts-file-sync",
				"app.kubernetes.io/managed-by": "microshift",
			},
			Annotations: map[string]string{
				"microshift.io/hosts-file-path": s.cfg.DNS.Hosts.File,
				"microshift.io/last-updated":    time.Now().Format(time.RFC3339),
			},
		},
		Data: map[string]string{
			"hosts": hostsContent,
		},
	}

	configMapsClient := client.CoreV1().ConfigMaps(namespace)

	// Try to get existing ConfigMap
	existing, err := configMapsClient.Get(ctx, configMapName, metav1.GetOptions{})
	if err != nil {
		// ConfigMap doesn't exist, create it
		_, err = configMapsClient.Create(ctx, configMap, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create ConfigMap: %w", err)
		}
		klog.Infof("%s created ConfigMap %s in namespace %s", s.Name(), configMapName, namespace)
	} else {
		// ConfigMap exists, update it
		existing.Data = configMap.Data
		existing.Annotations = configMap.Annotations
		_, err = configMapsClient.Update(ctx, existing, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update ConfigMap: %w", err)
		}
		klog.V(2).Infof("%s updated ConfigMap %s in namespace %s", s.Name(), configMapName, namespace)
	}

	return nil
}

func (s *HostsWatcherManager) readHostsFile() (string, error) {
	content, err := os.ReadFile(s.cfg.DNS.Hosts.File)
	if err != nil {
		return "", fmt.Errorf("failed to read hosts file %s: %w", s.cfg.DNS.Hosts.File, err)
	}
	return string(content), nil
}

func (s *HostsWatcherManager) getFileHash(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(content)
	return fmt.Sprintf("%x", hash), nil
}
