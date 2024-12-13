package healthcheck

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/openshift/microshift/pkg/config"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/cmd/get"
	"k8s.io/kubectl/pkg/cmd/rollout"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/utils/ptr"
)

type NamespaceWorkloads struct {
	Deployments  []string `json:"deployments"`
	DaemonSets   []string `json:"daemonsets"`
	StatefulSets []string `json:"statefulsets"`
}

func waitForWorkloads(timeout time.Duration, workloads map[string]NamespaceWorkloads) error {
	aeg := &AllErrGroup{}
	for ns, wls := range workloads {
		aeg.Go(func() error { return waitForWorkloadsInNamespace(timeout, ns, wls) })
	}

	errs := aeg.Wait()
	if errs != nil {
		return errs
	}
	return nil
}

func getCoreMicroShiftWorkloads() (map[string]NamespaceWorkloads, error) {
	cfg, err := config.ActiveConfig()
	if err != nil {
		return nil, err
	}

	workloads := map[string]NamespaceWorkloads{
		"openshift-ovn-kubernetes": {
			DaemonSets: []string{"ovnkube-master", "ovnkube-node"},
		},
		"openshift-service-ca": {
			Deployments: []string{"service-ca"},
		},
		"openshift-ingress": {
			Deployments: []string{"router-default"},
		},
		"openshift-dns": {
			DaemonSets: []string{
				"dns-default",
				"node-resolver",
			},
		},
	}
	fillOptionalWorkloadsIfApplicable(cfg, workloads)

	return workloads, nil
}

func waitForCoreWorkloads(timeout time.Duration) error {
	workloads, err := getCoreMicroShiftWorkloads()
	if err != nil {
		return err
	}

	return waitForWorkloads(timeout, workloads)
}

func fillOptionalWorkloadsIfApplicable(cfg *config.Config, workloads map[string]NamespaceWorkloads) {
	klog.V(2).Infof("Configured storage driver value: %q", string(cfg.Storage.Driver))
	if cfg.Storage.IsEnabled() {
		klog.Infof("LVMS is enabled")
		workloads["openshift-storage"] = NamespaceWorkloads{
			DaemonSets:  []string{"vg-manager"},
			Deployments: []string{"lvms-operator"},
		}
	}
	if comps := getExpectedCSIComponents(cfg); len(comps) != 0 {
		klog.Infof("At least one CSI Component is enabled")
		workloads["kube-system"] = NamespaceWorkloads{
			Deployments: comps,
		}
	}
}

func getExpectedCSIComponents(cfg *config.Config) []string {
	klog.V(2).Infof("Configured optional CSI components: %v", cfg.Storage.OptionalCSIComponents)

	if len(cfg.Storage.OptionalCSIComponents) == 0 {
		return []string{"csi-snapshot-controller", "csi-snapshot-webhook"}
	}

	// Validation fails when there's more than one component provided and one of them is "None".
	// In other words: if "None" is used, it can be the only element.
	if len(cfg.Storage.OptionalCSIComponents) == 1 && cfg.Storage.OptionalCSIComponents[0] == config.CsiComponentNone {
		return nil
	}

	deployments := []string{}
	for _, comp := range cfg.Storage.OptionalCSIComponents {
		if comp == config.CsiComponentSnapshot {
			deployments = append(deployments, "csi-snapshot-controller")
		}
		if comp == config.CsiComponentSnapshotWebhook {
			deployments = append(deployments, "csi-snapshot-webhook")
		}
	}
	return deployments
}

