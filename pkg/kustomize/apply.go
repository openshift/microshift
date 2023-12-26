package kustomize

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/openshift/microshift/pkg/config"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/cmd/apply"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/templates"
	"sigs.k8s.io/kustomize/api/konfig"
)

const (
	retryInterval = 10 * time.Second
	retryTimeout  = 1 * time.Minute
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

func (s *Kustomizer) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)
	defer close(ready)

	kustomizationPaths, err := s.cfg.Manifests.GetKustomizationPaths()
	if err != nil {
		return fmt.Errorf("failed to find any kustomization paths: %w", err)
	}

	for _, path := range kustomizationPaths {
		s.applyKustomizationPath(ctx, path)
	}

	return ctx.Err()
}

func (s *Kustomizer) applyKustomizationPath(ctx context.Context, path string) {
	kustomizationFileNames := konfig.RecognizedKustomizationFileNames()

	for _, filename := range kustomizationFileNames {
		kustomization := filepath.Join(path, filename)

		if _, err := os.Stat(kustomization); errors.Is(err, os.ErrNotExist) {
			klog.Infof("No kustomization found at " + kustomization)
			continue
		}

		klog.Infof("Applying kustomization at %v ", kustomization)
		if err := applyKustomizationWithRetries(ctx, path, s.kubeconfig); err != nil {
			klog.Errorf("Applying kustomization at %v failed: %v. Giving up.", kustomization, err)
		} else {
			klog.Infof("Kustomization at %v applied successfully.", kustomization)
		}
	}
}

func applyKustomizationWithRetries(ctx context.Context, kustomization string, kubeconfig string) error {
	return wait.PollUntilContextTimeout(ctx, retryInterval, retryTimeout, true, func(_ context.Context) (done bool, err error) {
		if err := applyKustomization(kustomization, kubeconfig); err != nil {
			klog.Infof("Applying kustomization failed: %s. Retrying in %s.", err, retryInterval)
			return false, nil
		}
		return true, nil
	})
}

func applyKustomization(kustomization string, kubeconfig string) error {
	klog.Infof("Applying kustomization at %s", kustomization)

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

	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)
	ioStreams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	groups := templates.CommandGroups{
		{
			Message: "Advanced Commands:",
			Commands: []*cobra.Command{
				apply.NewCmdApply("kubectl", f, ioStreams),
			},
		},
	}
	groups.Add(cmds)

	applyFlags := apply.NewApplyFlags(ioStreams)
	applyFlags.DeleteFlags.FileNameFlags.Kustomize = &kustomization
	applyFlags.AddFlags(cmds)

	o, err := applyFlags.ToOptions(f, cmds, "kubectl", nil)
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
