package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/yaml"

	"github.com/openshift/microshift/pkg/config"
)

type showConfigOptions struct {
	Mode string
	genericclioptions.IOStreams
}

func NewShowConfigCommand(ioStreams genericclioptions.IOStreams) *cobra.Command {
	opts := showConfigOptions{
		Mode: "default",
	}

	cfg := config.NewMicroshiftConfig()

	cmd := &cobra.Command{
		Use:   "show-config",
		Short: "Print MicroShift's configuration",
		Run: func(cmd *cobra.Command, args []string) {

			switch opts.Mode {
			case "default":
				cfg.NodeIP = ""
				cfg.NodeName = ""
			case "effective":
				// Load the current configuration
				if err := cfg.ReadAndValidate(config.GetConfigFile(), cmd.Flags()); err != nil {
					cmdutil.CheckErr(err)
				}
			default:
				cmdutil.CheckErr(fmt.Errorf("Unknown mode %q", opts.Mode))
			}

			marshalled, err := yaml.Marshal(cfg)
			cmdutil.CheckErr(err)

			fmt.Fprintf(ioStreams.Out, "%s\n", string(marshalled))
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Mode, "mode", "m", opts.Mode, "One of 'default' or 'effective'.")
	addRunFlags(cmd, cfg)

	return cmd
}