// waitForReadyNamespace waits for ready workloads (daemonsets, deployments, and statefulsets)
// in a given namespace.
func waitForWorkloadsInNamespace(timeout time.Duration, ns string, workloads NamespaceWorkloads) error {
	cliOptions := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	cliOptions.KubeConfig = ptr.To(filepath.Join(config.DataDir, "resources", string(config.KubeAdmin), "kubeconfig"))
	cliOptions.Namespace = &ns
	if homedir.HomeDir() == "" {
		// By default client writes cache to $HOME/.kube/cache.
		// However, when healthcheck is executed by greenboot, the $HOME is empty,
		// so discovery client wants to write to /.kube which is immutable on ostre
		// causing flood of warnings (and is not elegant to create new root level directory).
		cliOptions.CacheDir = ptr.To(filepath.Join("tmp", ".kube", "cache"))
	}
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(cliOptions)
	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	stdout := strings.Builder{}
	stderr := strings.Builder{}
	ioStreams := genericclioptions.IOStreams{In: os.Stdin, Out: &stdout, ErrOut: &stderr}
	rolloutOpts := rollout.NewRolloutStatusOptions(ioStreams)
	rolloutOpts.Timeout = timeout

	args := []string{}
	for _, ds := range workloads.DaemonSets {
		args = append(args, fmt.Sprintf("daemonset/%s", ds))
	}
	for _, deploy := range workloads.Deployments {
		args = append(args, fmt.Sprintf("deployment/%s", deploy))
	}
	for _, statefulset := range workloads.StatefulSets {
		args = append(args, fmt.Sprintf("statefulset/%s", statefulset))
	}

	err := rolloutOpts.Complete(f, args)
	if err != nil {
		klog.Errorf("Failed to complete 'rollout' options for %q namespace: %v", ns, err)
		return err
	}

	err = rolloutOpts.Validate()
	if err != nil {
		klog.Errorf("Failed to validate 'rollout' options for %q namespace: %v", ns, err)
		return err
	}

	klog.Infof("Waiting for following workloads in %q namespace: %s", ns, strings.Join(args, " "))
	err = rolloutOpts.Run()
	klog.V(2).Infof("Rollout output for %q namespace: stdout='%s' stderr='%s'",
		ns,
		strings.ReplaceAll(strings.TrimSpace(stdout.String()), "\n", "; "),
		stderr.String())
	if err != nil {
		klog.Errorf("Failed waiting for readiness of the workloads in %q namespace: %v", ns, err)
		return err
	}
	klog.Infof("Workloads in %q namespace are ready", ns)

	return nil
}

func logPodsAndEvents() {
	cliOptions := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	cliOptions.KubeConfig = ptr.To(filepath.Join(config.DataDir, "resources", string(config.KubeAdmin), "kubeconfig"))
	if homedir.HomeDir() == "" {
		// By default client writes cache to $HOME/.kube/cache.
		// However, when healthcheck is executed by greenboot, the $HOME is empty,
		// so discovery client wants to write to /.kube which is immutable on ostre
		// causing flood of warnings (and is not elegant to create new root level directory).
		cliOptions.CacheDir = ptr.To(filepath.Join("tmp", ".kube", "cache"))
	}
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(cliOptions)
	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	output := strings.Builder{}
	ioStreams := genericclioptions.IOStreams{In: os.Stdin, Out: &output, ErrOut: &output}

	cmdGet := get.NewCmdGet("", f, ioStreams)
	opts := get.NewGetOptions("", ioStreams)
	opts.AllNamespaces = true
	opts.PrintFlags.OutputFormat = ptr.To("wide")
	if err := opts.Complete(f, cmdGet, []string{"DUMMY"}); err != nil {
		klog.Errorf("Failed to complete get's options: %v", err)
		return
	}

	if err := opts.Validate(); err != nil {
		klog.Errorf("Failed to validate get's options: %v", err)
		return
	}

	output.WriteString("---------- PODS:\n")
	if err := opts.Run(f, []string{"pods"}); err != nil {
		klog.Errorf("Failed to run 'get pods': %v", err)
		return
	}
	output.WriteString("\n---------- EVENTS:\n")
	opts.SortBy = ".metadata.creationTimestamp"
	if err := opts.Run(f, []string{"events"}); err != nil {
		klog.Errorf("Failed to run 'get events': %v", err)
		return
	}

	klog.Infof("DEBUG INFORMATION\n%s", output.String())
}
