package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/openshift/microshift/pkg/util"

	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:          "microshift-tests",
		Short:        "",
		SilenceUsage: true,

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			wd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working dir: %w", err)
			}

			// Check if app is executed from microshift/test/ directory
			varPath := filepath.Join(wd, "..", "Makefile.kube_git.var")
			if exists, err := util.PathExists(varPath); err != nil {
				return fmt.Errorf("failed checking if %s exists: %w", varPath, err)
			} else if !exists {
				return fmt.Errorf("could not find %q - microshift-tests must be executed in MicroShift's repository test/", varPath)
			}

			return nil
		},

		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
			os.Exit(1)
		},
	}

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
