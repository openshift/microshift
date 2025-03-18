package kustomize

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/openshift/microshift/pkg/config"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/cmd/apply"
	"k8s.io/kubectl/pkg/cmd/delete"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/templates"
)

const (
	retryInterval = 10 * time.Second
	retryTimeout  = 10 * time.Minute
)

type Kustomizer struct {
	cfg        *config.Config
	kubeconfig string
}

func NewKustomizer(cfg *config.Config) *Kustomizer {
	return &Kustomizer{
		cfg:        cfg,
		kubeconfig: cfg.KubeConfigPath(config.KubeAdmin),
	}
}

func (s *Kustomizer) Name() string           { return "kustomizer" }
func (s *Kustomizer) Dependencies() []string { return []string{"kube-apiserver"} }

func (s *Kustomizer) RunStandalone(ctx context.Context) {
	ready, stopped := make(chan struct{}), make(chan struct{})
	go func() {
		if err := s.Run(ctx, ready, stopped); err != nil && !errors.Is(err, context.Canceled) {
			klog.Errorf("Kustomizer failed: %v", err)
		}
	}()
}

func (s *Kustomizer) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)
	defer close(ready)

	kustomizationPaths, err := s.cfg.Manifests.GetKustomizationPaths()
	if err != nil {
		return fmt.Errorf("failed to find any kustomization paths: %w", err)
	}
	deletePaths, err := s.cfg.Manifests.GetKustomizationDeletePaths()
	if err != nil {
		return fmt.Errorf("failed to find any delete kustomization paths: %w", err)
	}

	for _, path := range deletePaths {
		s.handleKustomizationPath(ctx, path, "Deleting", deleteKustomization)
	}

	for _, path := range kustomizationPaths {
		s.handleKustomizationPath(ctx, path, "Applying", applyKustomization)
	}

	return ctx.Err()
}

func (s *Kustomizer) handleKustomizationPath(ctx context.Context, path string, verb string, actionFunc func(string, string) error) {
	klog.Infof("%s kustomization at %v ", verb, path)
	err := wait.PollUntilContextTimeout(ctx, retryInterval, retryTimeout, true, func(_ context.Context) (done bool, err error) {
		if err := actionFunc(path, s.kubeconfig); err != nil {
			klog.Infof("%s kustomization failed: %s. Retrying in %s.", verb, err, retryInterval)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		klog.Errorf("%s kustomization at %v failed: %v. Giving up.", verb, path, err)
	} else {
		klog.Infof("%s kustomization at %v was successful.", verb, path)
	}
}

func applyKustomization(kustomization string, kubeconfig string) error {
	kubectlCmd := getKubectlCmd(kustomization, kubeconfig)
	applyFlags := apply.NewApplyFlags(kubectlCmd.ioStreams)
	applyFlags.DeleteFlags.FileNameFlags.Kustomize = &kustomization
	applyFlags.AddFlags(kubectlCmd.cmds)

	o, err := applyFlags.ToOptions(kubectlCmd.factory, kubectlCmd.cmds, "kubectl", nil)
	if err != nil {
		return err
	}

	// Enable server-side apply to ensure big resources are applied successfully.
	o.ServerSideApply = true
	// Force conflicts to ensure that resources are applied even if they have changed on kube.
	o.ForceConflicts = true

	if err := o.Validate(); err != nil {
		return err
	}
	return o.Run()
}

func deleteKustomization(kustomization string, kubeconfig string) error {
	kubectlCmd := getKubectlCmd(kustomization, kubeconfig)
	cmdutil.AddDryRunFlag(kubectlCmd.cmds)

	deleteFlags := delete.NewDeleteFlags("")
	deleteFlags.FileNameFlags.Kustomize = &kustomization
	ignoreNotFound := true
	deleteFlags.IgnoreNotFound = &ignoreNotFound
	deleteFlags.AddFlags(kubectlCmd.cmds)

	o, err := deleteFlags.ToOptions(nil, kubectlCmd.ioStreams)
	if err != nil {
		return err
	}
	warningLogger := logWritter{f: klog.Warningf, prelude: fmt.Sprintf("Kustomization warning (%q)", kustomization)}
	o.WarningPrinter = printers.NewWarningPrinter(warningLogger, printers.WarningPrinterOptions{})

	if err := o.Complete(kubectlCmd.factory, []string{}, kubectlCmd.cmds); err != nil {
		return err
	}

	if err := o.Validate(); err != nil {
		return err
	}
	return o.RunDelete(kubectlCmd.factory)
}

type kubectlCmd struct {
	cmds      *cobra.Command
	factory   cmdutil.Factory
	ioStreams genericclioptions.IOStreams
}

func getKubectlCmd(kustomization, kubeconfig string) kubectlCmd {
	cmds := &cobra.Command{
		Use:   "kubectl",
		Short: "kubectl",
	}
	persistFlags := cmds.PersistentFlags()
	persistFlags.SetNormalizeFunc(cliflag.WarnWordSepNormalizeFunc)
	persistFlags.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)

	kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	kubeConfigFlags.KubeConfig = &kubeconfig
	kubeConfigFlags.AddFlags(persistFlags)

	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)
	matchVersionKubeConfigFlags.AddFlags(persistFlags)

	stdoutLogger := logWritter{f: klog.Infof, prelude: fmt.Sprintf("Kustomization (%q)", kustomization)}
	stderrLogger := logWritter{f: klog.Errorf, prelude: fmt.Sprintf("Kustomization error (%q)", kustomization)}
	ioStreams := genericclioptions.IOStreams{In: os.Stdin, Out: stdoutLogger, ErrOut: stderrLogger}

	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)
	groups := templates.CommandGroups{
		{
			Message: "Basic Commands (Intermediate):",
			Commands: []*cobra.Command{
				delete.NewCmdDelete(f, ioStreams),
			},
		},
		{
			Message: "Advanced Commands:",
			Commands: []*cobra.Command{
				apply.NewCmdApply("kubectl", f, ioStreams),
			},
		},
	}
	groups.Add(cmds)

	return kubectlCmd{cmds: cmds, factory: f, ioStreams: ioStreams}
}

type logWritter struct {
	f       func(format string, args ...interface{})
	prelude string
}

func (lw logWritter) Write(p []byte) (n int, err error) {
	lw.f("%s: %s", lw.prelude, string(p))
	return len(p), nil
}
