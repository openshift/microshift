package healthcheck

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/openshift/microshift/pkg/config"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/cmd/get"
	"k8s.io/kubectl/pkg/cmd/rollout"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func waitForNamespaces(timeout time.Duration, namespaces []string) error {
	aeg := &AllErrGroup{}
	for _, ns := range namespaces {
		aeg.Go(func() error { return waitForReadyNamespace(timeout, ns) })
	}

	errs := aeg.Wait()
	if errs != nil {
		return errs
	}
	return nil
}

func getCoreMicroShiftNamespaces() ([]string, error) {
	cfg, err := config.ActiveConfig()
	if err != nil {
		return nil, err
	}
	namespaces := []string{"openshift-ovn-kubernetes", "openshift-service-ca", "openshift-ingress", "openshift-dns"}
	namespaces = append(namespaces, getOptionalNamespacesIfApplicable(cfg)...)
	return namespaces, nil
}

func waitForCoreWorkloads(timeout time.Duration) error {
	namespaces, err := getCoreMicroShiftNamespaces()
	if err != nil {
		return err
	}

	return waitForNamespaces(timeout, namespaces)
}

func getOptionalNamespacesIfApplicable(cfg *config.Config) []string {
	namespaces := []string{}

	klog.V(2).Infof("Configured storage driver value: %q", string(cfg.Storage.Driver))
	if cfg.Storage.IsEnabled() {
		klog.Infof("LVMS is enabled")
		namespaces = append(namespaces, "openshift-storage")
	}
	if csiComponentsAreExpected(cfg) {
		klog.Infof("At least one CSI Component is enabled")
		namespaces = append(namespaces, "kube-system")
	}
	return namespaces
}

func csiComponentsAreExpected(cfg *config.Config) bool {
	klog.V(2).Infof("Configured optional CSI components: %v", cfg.Storage.OptionalCSIComponents)

	if len(cfg.Storage.OptionalCSIComponents) == 0 {
		return true
	}

	// Validation fails when there's more than one component provided and one of them is "None".
	// In other words: if "None" is used, it can be the only element.
	if len(cfg.Storage.OptionalCSIComponents) == 1 && cfg.Storage.OptionalCSIComponents[0] == config.CsiComponentNone {
		return false
	}

	return true
}

// waitForReadyNamespace waits for ready workloads (daemonsets, deployments, and statefulsets)
// in a given namespace.
func waitForReadyNamespace(timeout time.Duration, ns string) error {
	kubeconfig := filepath.Join(config.DataDir, "resources", string(config.KubeAdmin), "kubeconfig")
	cliOptions := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	cliOptions.KubeConfig = &kubeconfig
	cliOptions.Namespace = &ns
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(cliOptions)
	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	stdout := strings.Builder{}
	stderr := strings.Builder{}
	ioStreams := genericclioptions.IOStreams{In: os.Stdin, Out: &stdout, ErrOut: &stderr}
	rolloutOpts := rollout.NewRolloutStatusOptions(ioStreams)
	rolloutOpts.Timeout = timeout
	err := rolloutOpts.Complete(f, []string{"daemonset,deployment,statefulset"})
	if err != nil {
		klog.Errorf("Failed to complete 'rollout' options for %q namespace: %v", ns, err)
		return err
	}

	err = rolloutOpts.Validate()
	if err != nil {
		klog.Errorf("Failed to validate 'rollout' options for %q namespace: %v", ns, err)
		return err
	}

	klog.Infof("Waiting for workloads in %q namespace", ns)
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
	kubeconfig := filepath.Join(config.DataDir, "resources", string(config.KubeAdmin), "kubeconfig")
	cliOptions := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	cliOptions.KubeConfig = &kubeconfig
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(cliOptions)
	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	output := strings.Builder{}
	ioStreams := genericclioptions.IOStreams{In: os.Stdin, Out: &output, ErrOut: &output}

	cmdGet := get.NewCmdGet("", f, ioStreams)
	if err := cmdGet.Flag("output").Value.Set("wide"); err != nil {
		klog.Errorf("Failed to set up --output=wide flag: %v", err)
	}
	if err := cmdGet.Flag("all-namespaces").Value.Set("true"); err != nil {
		klog.Errorf("Failed to set up --all-namespaces=true flag: %v", err)
	}

	output.WriteString("Pods:\n")
	cmdGet.Run(cmdGet, []string{"pods"})

	output.WriteString("\nEvents:\n")
	if err := cmdGet.Flag("sort-by").Value.Set(".metadata.creationTimestamp"); err != nil {
		klog.Errorf("Failed to set up --sort-by=.metadata.creationTimestamp flag: %v", err)
	}
	cmdGet.Run(cmdGet, []string{"events"})

	klog.Infof("Debug information:\n%s", output.String())
}

// AllErrGroup is a helper to wait for all goroutines and get all errors that occurred.
// It's based on sync.WaitGroup (which doesn't capture any errors) and errgroup.Group (which only captures the first error).
type AllErrGroup struct {
	wg   sync.WaitGroup
	mu   sync.Mutex
	errs []error

	amount int
}

func (g *AllErrGroup) Go(f func() error) {
	g.wg.Add(1)
	g.amount += 1
	go func() {
		defer g.wg.Done()
		if err := f(); err != nil {
			g.mu.Lock()
			defer g.mu.Unlock()
			g.errs = append(g.errs, err)
		}
	}()
}

func (g *AllErrGroup) Wait() error {
	klog.V(2).Infof("Waiting for %d goroutines", g.amount)
	g.wg.Wait()
	return errors.Join(g.errs...)
}
