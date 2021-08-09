package kustomize

import (
	"context"
	"errors"
	"flag"
	"os"
	"path/filepath"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/kubectl/pkg/cmd/apply"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
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
		logrus.Infof("Applying kustomization at " + kustomization)
		ApplyKustomization(s.path, s.kubeconfig)
	} else {
		logrus.Infof("No kustomization found at " + kustomization)
	}

	return ctx.Err()
}

func ApplyKustomization(path string, kubeconfig string) error {
	cmds := &cobra.Command{
		Use:   "kubectl",
		Short: i18n.T("kubectl controls the Kubernetes cluster manager"),
	}
	flags := cmds.PersistentFlags()
	flags.SetNormalizeFunc(cliflag.WarnWordSepNormalizeFunc) // Warn for "_" flags
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
		"-k", path,
	}

	util.Must(applyCommand.ParseFlags(args))
	applyCommand.Run(applyCommand, nil)

	return nil
}
