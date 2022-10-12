package main

import (
	"os"

	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/component-base/cli"

	cmds "github.com/openshift/microshift/pkg/cmd"
	"github.com/openshift/microshift/pkg/config"
)

func main() {
	command := newCommand()
	code := cli.Run(command)
	os.Exit(code)
}

func newCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "microshift",
		Short: "MicroShift, a minimal OpenShift",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			os.Exit(1)
		},
	}
	originalHelpFunc := cmd.HelpFunc()
	cmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		config.HideUnsupportedFlags(command.Flags())
		originalHelpFunc(command, strings)
	})

	ioStreams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}

	cmd.AddCommand(cmds.NewRunMicroshiftCommand())
	cmd.AddCommand(cmds.NewVersionCommand(ioStreams))
	cmd.AddCommand(cmds.NewShowConfigCommand(ioStreams))
	return cmd
}
