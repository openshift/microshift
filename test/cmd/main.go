package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/openshift/microshift/pkg/util"
	"github.com/spf13/cobra"
	"k8s.io/component-base/cli"
)

var (
	microshiftRepo = ""
)

func main() {
	cmd := &cobra.Command{
		Use:   "microshift-tests",
		Short: "",

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			repo, err := filepath.Abs(microshiftRepo)
			if err != nil {
				return err
			}

			gomodPath := filepath.Join(repo, "go.mod")
			if exists, err := util.PathExists(gomodPath); err != nil {
				return fmt.Errorf("failed checking if %s exists: %w", gomodPath, err)
			} else if !exists {
				return fmt.Errorf("%s does not exists - is the program executed in MicroShift repository root dir? Alternatively use --repo", gomodPath)
			}
			microshiftRepo = repo
			return nil
		},

		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("MicroShift repo: %q\n", microshiftRepo)

			_ = cmd.Help()
			os.Exit(1)
		},
	}

	cmd.PersistentFlags().StringVarP(&microshiftRepo, "repo", "r", ".", "Path to the MicroShift repository")

	os.Exit(cli.Run(cmd))
}
