package main

import (
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/component-base/cli"
	"k8s.io/component-base/logs"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	cmd := &cobra.Command{
		Use:   "microshift-etcd",
		Short: "MicroShift, a minimal OpenShift",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			os.Exit(1)
		},
	}

	cmd.AddCommand(NewRunEtcdCommand())
	cmd.AddCommand(NewVersionCommand(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}))

	os.Exit(cli.Run(cmd))
}
