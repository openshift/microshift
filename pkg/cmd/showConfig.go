package cmd

import (
	"fmt"

	"github.com/openshift/microshift/pkg/config"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/yaml"
)

type showConfigOptions struct {
	Mode string
	genericclioptions.IOStreams
}

func NewShowConfigCommand(ioStreams genericclioptions.IOStreams) *cobra.Command {
	opts := showConfigOptions{
		Mode: "default",
	}

	cmd := &cobra.Command{
		Use:   "show-config",
		Short: "Print MicroShift's configuration",
		Run: func(cmd *cobra.Command, args []string) {
			var cfg *config.Config
			var err error

			if opts.Mode == "effective" {
				cfg, err = config.GetActiveConfig()
				if err != nil {
					cmdutil.CheckErr(err)
				}
			} else {
				cfg = config.NewMicroshiftConfig()
			}

			marshalled, err := yaml.Marshal(cfg)
			cmdutil.CheckErr(err)

			fmt.Fprintf(ioStreams.Out, "%s\n", string(marshalled))
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Mode, "mode", "m", opts.Mode, "One of 'default' or 'effective'.")

	return cmd
}
