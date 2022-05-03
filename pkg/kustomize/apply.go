package kustomize

import (
	"context"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"time"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
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

type Kustomizer struct {
	paths      []string
	kubeconfig string
}

func NewKustomizer(cfg *config.MicroshiftConfig) *Kustomizer {
	return &Kustomizer{
		paths:      cfg.Manifests,
		kubeconfig: filepath.Join(cfg.DataDir, "resources", "kubeadmin", "kubeconfig"),
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
			klog.Fatalf("Applying kustomization failed: %s. Giving up.", err)
		} else {
			klog.Warningf("Kustomization applied successfully.")
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
	flags := cmds.PersistentFlags()
	flags.SetNormalizeFunc(cliflag.WarnWordSepNormalizeFunc)
	flags.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)

	kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	kubeConfigFlags.AddFlags(flags)
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)
	matchVersionKubeConfigFlags.AddFlags(cmds.PersistentFlags())
	cmds.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)
	ioStreams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	applyCommand := apply.NewCmdApply("kubectl", f, ioStreams)
	groups := templates.CommandGroups{
		{
			Message: "Advanced Commands:",
			Commands: []*cobra.Command{
				applyCommand,
			},
		},
	}
	groups.Add(cmds)

	args := []string{
		"--kubeconfig=" + kubeconfig,
		"-k", kustomization,
	}
	util.Must(applyCommand.ParseFlags(args))
	applyFlags := apply.NewApplyFlags(f, ioStreams)
	applyFlags.AddFlags(cmds)
	o, err := applyFlags.ToOptions(cmds, "kubectl", args)
	if err != nil {
		return err
	}

	if err := o.Validate(cmds, args); err != nil {
		return err
	}
	return o.Run()
}
