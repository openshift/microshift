package compose

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewComposeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compose target",
		Short: "",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		repoFlag := cmd.Flag("repo")
		if repoFlag == nil {
			return fmt.Errorf("repo flag is nil")
		}

		td, err := NewTemplatingData(repoFlag.Value.String())
		if err != nil {
			return err
		}
		_ = td
		return nil
	}

	return cmd
}
