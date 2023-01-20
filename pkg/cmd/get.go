package cmd

import (
	"os/user"

	"github.com/openshift/microshift/pkg/config"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/cmd/get"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func NewGetCommand(ioStreams genericclioptions.IOStreams) *cobra.Command {

	kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()

	// Initialize the kubeconfig default path to the location we
	// expect the server to have generated the file, but only if the
	// user is root. That leaves the default lookup logic from kubectl
	// in place for other users.
	userInfo, err := user.Current()
	if err != nil {
		klog.Fatalf("Failed to detect the current user %v", err)
	}
	if userInfo.Uid == "0" {
		cfg := config.NewMicroshiftConfig()
		kubeconfig := cfg.KubeConfigPath(config.KubeAdmin)
		kubeConfigFlags.KubeConfig = &kubeconfig
	}

	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)
	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	cmd := get.NewCmdGet("kubectl", f, ioStreams)
	persistFlags := cmd.PersistentFlags()
	persistFlags.SetNormalizeFunc(cliflag.WarnWordSepNormalizeFunc)
	persistFlags.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)
	kubeConfigFlags.AddFlags(persistFlags)

	return cmd
}
