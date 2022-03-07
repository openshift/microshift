package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/component-base/logs"

	cmds "github.com/openshift/microshift/pkg/cmd"
	"github.com/openshift/microshift/pkg/config"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	config.InitGlobalFlags()

	logs.InitLogs()
	defer logs.FlushLogs()

	command := newCommand()
	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
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

	ioStreams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}

	cmd.AddCommand(cmds.NewRunMicroshiftCommand())
	cmd.AddCommand(cmds.NewVersionCommand(ioStreams))
	return cmd
}
