package compose

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func NewComposeCmd() *cobra.Command {
	templatingDataInput := ""

	cmd := &cobra.Command{
		Use:   "compose target",
		Short: "",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		repoFlag := cmd.Flag("repo")
		if repoFlag == nil {
			return fmt.Errorf("repo flag is nil")
		}

		td, err := NewTemplatingData(repoFlag.Value.String(), templatingDataInput)
		if err != nil {
			return err
		}
		_ = td
		return nil
	}

	cmd.PersistentFlags().StringVar(&templatingDataInput, "templating-data", "", "Provide path to partial templating data to skip querying remote repository.")

	cmd.AddCommand(templatingDataSubCmd())

	return cmd
}

func templatingDataSubCmd() *cobra.Command {
	full := false

	cmd := &cobra.Command{
		Use:   "templating-data",
		Short: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoFlag := cmd.Flag("repo")
			if repoFlag == nil {
				return fmt.Errorf("repo flag is nil")
			}

			td, err := NewTemplatingData(repoFlag.Value.String(), "")
			if err != nil {
				return err
			}

			// Only serialize entire templating data if requested.
			if full {
				b, err := json.MarshalIndent(td, "", "    ")
				if err != nil {
					return fmt.Errorf("failed to marshal templating data to json: %w", err)
				}
				fmt.Printf("%s", string(b))
				return nil
			}

			// By default this will only include information that change less often (i.e. RHOCP and OpenShift mirror related) and take longer to obtain.
			// Information obtained from local files is quick and can change more often.
			reducedTD := make(map[string]interface{})
			reducedTD["Current"] = td.Current
			reducedTD["Previous"] = td.Previous
			reducedTD["YMinus2"] = td.YMinus2
			reducedTD["RHOCPMinorY"] = td.RHOCPMinorY
			reducedTD["RHOCPMinorY1"] = td.RHOCPMinorY1
			reducedTD["RHOCPMinorY2"] = td.RHOCPMinorY2
			b, err := json.MarshalIndent(reducedTD, "", "    ")
			if err != nil {
				return fmt.Errorf("failed to marshal reduced templating data to json: %w", err)
			}
			fmt.Printf("%s", string(b))

			return nil
		},
	}

	cmd.Flags().BoolVar(&full, "full", false, "Obtain full templating data, including local RPM information (source, base, fake)")

	return cmd
}
