package main

import (
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"k8s.io/component-base/logs"

	cmds "github.com/openshift/microshift/pkg/cmd"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	pflag.CommandLine.SetNormalizeFunc(utilflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	logs.InitLogs()
	defer logs.FlushLogs()

	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	command := newCommand()
	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func newCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ushift",
		Short: "Micro OpenShift",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			os.Exit(1)
		},
	}
	cmd.AddCommand(cmds.InitCmd)
	cmd.AddCommand(cmds.ControllerCmd)
	cmd.AddCommand(cmds.NodeCmd)
}
