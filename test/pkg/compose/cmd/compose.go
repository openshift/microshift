package compose

import "github.com/spf13/cobra"

func NewComposeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "compose",
	}
	return cmd
}
