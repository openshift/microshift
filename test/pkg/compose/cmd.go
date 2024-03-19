package compose

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func NewComposeCmd() *cobra.Command {
	templatingDataInput := ""
	buildInstallers := true
	sourceOnly := false
	dryRun := false
	force := false

	cmd := &cobra.Command{
		Use:   "compose target",
		Short: "",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("argument must be provided")
		}

		testDirFlag := cmd.Flag("test-dir")
		if testDirFlag == nil {
			return fmt.Errorf("repo flag is nil")
		}
		testDir := testDirFlag.Value.String()

		td, err := NewTemplatingData(testDir, templatingDataInput)
		if err != nil {
			return err
		}

		var composer Composer
		if dryRun {
			composer = NewDryRunComposer()
		} else {
			composer = NewComposer(testDir)
		}

		var ostree Ostree
		if dryRun {
			ostree = NewDryRunOstree()
		} else {
			ostree = NewOstree(filepath.Join(testDir, "..", "_output", "test-images", "repo"))
		}

		if err := NewSourceConfigurer(composer, td).ConfigureSources(); err != nil {
			return err
		}

		blueprintPath := filepath.Join(testDir, "image-blueprints")
		blueprints, err := filepath.Abs(blueprintPath)
		if err != nil {
			return err
		}
		blueprintsDirfilesys := os.DirFS(blueprints)

		buildPath, err := filepath.Abs(filepath.Join(testDir, args[0]))
		if err != nil {
			return err
		}
		buildPath = strings.TrimLeft(strings.ReplaceAll(buildPath, blueprints, ""), "/")
		buildPlanner := BuildPlanner{
			Opts: &BuildOpts{
				Filesys:         blueprintsDirfilesys,
				BuildInstallers: buildInstallers,
				SourceOnly:      sourceOnly,
				Force:           force,
				TplData:         td,
				Composer:        composer,
				Ostree:          ostree,
			},
		}

		toBuild, err := buildPlanner.ConstructBuildTree(buildPath)
		if err != nil {
			return err
		}

		builder := BuildRunner{}
		err = builder.Build(toBuild)
		if err != nil {
			return err
		}

		return nil
	}

	cmd.PersistentFlags().StringVar(&templatingDataInput, "templating-data", "", "Provide path to partial templating data to skip querying remote repository.")
	cmd.PersistentFlags().BoolVarP(&buildInstallers, "build-installers", "I", true, "Build ISO image installers.")
	cmd.PersistentFlags().BoolVarP(&sourceOnly, "source-only", "s", false, "Build only source blueprints.")
	cmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "Dry run - no real interaction with the Composer")
	cmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "Rebuild existing artifacts (ostree commits, ISO images)")

	cmd.AddCommand(templatingDataSubCmd())

	return cmd
}

func templatingDataSubCmd() *cobra.Command {
	full := false

	cmd := &cobra.Command{
		Use:   "templating-data",
		Short: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			testDirFlag := cmd.Flag("test-dir")
			if testDirFlag == nil {
				return fmt.Errorf("repo flag is nil")
			}
			testDir := testDirFlag.Value.String()

			td, err := NewTemplatingData(testDir, "")
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
