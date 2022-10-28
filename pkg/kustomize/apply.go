package kustomize

import (
	"context"
	"errors"
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
)

const (
	retryInterval = 10 * time.Second
	retryTimeout  = 1 * time.Minute
)

var microshiftManifestsDir = config.GetManifestsDir()

type Kustomizer struct {
	paths      []string
	kubeconfig string
}

func NewKustomizer(cfg *config.MicroshiftConfig) *Kustomizer {
	return &Kustomizer{
		paths:      microshiftManifestsDir,
		kubeconfig: cfg.KubeConfigPath(config.KubeAdmin),
	}
}

func (s *Kustomizer) Name() string           { return "kustomizer" }
func (s *Kustomizer) Dependencies() []string { return []string{"kube-apiserver"} }

func (s *Kustomizer) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)
	defer close(ready)

	for _, path := range s.paths {
		s.ApplyKustomizationPath(path)
	}

	return ctx.Err()
}

func (s *Kustomizer) ApplyKustomizationPath(path string) {
	kustomization := filepath.Join(path, "kustomization.yaml")
	if _, err := os.Stat(kustomization); !errors.Is(err, os.ErrNotExist) {
		klog.Infof("Applying kustomization at %v ", kustomization)
		if err := ApplyKustomizationWithRetries(path, s.kubeconfig); err != nil {
			klog.Fatalf("Applying kustomization at %v failed: %s. Giving up.", kustomization, err)
		} else {
			klog.Infof("Kustomization at %v applied successfully.", kustomization)
		}
	} else {
		klog.Infof("No kustomization found at " + kustomization)
	}
}

func ApplyKustomizationWithRetries(kustomization string, kubeconfig string) error {
	return wait.Poll(retryInterval, retryTimeout, func() (bool, error) {
		if err := ApplyKustomization(kustomization, kubeconfig); err != nil {
			klog.Infof("Applying kustomization failed: %s. Retrying in %s.", err, retryInterval)
			return false, nil
		}
		return true, nil
	})
}

func ApplyKustomization(kustomization string, kubeconfig string) error {
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

	applyFlags := apply.NewApplyFlags(f, ioStreams)
	applyFlags.DeleteFlags.FileNameFlags.Kustomize = &kustomization
	applyFlags.AddFlags(cmds)

	o, err := applyFlags.ToOptions(cmds, "kubectl", nil)
	if err != nil {
		return err
	}

	if err := o.Validate(); err != nil {
		return err
	}
	return o.Run()
}
