package compose

import (
	"github.com/spf13/cobra"
)

func NewComposeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compose target",
		Short: "",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return nil
	}

	return cmd
}
