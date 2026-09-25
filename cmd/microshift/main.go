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
	code := runCommand(command, os.Args[1:])
	os.Exit(code)
}

func runCommand(command *cobra.Command, args []string) int {
	// Certificate commands own their JSON/YAML error output. Keep the existing
	// logging and diagnostic behavior for every other command.
	selected, _, _ := command.Find(args)
	for current := selected; current != nil; current = current.Parent() {
		if current.Name() == "certs" {
			return cmds.RunCertsCommand(command, args)
		}
	}
	command.SetArgs(args)
	return cli.Run(command)
}

func newCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "microshift",
		Short: "MicroShift, a minimal OpenShift",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help() // err is always nil
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
	cmd.AddCommand(cmds.NewBackupCommand())
	cmd.AddCommand(cmds.NewRestoreCommand())
	cmd.AddCommand(cmds.NewHealthcheckCommand())
	cmd.AddCommand(cmds.NewCertsCommand(ioStreams))
	cmd.AddCommand(cmds.NewAddNodeCommand())
	cmd.AddCommand(cmds.NewC2CCProbeCommand())
	return cmd
}
