package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/openshift/microshift/pkg/release"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

type imagesOptions struct {
	Output string
}

func NewImagesCommand(ioStreams genericclioptions.IOStreams) *cobra.Command {
	opts := imagesOptions{}

	cmd := &cobra.Command{
		Use:   "images",
		Short: "Print container images used by this MicroShift version",
		Run: func(cmd *cobra.Command, args []string) {
			var marshalled []byte
			var err error

			switch opts.Output {
			case "":
				separator := ""
				for _, v := range release.Image {
					fmt.Fprintf(ioStreams.Out, "%s%s", separator, v)
					separator = " "
				}
			case "json":
				marshalled, err = json.MarshalIndent(&release.Image, "", "  ")
				if err != nil {
					cmdutil.CheckErr(err)
				}
				fmt.Fprintf(ioStreams.Out, "%s\n", string(marshalled))
			case "yaml":
				marshalled, err = yaml.Marshal(&release.Image)
				if err != nil {
					cmdutil.CheckErr(err)
				}
				fmt.Fprintf(ioStreams.Out, "%s", string(marshalled))
			case "toml":
				for _, v := range release.Image {
					fmt.Fprintf(ioStreams.Out, "[[containers]]\nsource = \"%v\"\n\n", v)
				}
			default:
				cmdutil.CheckErr(fmt.Errorf("Unknown output format %q", opts.Output))
			}
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Output, "output", "o", opts.Output, "One of 'json', 'yaml' or 'toml'.")
	return cmd
}
