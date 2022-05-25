package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/openshift/microshift/pkg/release"
	"github.com/openshift/microshift/pkg/version"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

type VersionOptions struct {
	Output string

	genericclioptions.IOStreams
}

func NewVersionOptions(ioStreams genericclioptions.IOStreams) *VersionOptions {
	return &VersionOptions{
		IOStreams: ioStreams,
	}
}

func NewVersionCommand(ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := NewVersionOptions(ioStreams)
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print MicroShift version information",
		Run: func(cmd *cobra.Command, args []string) {
			// cmdutil.CheckErr(o.Validate())
			// cmdutil.CheckErr(o.Complete(f, cmd))
			cmdutil.CheckErr(o.Run())
		},
	}

	cmd.Flags().StringVarP(&o.Output, "output", "o", o.Output, "One of 'yaml' or 'json'.")

	return cmd
}

func (o *VersionOptions) Run() error {
	versionInfo := version.Get()

	switch o.Output {
	case "":
		fmt.Fprintf(o.Out, "MicroShift Version: %s\n", versionInfo.String())
		fmt.Fprintf(o.Out, "Base OCP Version: %s\n", release.Base)
	case "yaml":
		marshalled, err := yaml.Marshal(&versionInfo)
		if err != nil {
			return err
		}
		fmt.Fprintln(o.Out, string(marshalled))
	case "json":
		marshalled, err := json.MarshalIndent(&versionInfo, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(o.Out, string(marshalled))
	default:
		// There is a bug in the program if we hit this case.
		// However, we follow a policy of never panicking.
		return fmt.Errorf("VersionOptions were not validated: --output=%q should have been rejected", o.Output)
	}

	return nil
}
