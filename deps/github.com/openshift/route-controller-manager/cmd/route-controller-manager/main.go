package main

import (
	"os"

	"github.com/spf13/cobra"
	"k8s.io/component-base/cli"

	route_controller_manager "github.com/openshift/route-controller-manager/pkg/cmd/route-controller-manager"
)

func main() {
	command := NewRouteControllerManagerCommand()
	code := cli.Run(command)
	os.Exit(code)
}

func NewRouteControllerManagerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "route-controller-manager",
		Short: "Command for additional management of Ingress and Route resources",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			os.Exit(1)
		},
	}
	start := route_controller_manager.NewRouteControllerManagerCommand("start")
	cmd.AddCommand(start)

	return cmd
}
