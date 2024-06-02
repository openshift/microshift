package compose

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	uutil "github.com/openshift/microshift/pkg/util"
	"github.com/openshift/microshift/test/pkg/compose/helpers"
	"github.com/openshift/microshift/test/pkg/compose/sources"
	"github.com/openshift/microshift/test/pkg/compose/templatingdata"
	"github.com/openshift/microshift/test/pkg/util"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

func newBuildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Run whole pipeline to build images.",

		RunE: func(cmd *cobra.Command, args []string) error {
			hostIP, err := uutil.GetHostIP("")
			if err != nil {
				return err
			}
			composer, err := helpers.NewComposer(paths, fmt.Sprintf("http://%s:8080/repo", hostIP))
			if err != nil {
				return err
			}
			events := util.NewEventManager("build")
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

			opts := &sources.SourceConfigurerOpts{
				Composer:  composer,
				TplData:   templatingdata.NewShim(tplData),
				Events:    events,
				SourcesFS: sourcesFS,
			}
			sourceConfigurer := sources.SourceConfigurer{Opts: opts}
			if err := sourceConfigurer.ConfigureSources(); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}
