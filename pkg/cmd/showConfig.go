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
				cfg.Node.NodeIP = ""
				cfg.Node.HostnameOverride = ""
			case "effective":
				// Load the current configuration
				if err := cfg.ReadAndValidate(config.GetConfigFile()); err != nil {
					cmdutil.CheckErr(err)
				}
			default:
				cmdutil.CheckErr(fmt.Errorf("Unknown mode %q", opts.Mode))
			}

			// map back from internal representation to user config
			userCfg := config.Config{
				Network: config.Network{
					ClusterNetwork: []config.ClusterNetworkEntry{
						{CIDR: cfg.Cluster.ClusterCIDR},
					},
					ServiceNetwork:       []string{cfg.Cluster.ServiceCIDR},
					ServiceNodePortRange: cfg.Cluster.ServiceNodePortRange,
				},
				DNS:  cfg.DNS,
				Node: cfg.Node,
				ApiServer: config.ApiServer{
					SubjectAltNames: cfg.SubjectAltNames,
				},
				Debugging: cfg.Debugging,
			}
			marshalled, err := yaml.Marshal(userCfg)
			cmdutil.CheckErr(err)

			fmt.Fprintf(ioStreams.Out, "%s\n", string(marshalled))
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Mode, "mode", "m", opts.Mode, "One of 'default' or 'effective'.")

	return cmd
}
