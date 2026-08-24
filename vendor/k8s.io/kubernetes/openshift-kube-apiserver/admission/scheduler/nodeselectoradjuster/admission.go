package nodeselectoradjuster

// The NodeSelectorAdjuster admission plugin removes the
// node-role.kubernetes.io/master node selector from qualifying pods. It only
// activates on HCP (hosted control plane) clusters, detected by
// POD_NAMESPACE != openshift-kube-apiserver. On standalone OpenShift clusters
// the plugin does not register itself and takes no action, leaving the master
// node selector in place so that qualifying pods run on master nodes as
// intended.

import (
	"context"
	"fmt"
	"io"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiserver/pkg/admission"
	coreapi "k8s.io/kubernetes/pkg/apis/core"
)

const (
	// PluginName is the name used to identify this plugin in the admission chain.
	PluginName = "scheduling.openshift.io/NodeSelectorAdjuster"

	// masterRoleKey is the node role label used as a node selector key in
	// 4.x VPA operator manifests.
	masterRoleKey = "node-role.kubernetes.io/master"

	// vpaOperatorLabelKey / vpaOperatorLabelValue identify the VPA operator pod.
	vpaOperatorLabelKey   = "k8s-app"
	vpaOperatorLabelValue = "vertical-pod-autoscaler-operator"
	// vpaOperatorNamespace is the namespace the VPA operator is expected to run in.
	vpaOperatorNamespace = "openshift-vertical-pod-autoscaler"

	// standaloneEnvVar is the environment variable checked at start-up.
	// It is injected by the downward API and reflects the namespace the
	// kube-apiserver pod runs in.
	standaloneEnvVar = "POD_NAMESPACE"
	// standaloneEnvValue is the namespace used by the kube-apiserver on a
	// standalone OpenShift cluster.
	standaloneEnvValue = "openshift-kube-apiserver"
)

// IsStandalone reports whether the current process is running inside a standalone
// OpenShift cluster. It is checked once at start-up to decide whether the plugin
// should register itself.
func IsStandalone() bool {
	return os.Getenv(standaloneEnvVar) == standaloneEnvValue
}

// Register adds the plugin to the admission plugin registry. It must only be
// called when IsStandalone() returns false (i.e. on HCP clusters).
func Register(plugins *admission.Plugins) {
	plugins.Register(PluginName, func(_ io.Reader) (admission.Interface, error) {
		return &nodeSelectorAdjuster{
			Handler: admission.NewHandler(admission.Create),
		}, nil
	})
}

// nodeSelectorAdjuster implements admission.MutationInterface.
type nodeSelectorAdjuster struct {
	*admission.Handler
}

var _ admission.MutationInterface = &nodeSelectorAdjuster{}

// Admit examines newly-created Pod objects and, for qualifying pods, removes the
// master node selector so that they can run on worker nodes in HCP clusters.
func (p *nodeSelectorAdjuster) Admit(_ context.Context, attr admission.Attributes, _ admission.ObjectInterfaces) error {
	if attr.GetResource().GroupResource() != corev1.Resource("pods") || attr.GetSubresource() != "" {
		return nil
	}

	pod, ok := attr.GetObject().(*coreapi.Pod)
	if !ok {
		return admission.NewForbidden(attr, fmt.Errorf("unexpected object type: %T", attr.GetObject()))
	}

	if !requiresNodeSelectorAdjustment(pod) {
		return nil
	}

	delete(pod.Spec.NodeSelector, masterRoleKey)
	return nil
}

// ValidateInitialization satisfies admission.InitializationValidator. The plugin
// has no external dependencies to validate.
func (p *nodeSelectorAdjuster) ValidateInitialization() error {
	return nil
}

// requiresNodeSelectorAdjustment returns true when the pod carries a label that
// opts it in to node placement and lives in a namespace where that
// label is expected. Control-plane-adjacent Day 2 operators can be added here.
func requiresNodeSelectorAdjustment(pod *coreapi.Pod) bool {
	// For VPA, we only want to update if the sole node selector is the
	// master role key, which is the default from the 4.x VPA operator manifests.
	if pod.Labels[vpaOperatorLabelKey] == vpaOperatorLabelValue &&
		pod.Namespace == vpaOperatorNamespace && len(pod.Spec.NodeSelector) == 1 {
		if _, found := pod.Spec.NodeSelector[masterRoleKey]; found {
			return true
		}
	}
	return false
}
