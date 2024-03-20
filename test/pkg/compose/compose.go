package compose

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/openshift/microshift/pkg/util"
	"github.com/openshift/microshift/test/pkg/compose/build"
	"github.com/openshift/microshift/test/pkg/compose/helpers"
	"github.com/openshift/microshift/test/pkg/compose/sources"
	"github.com/openshift/microshift/test/pkg/compose/templatingdata"
	"github.com/spf13/cobra"
)

var (
	microShiftRepoRootPath string
	testDirPath            string
	artifactsMainDir       string

	templatingDataFragmentFilepath string

	hostIP string

	force           bool
	dryRun          bool
	buildInstallers bool
	sourceOnly      bool
)

func NewComposeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compose targets...",
		Short: "",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			testDir := cmd.Flag("test-dir").Value.String()
			testDirAbs, err := filepath.Abs(testDir)
			if err != nil {
				return err
			}

			testDirPath = testDirAbs
			microShiftRepoRootPath = filepath.Join(testDirAbs, "..")
			artifactsMainDir = filepath.Join(microShiftRepoRootPath, "_output", "test-images")

			hostIP, err = util.GetHostIP("")
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			args = []string{
				"./image-blueprints/layer1-base",
				"./image-blueprints/layer2-presubmit",
				"./image-blueprints/layer3-periodic",
			}

			fmt.Printf("No argument provided - running following layers: %v\n", args)
		}

		var composer helpers.Composer
		var ostree helpers.Ostree
		if dryRun {
			ostree = helpers.NewDryRunOstree()
			composer = helpers.NewDryRunComposer()
		} else {
			ostree = helpers.NewOstree(filepath.Join(artifactsMainDir, "repo"))
			composer = helpers.NewComposer(testDirPath, fmt.Sprintf("http://%s:8080/repo", hostIP))
		}

		td, err := templatingdata.New(&templatingdata.TemplatingDataOpts{
			ArtifactsMainDir:               artifactsMainDir,
			TemplatingDataFragmentFilepath: templatingDataFragmentFilepath,
		})
		if err != nil {
			return err
		}

		sourceConfigurer := sources.SourceConfigurer{Opts: &sources.SourceConfigurerOpts{
			Composer:    composer,
			TplData:     td,
			TestDirPath: testDirPath,
		}}
		if err := sourceConfigurer.ConfigureSources(); err != nil {
			return err
		}

		blueprintsPath := filepath.Join(testDirPath, "image-blueprints")
		buildPlanner := build.Planner{
			Opts: &build.PlannerOpts{
				Filesys:          os.DirFS(blueprintsPath),
				TplData:          td,
				SourceOnly:       sourceOnly,
				BuildInstallers:  buildInstallers,
				ArtifactsMainDir: artifactsMainDir,
			},
		}

		buildPaths := []string{}
		for _, arg := range args {
			// As a result of using os.DirFS starting at ./test/image-blueprints these paths need to be carefully crafted
			buildPath := filepath.Join(testDirPath, arg)
			buildPath = strings.TrimLeft(strings.ReplaceAll(buildPath, blueprintsPath, ""), "/")
			buildPaths = append(buildPaths, buildPath)
		}
		toBuild, err := buildPlanner.ConstructBuildTree(buildPaths)
		if err != nil {
			return err
		}

		builder := build.Runner{
			Opts: &build.Opts{
				Composer:         composer,
				Ostree:           ostree,
				Force:            force,
				DryRun:           dryRun,
				ArtifactsMainDir: artifactsMainDir,
			},
		}
		err = builder.Build(toBuild)
		if err != nil {
			return err
		}

		return nil
	}

	cmd.PersistentFlags().StringVar(&templatingDataFragmentFilepath, "templating-data", "", "Provide path to partial templating data to skip querying remote repository.")
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
			td, err := templatingdata.New(&templatingdata.TemplatingDataOpts{
				ArtifactsMainDir:               artifactsMainDir,
				TemplatingDataFragmentFilepath: templatingDataFragmentFilepath,
			})
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
