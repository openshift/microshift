package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/openshift/microshift/pkg/util"
	"github.com/openshift/microshift/test/pkg/compose"
	"github.com/spf13/cobra"
)

var (
	testDir = ""
)

func main() {
	cmd := &cobra.Command{
		Use:          "microshift-tests",
		Short:        "",
		SilenceUsage: true,

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			dir, err := filepath.Abs(testDir)
			if err != nil {
				return err
			}

			// Verify if parent dir is MicroShift's repo root
			varPath := filepath.Join(dir, "..", "Makefile.kube_git.var")
			if exists, err := util.PathExists(varPath); err != nil {
				return fmt.Errorf("failed checking if %s exists: %w", varPath, err)
			} else if !exists {
				return fmt.Errorf("could not find Makefile.kube_git.var in working directory's parent (%q) - is the program executed in MicroShift repository test/ dir? Alternatively use --test-dir", varPath)
			}

			return nil
		},

		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
			os.Exit(1)
		},
	}

	cmd.PersistentFlags().StringVar(&testDir, "test-dir", ".", "Path to the `test/` directory in the MicroShift repository")

	cmd.AddCommand(compose.NewComposeCmd())

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
