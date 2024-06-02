package compose

import (
	"github.com/openshift/microshift/test/pkg/compose/templatingdata"
	"github.com/openshift/microshift/test/pkg/util"
	"k8s.io/klog/v2"

	"github.com/spf13/cobra"
)

// Variables common to all compose (sub)commands. Set by `compose` PreRun.
var (
	// Structure containing all relevants filesystem paths so no module needs to calcualate them individually.
	paths *util.Paths

	tplData *templatingdata.TemplatingData

	templatingDataFragmentFilepath string
	skipContainerImagesExtraction  bool
)

func NewComposeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "compose",

		PersistentPreRunE: composePreRun,
	}

	cmd.PersistentFlags().StringVar(&templatingDataFragmentFilepath,
		"templating-data", "",
		"Path to partial templating data to skip querying remote repository. ")

	cmd.PersistentFlags().BoolVarP(&skipContainerImagesExtraction,
		"skip-container-images-extraction", "E", false,
		"Skip extraction of images from microshift-release-info RPMs")

	cmd.AddCommand(newTemplatingDataCmd())
	cmd.AddCommand(newBuildCmd())

	return cmd
}

func composePreRun(cmd *cobra.Command, args []string) error {
	var err error

	paths, err = util.NewPaths()
	if err != nil {
		return err
	}
	klog.InfoS("Constructed Paths struct", "paths", paths)

	tplDataOpts := &templatingdata.TemplatingDataOpts{
		Paths:                          paths,
		TemplatingDataFragmentFilepath: templatingDataFragmentFilepath,
		SkipContainerImagesExtraction:  skipContainerImagesExtraction,
	}
	tplData, err = tplDataOpts.Construct()
	if err != nil {
		return err
	}
	klog.InfoS("Constructed TemplatingData struct", "TemplatingData", tplData)

	return nil
}
