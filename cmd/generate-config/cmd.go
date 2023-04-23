package main

import (
	"os"

	"github.com/spf13/cobra"
	"k8s.io/component-base/cli"
)

func main() {
	command := newCommand()
	code := cli.Run(command)
	os.Exit(code)
}

func newCommand() *cobra.Command {
	opt := configGenOpts{}

	cmd := &cobra.Command{
		Use:   "generate-config",
		Short: "use openapiv3 schemas in CRDs format to generate yaml or embed in files",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opt.Options(); err != nil {
				return err
			}
			return opt.Run()
		},
	}
	opt.BindFlags(cmd.Flags())

	return cmd
}
