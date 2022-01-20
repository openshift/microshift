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
	path       string
	kubeconfig string
}

func NewKustomizer(cfg *config.MicroshiftConfig) *Kustomizer {
	return &Kustomizer{
		path:       filepath.Join(cfg.DataDir, "manifests"),
		kubeconfig: filepath.Join(cfg.DataDir, "resources", "kubeadmin", "kubeconfig"),
	}
}

func (s *Kustomizer) Name() string           { return "kustomizer" }
func (s *Kustomizer) Dependencies() []string { return []string{"kube-apiserver"} }

func (s *Kustomizer) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)
	defer close(ready)

	kustomization := filepath.Join(s.path, "kustomization.yaml")
	if _, err := os.Stat(kustomization); !errors.Is(err, os.ErrNotExist) {
		klog.Infof("Applying kustomization at %v ", kustomization)
		if err := ApplyKustomizationWithRetries(s.path, s.kubeconfig); err != nil {
			klog.Warningf("Applying kustomization failed: %s. Giving up.", err)
		} else {
			klog.Warningf("Kustomization applied successfully.")
		}
	} else {
		klog.Infof("No kustomization found at " + kustomization)
	}

	return ctx.Err()
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

	o := apply.NewApplyOptions(ioStreams)
	o.DeleteFlags.FileNameFlags.Kustomize = &kustomization
	if err := o.Complete(f, applyCommand); err != nil {
		return err
	}
	return o.Run()
}
