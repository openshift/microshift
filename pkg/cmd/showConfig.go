package cmd

import (
	"fmt"
	"os"

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
		Mode: "effective",
	}

	cmd := &cobra.Command{
		Use:   "show-config",
		Short: "Print MicroShift's configuration",
		Run: func(cmd *cobra.Command, args []string) {
			var cfg *config.Config
			var err error

			if os.Geteuid() > 0 {
				cmdutil.CheckErr(fmt.Errorf("command requires root privileges"))
			}

			switch opts.Mode {
			case "effective":
				cfg, err = config.ActiveConfig()
				if err != nil {
					cmdutil.CheckErr(err)
				}
				err = cfg.EnsureNodeNameHasNotChanged()
				if err != nil {
					cmdutil.CheckErr(err)
				}
			case "default":
				cfg = config.NewDefault()
			default:
				cmdutil.CheckErr(fmt.Errorf("unrecognized mode %q", opts.Mode))
			}

			marshalled, err := yaml.Marshal(cfg)
			cmdutil.CheckErr(err)

			fmt.Fprintf(ioStreams.Out, "%s\n", string(marshalled))

			for _, w := range cfg.Warnings {
				fmt.Fprintf(ioStreams.Out, "# WARNING: %s\n", w)
			}
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Mode, "mode", "m", opts.Mode, "One of 'default' or 'effective'.")

	return cmd
}
