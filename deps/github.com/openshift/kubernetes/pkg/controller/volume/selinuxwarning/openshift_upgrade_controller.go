package selinuxwarning

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	applyconfigurationscorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/controller/volume/selinuxwarning/cache"
	"k8s.io/kubernetes/pkg/features"
)

const (
	checkInterval      = 30 * time.Second
	configMapNamespace = "openshift-config"
	configMapName      = "selinux-conflicts"
	fieldManager       = "selinux-conflicts-reporter"
)

type SELinuxConflictsReporterController struct {
	kubeClient        clientset.Interface
	conflictCounter   cache.ConflictCounter
	previousConflicts metav1.ConditionStatus
}

func NewSELinuxConflictsReporterController(kubeClient clientset.Interface, volumeCache cache.VolumeCache) *SELinuxConflictsReporterController {
	return &SELinuxConflictsReporterController{
		kubeClient: kubeClient,
		// Ugly retype to avoid more carry patches in Kubernetes code.
		// We added ConflictCounter in cache/openshift_patch.go,
		// therefore we know that VolumeCache implements it.
		conflictCounter:   volumeCache.(cache.ConflictCounter),
		previousConflicts: metav1.ConditionUnknown,
	}
}

func (c *SELinuxConflictsReporterController) Run(ctx context.Context) {
	logger := klog.FromContext(ctx)
	if !utilfeature.DefaultFeatureGate.Enabled(features.SELinuxMountGAReadiness) {
		logger.V(2).Info("SELinuxMountGAReadiness feature gate is disabled, not starting OpenShift SELinux conflicts reporter")
		return
	}
	logger.V(2).Info("Starting OpenShift SELinux conflicts reporter")
	timer := time.NewTimer(checkInterval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			c.reportSELinuxConflicts(ctx)
			timer.Reset(checkInterval)
		}
	}
}

func (c *SELinuxConflictsReporterController) reportSELinuxConflicts(ctx context.Context) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("Checking for SELinux conflicts")

	currentConflicts := c.getConflicts(logger)
	if currentConflicts == c.previousConflicts {
		logger.V(4).Info("SELinux conflict status did not change since last check")
		return
	}
	logger.V(4).Info("SELinux conflict status changed, updating the config map")
	if err := c.applySELinuxConflictsConfigMap(ctx, currentConflicts); err != nil {
		logger.Error(err, "Error saving conflicts config map")
		// To keep it simple: no exponential backoff try again in the next iteration.
		return
	}
	logger.V(2).Info("SELinux conflict updated", "Conflicts", currentConflicts)
	c.previousConflicts = currentConflicts
}

func (c *SELinuxConflictsReporterController) getConflicts(logger klog.Logger) metav1.ConditionStatus {
	conflictsCount := c.conflictCounter.GetConflictCount()
	if conflictsCount > 0 {
		logger.V(4).Info("Found SELinux-conflicting pods", "conflictsCount", conflictsCount)
		return metav1.ConditionTrue
	}
	logger.V(4).Info("Found no SELinux-conflicting pods")
	return metav1.ConditionFalse
}

func (c *SELinuxConflictsReporterController) applySELinuxConflictsConfigMap(ctx context.Context, conflictsPresent metav1.ConditionStatus) error {
	cm := applyconfigurationscorev1.ConfigMap(configMapName, configMapNamespace).
		WithData(map[string]string{
			"conflictsPresent": string(conflictsPresent),
		}).WithAnnotations(map[string]string{
		"Description": "This config map is used to report presence of SELinux conflicts from kube-controller-manager to storage Upgradeable condition in OpenShift 5.0",
	})
	_, err := c.kubeClient.CoreV1().ConfigMaps(configMapNamespace).Apply(ctx, cm, metav1.ApplyOptions{FieldManager: fieldManager, Force: true})
	if err != nil {
		return fmt.Errorf("error applying config map %s: %w", configMapName, err)
	}
	return nil
}
