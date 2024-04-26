package compose

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/openshift/microshift/pkg/util"
	"github.com/openshift/microshift/test/pkg/compose/build"
	"github.com/openshift/microshift/test/pkg/compose/helpers"
	"github.com/openshift/microshift/test/pkg/compose/sources"
	"github.com/openshift/microshift/test/pkg/compose/templatingdata"
	"github.com/openshift/microshift/test/pkg/testutil"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

var (
	paths *testutil.Paths

	templatingDataFragmentFilepath string

	hostIP string

	force                         bool
	dryRun                        bool
	buildInstallers               bool
	sourceOnly                    bool
	skipContainerImagesExtraction bool
	skipOSBuildLogCollection      bool
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

			if sourceOnly {
				buildInstallers = true
				force = true
			}
			paths, err = testutil.NewPaths(testDirAbs)
			if err != nil {
				return err
			}
			klog.InfoS("Constructed Paths struct", "path", paths)

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
				// Assumption: CWD is test/
				"./image-blueprints/layer1-base",
				"./image-blueprints/layer2-presubmit",
				"./image-blueprints/layer3-periodic",
			}

			klog.InfoS("No argument provided - running default set of layers", "layers", args)
		}

		if !skipOSBuildLogCollection {
			defer func() {
				if err := saveOSBuildLogs(paths.BuildLogsDir); err != nil {
					klog.ErrorS(err, "Failed to save logs of osbuild service(s)")
				}
			}()
		}

		composer, ostree, podman, err := getHelpers()
		if err != nil {
			return err
		}

		td, err := (&templatingdata.TemplatingDataOpts{
			Paths:                          paths,
			TemplatingDataFragmentFilepath: templatingDataFragmentFilepath,
			SkipContainerImagesExtraction:  skipContainerImagesExtraction,
		}).Construct()
		if err != nil {
			return err
		}

		if !skipContainerImagesExtraction {
			if err := persistImages(td); err != nil {
				return err
			}
		}

		events := testutil.NewEventManager("compose")
		defer func() {
			junitFile := filepath.Join(paths.BuildLogsDir, "junit_compose.xml")
			junit := events.GetJUnit()
			err = junit.WriteToFile(junitFile)
			if err != nil {
				klog.ErrorS(err, "Failed to write junit to a file", "file", junitFile)
			}

			intervalsFile := filepath.Join(paths.BuildLogsDir, "intervals_compose.json")
			timelinesFile := filepath.Join(paths.BuildLogsDir, "e2e-timelines_spyglass_compose.html")
			err = events.WriteToFiles(intervalsFile, timelinesFile)
			if err != nil {
				klog.ErrorS(err, "Failed to write events to a files", "file", timelinesFile)
			}
		}()

		fileSystem := os.DirFS(paths.MicroShiftRepoRootPath)
		testFS, err := fs.Sub(fileSystem, "test")
		if err != nil {
			klog.ErrorS(err, "Failed to get 'test' subFS")
			return err
		}
		sourcesFS, err := fs.Sub(testFS, "package-sources")
		if err != nil {
			klog.ErrorS(err, "Failed to get 'package-sources' subFS")
			return err
		}
		sourceConfigurer := sources.SourceConfigurer{Opts: &sources.SourceConfigurerOpts{
			Composer:  composer,
			TplData:   td,
			Events:    events,
			SourcesFS: sourcesFS,
		}}
		if err := sourceConfigurer.ConfigureSources(); err != nil {
			return err
		}

		blueprintsFS, err := fs.Sub(testFS, "image-blueprints")
		if err != nil {
			klog.ErrorS(err, "Failed to get 'image-blueprints' subFS")
			return err
		}
		buildPlanner := build.Planner{
			Opts: &build.PlannerOpts{
				TplData:         td,
				SourceOnly:      sourceOnly,
				BuildInstallers: buildInstallers,
				Proxy:           build.NewProxy(),
				BlueprintsFS:    blueprintsFS,
				Paths:           paths,
				Events:          events,
			},
		}

		buildPaths, err := argsToBuildPaths(args)
		if err != nil {
			return err
		}
		buildPlan, err := buildPlanner.CreateBuildPlan(buildPaths)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(context.Background())
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			sig := <-sigs
			klog.InfoS("Received signal - canceling context", "signal", sig)
			cancel()
		}()

		builder := build.Runner{
			Opts: &build.Opts{
				Composer: composer,
				Ostree:   ostree,
				Podman:   podman,
				Force:    force,
				DryRun:   dryRun,
				Paths:    paths,
				Events:   events,
			},
		}
		err = builder.Build(ctx, buildPlan)
		if err != nil {
			return err
		}

		return nil
	}

	cmd.PersistentFlags().StringVar(&templatingDataFragmentFilepath, "templating-data", "", "Provide path to partial templating data to skip querying remote repository.")
	cmd.PersistentFlags().BoolVarP(&buildInstallers, "build-installers", "I", true, "Build ISO image installers.")
	cmd.PersistentFlags().BoolVarP(&sourceOnly, "source-only", "s", false, "Build only source blueprints. Implies --build-installers and --force.")
	cmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "Dry run - no real interaction with the Composer")
	cmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "Rebuild existing artifacts (ostree commits, ISO images)")
	cmd.PersistentFlags().BoolVarP(&skipContainerImagesExtraction, "skip-container-images-extraction", "E", false, "Skip extraction of images from microshift-release-info RPMs")
	cmd.PersistentFlags().BoolVarP(&skipOSBuildLogCollection, "skip-osbuild-log-collection", "L", false, "Skip collection of osbuild logs (journals)")

	cmd.AddCommand(templatingDataSubCmd())
	cmd.AddCommand(buildPlanSubCmd())

	return cmd
}

