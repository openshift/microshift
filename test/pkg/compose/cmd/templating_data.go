package compose

import (
	"fmt"

	"github.com/openshift/microshift/test/pkg/compose/templatingdata"

	"github.com/spf13/cobra"
)

func newTemplatingDataCmd() *cobra.Command {
	full := false

	cmd := &cobra.Command{
		Use:   "templating-data",
		Short: "Get templating data with values queried against remotes.",
		Long: `Get templating data with values queried against remotes.

Can be given to other commands to speed up initial phases of build.
For example:
  $ microshift-tests compose templating-data > ~/tplData.json
  $ microshift-tests compose --templating-data ~/tplData.json TARGET`,

		RunE: func(cmd *cobra.Command, args []string) error {
			tplDataOpts := &templatingdata.TemplatingDataOpts{
				Paths:                          paths,
				TemplatingDataFragmentFilepath: templatingDataFragmentFilepath,
				SkipContainerImagesExtraction:  skipContainerImagesExtraction,
			}

			tplData, err := tplDataOpts.Construct()
			if err != nil {
				return err
			}

			var output string
			if full {
				// Serialize whole templating data only on demand.
				// Primarily for debug: if templating-data-fragment is supplied,
				// local values are recalculated anyway because it's cheap and they can change often.
				output, err = tplData.String()
			} else {
				// By default this will only include information that change less often
				// and take longer to obtain (i.e. RHOCP and OpenShift mirror related).
				output, err = tplData.FragmentString()
			}

			if err != nil {
				return err
			}

			fmt.Printf("%s\n", output)

			return nil
		},
	}

	cmd.Flags().BoolVar(&full,
		"full", false,
		"Obtain full templating data, including local RPM information (source, base, fake)")

	return cmd
}
