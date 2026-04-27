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

const (
	dnsConfigMapNamespace = "openshift-dns"
	dnsConfigMapName      = "dns-default"
	dnsConfigMapKey       = "Corefile"
)

type DNSConfigurationWatcherManager struct {
	file       string
	kubeconfig string
}

func NewDNSConfigurationWatcherManager(cfg *config.Config) *DNSConfigurationWatcherManager {
	return &DNSConfigurationWatcherManager{
		file:       cfg.DNS.ConfigFile,
		kubeconfig: cfg.KubeConfigPath(config.KubeAdmin),
	}
}

func (s *DNSConfigurationWatcherManager) Name() string           { return "dns-configuration-watcher-manager" }
func (s *DNSConfigurationWatcherManager) Dependencies() []string { return []string{"kube-apiserver"} }

func (s *DNSConfigurationWatcherManager) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)

	if s.file == "" {
		klog.Infof("%s is disabled (not configured)", s.Name())
		close(ready)
		return ctx.Err()
	}

	kubeClient, err := s.createKubeClient()
	if err != nil {
		klog.Errorf("%s failed to create Kubernetes client: %v", s.Name(), err)
		return err
	}

	if err := s.updateConfigMap(ctx, kubeClient); err != nil {
		klog.Errorf("%s failed to create initial ConfigMap: %v", s.Name(), err)
		return err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		klog.Errorf("%s failed to create file watcher: %v", s.Name(), err)
		return err
	}
	defer func() {
		if cerr := watcher.Close(); cerr != nil {
			klog.Errorf("%s failed to close file watcher: %v", s.Name(), cerr)
		}
	}()

	if err := s.setupWatches(watcher); err != nil {
		return err
	}
	close(ready)

	lastHash, err := s.getFileHash()
	if err != nil {
		klog.Warningf("%s failed to get initial file hash: %v", s.Name(), err)
	}

	klog.Infof("%s is ready and watching %s", s.Name(), s.file)

	return s.eventLoop(ctx, watcher, kubeClient, lastHash)
}

func (s *DNSConfigurationWatcherManager) setupWatches(watcher *fsnotify.Watcher) error {
	filesToWatch := []string{
		s.file,
		filepath.Dir(s.file),
	}
	for i, file := range filesToWatch {
		if err := watcher.Add(file); err != nil {
			if i == 0 {
				klog.Errorf("%s failed to watch DNS config file %s: %v", s.Name(), s.file, err)
				return err
			}
			klog.Warningf("%s failed to watch DNS config directory %s: %v", s.Name(), file, err)
		}
	}
	return nil
}

func (s *DNSConfigurationWatcherManager) eventLoop(ctx context.Context, watcher *fsnotify.Watcher, kubeClient kubernetes.Interface, initHash string) error {
	lastHash := initHash

	for {
		select {
		case <-ctx.Done():
			klog.Infof("%s stopping", s.Name())
			return watcher.Close()

		case event, ok := <-watcher.Events:
			if !ok {
				return fmt.Errorf("%s watcher channel closed", s.Name())
			}
			if s.isRelevantEvent(event) {
				updated, newHash, updateErr := s.handleChange(ctx, kubeClient, lastHash)
				if updateErr != nil {
					klog.Errorf("%s failed to process DNS config file change: %v", s.Name(), updateErr)
					continue
				}
				if updated {
					lastHash = newHash
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

func (s *DNSConfigurationWatcherManager) isRelevantEvent(event fsnotify.Event) bool {
	if event.Name != s.file {
		return false
	}
	return event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create
}

func (s *DNSConfigurationWatcherManager) handleChange(ctx context.Context, kubeClient kubernetes.Interface, lastHash string) (bool, string, error) {
	klog.Infof("%s detected change in DNS config file: %s", s.Name(), s.file)
	currentHash, err := s.getFileHash()
	if err != nil {
		klog.Warningf("%s failed to get file hash after change: %v", s.Name(), err)
		return false, lastHash, err
	}
	if currentHash == lastHash {
		klog.V(2).Infof("%s file hash unchanged, skipping update", s.Name())
		return false, lastHash, nil
	}
	if err := s.updateConfigMap(ctx, kubeClient); err != nil {
		klog.Errorf("%s failed to update ConfigMap: %v", s.Name(), err)
		return false, currentHash, err
	}
	klog.Infof("%s successfully updated ConfigMap %s/%s", s.Name(), dnsConfigMapNamespace, dnsConfigMapName)
	return true, currentHash, nil
}

func (s *DNSConfigurationWatcherManager) createKubeClient() (*kubernetes.Clientset, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", s.kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}
	return client, nil
}

func (s *DNSConfigurationWatcherManager) updateConfigMap(ctx context.Context, client kubernetes.Interface) error {
	content, err := os.ReadFile(s.file)
	if err != nil {
		return fmt.Errorf("failed to read DNS config file %s: %w", s.file, err)
	}

	return s.createOrUpdateConfigMap(ctx, client, string(content))
}

func (s *DNSConfigurationWatcherManager) createOrUpdateConfigMap(ctx context.Context, client kubernetes.Interface, corefileContent string) error {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dnsConfigMapName,
			Namespace: dnsConfigMapNamespace,
			Labels: map[string]string{
				"dns.operator.openshift.io/owning-dns": "default",
			},
			Annotations: map[string]string{
				"microshift.io/dns-config-file": s.file,
				"microshift.io/last-updated":    time.Now().Format(time.RFC3339),
			},
		},
		Data: map[string]string{
			dnsConfigMapKey: corefileContent,
		},
	}

	configMapsClient := client.CoreV1().ConfigMaps(dnsConfigMapNamespace)

	existing, err := configMapsClient.Get(ctx, dnsConfigMapName, metav1.GetOptions{})
	if err != nil {
		_, err = configMapsClient.Create(ctx, configMap, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create ConfigMap: %w", err)
		}
		klog.Infof("%s created ConfigMap %s/%s", s.Name(), dnsConfigMapNamespace, dnsConfigMapName)
	} else {
		existing.Data = configMap.Data
		existing.Annotations = configMap.Annotations
		_, err = configMapsClient.Update(ctx, existing, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update ConfigMap: %w", err)
		}
		klog.V(2).Infof("%s updated ConfigMap %s/%s", s.Name(), dnsConfigMapNamespace, dnsConfigMapName)
	}

	return nil
}

func (s *DNSConfigurationWatcherManager) getFileHash() (string, error) {
	content, err := os.ReadFile(s.file)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(content)
	return fmt.Sprintf("%x", hash), nil
}