func templatingDataSubCmd() *cobra.Command {
	full := false

	cmd := &cobra.Command{
		Use:   "templating-data",
		Short: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			td, err := (&templatingdata.TemplatingDataOpts{
				Paths:                          paths,
				TemplatingDataFragmentFilepath: templatingDataFragmentFilepath,
				SkipContainerImagesExtraction:  skipContainerImagesExtraction,
			}).Construct()
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

func buildPlanSubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build-plan",
		Short: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				args = []string{
					"./image-blueprints/layer1-base",
					"./image-blueprints/layer2-presubmit",
					"./image-blueprints/layer3-periodic",
				}

				klog.InfoS("No argument provided - running default set of layers", "layers", args)
			}

			td, err := (&templatingdata.TemplatingDataOpts{
				Paths:                          paths,
				TemplatingDataFragmentFilepath: templatingDataFragmentFilepath,
			}).Construct()
			if err != nil {
				return err
			}

			buildPlanner := build.Planner{
				Opts: &build.PlannerOpts{
					TplData:         td,
					SourceOnly:      sourceOnly,
					BuildInstallers: buildInstallers,
					Paths:           paths,
				},
			}
			buildPaths, err := argsToBuildPaths(args)
			if err != nil {
				return err
			}
			buildPlan, err := buildPlanner.CreateBuildPlan(buildPaths)
			if err != nil {
				return err
			}
			b, err := json.MarshalIndent(buildPlan, "", "    ")
			if err != nil {
				return fmt.Errorf("failed to marshal build plan to json: %w", err)
			}
			fmt.Printf("%s", string(b))

			return nil
		},
	}

	return cmd
}

func getHelpers() (helpers.Composer, helpers.Ostree, helpers.Podman, error) {
	if dryRun {
		return helpers.NewDryRunComposer(), helpers.NewDryRunOstree(), helpers.NewDryRunPodman(), nil
	}

	ostree, err := helpers.NewOstree(paths.OSTreeRepoDir)
	if err != nil {
		return nil, nil, nil, err
	}
	composer, err := helpers.NewComposer(paths, fmt.Sprintf("http://%s:8080/repo", hostIP))
	if err != nil {
		return nil, nil, nil, err
	}

	return composer, ostree, helpers.NewPodman(), nil
}

func persistImages(td *templatingdata.TemplatingData) error {
	dest := filepath.Join(paths.ArtifactsMainDir, "container-images-list")
	klog.InfoS("Writing all image references from TemplatingData to a file", "path", dest)

	images := []string{}
	images = append(images, td.Base.Images...)
	images = append(images, td.FakeNext.Images...)
	images = append(images, td.Source.Images...)
	images = append(images, td.Current.Images...)
	images = append(images, td.Previous.Images...)
	images = append(images, td.YMinus2.Images...)

	imageSet := make(map[string]struct{})
	for _, img := range images {
		imageSet[img] = struct{}{}
	}

	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create file %q: %w", dest, err)
	}
	defer f.Close()

	for img := range imageSet {
		_, err := f.WriteString(img + "\n")
		if err != nil {
			return fmt.Errorf("failed to write %q to file %q: %w", img, dest, err)
		}
	}

	return nil
}

func saveOSBuildLogs(dir string) error {
	conn, err := dbus.NewWithContext(context.Background())
	if err != nil {
		return err
	}
	defer conn.Close()

	units, err := conn.ListUnitsByPatternsContext(context.Background(), []string{}, []string{"osbuild-worker*.service", "osbuild-composer*.service"})
	if err != nil {
		return err
	}

	errs := []error{}

	for _, unit := range units {
		logFile := filepath.Join(dir, unit.Name+".log")
		_, _, err := testutil.RunCommand("bash", "-c", fmt.Sprintf("sudo journalctl -u %s &> %s", unit.Name, logFile))
		if err != nil {
			klog.ErrorS(err, "Failed to write journal of osbuild service", "service", unit, "filepath", logFile)
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func argsToBuildPaths(args []string) ([]string, error) {
	buildPaths := []string{}
	for _, arg := range args {
		abs_arg, err := filepath.Abs(arg)
		if err != nil {
			klog.ErrorS(err, "failed to convert arg to absolute path", "arg", arg)
			return nil, err
		}
		rel_arg, err := filepath.Rel(paths.ImageBlueprintsPath, abs_arg)
		if err != nil {
			klog.ErrorS(err, "failed to convert absolute path to relative path", "arg", arg, "abs_arg", abs_arg)
			return nil, err
		}
		buildPaths = append(buildPaths, rel_arg)
	}
	return buildPaths, nil
}
