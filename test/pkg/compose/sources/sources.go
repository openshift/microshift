package sources

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/openshift/microshift/test/pkg/compose/helpers"
	"github.com/openshift/microshift/test/pkg/compose/templatingdata"
	"github.com/openshift/microshift/test/pkg/testutil"

	"k8s.io/klog/v2"
)

type SourceConfigurerOpts struct {
	Composer    helpers.Composer
	TplData     *templatingdata.TemplatingData
	TestDirPath string
	Events      *testutil.EventManager
}

type SourceConfigurer struct {
	Opts *SourceConfigurerOpts
}

func (sc *SourceConfigurer) ConfigureSources() error {
	existingSources, err := sc.Opts.Composer.ListSources()
	if err != nil {
		return err
	}

	sourcesDir := filepath.Join(sc.Opts.TestDirPath, "package-sources")
	err = filepath.Walk(sourcesDir, func(path string, fileInfo fs.FileInfo, _ error) error {
		if fileInfo.IsDir() {
			return nil
		}
		start := time.Now()

		dataBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %q: %w", path, err)
		}
		data := string(dataBytes)

		// Get source name/id directly from the TOML file to not operate on assumption
		// that filename without extension is name of the composer Source.
		name, err := testutil.GetTOMLFieldValue(data, "id")
		if err != nil {
			return err
		}

		result, err := sc.Opts.TplData.Template(name, data)
		if err != nil {
			return err
		}

		if len(result) == 0 {
			if slices.Contains(existingSources, name) {
				klog.InfoS("Template is empty but exists in composer - removing", "name", name)
				if err := sc.Opts.Composer.DeleteSource(name); err != nil {
					klog.ErrorS(err, "Deleting composer source failed")
					return err
				}
			} else {
				klog.InfoS("Template is empty - not adding", "name", name)
				sc.Opts.Events.AddEvent(&testutil.SkippedEvent{
					Event: testutil.Event{
						Name:      name,
						Suite:     "sources",
						ClassName: "source",
						Start:     start,
						End:       time.Now(),
					},
					Message: "Empty result of templating",
				})
			}
			return nil
		}

		klog.InfoS("Adding source to the composer", "name", name)
		if err := sc.Opts.Composer.AddSource(result); err != nil {
			klog.ErrorS(err, "Adding composer source failed")
			sc.Opts.Events.AddEvent(&testutil.FailedEvent{
				Event: testutil.Event{
					Name:      name,
					Suite:     "sources",
					ClassName: "source",
					Start:     start,
					End:       time.Now(),
				},
				Message: "Adding composer source failed",
				Content: err.Error(),
			})
			return err
		}

		sc.Opts.Events.AddEvent(&testutil.Event{
			Name:      name,
			Suite:     "sources",
			ClassName: "source",
			Start:     start,
			End:       time.Now(),
		})

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to add sources to the composer: %w", err)
	}

	return nil
}
