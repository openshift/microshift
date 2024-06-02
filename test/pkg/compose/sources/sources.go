package sources

import (
	"errors"
	"fmt"
	"io/fs"
	"slices"
	"time"

	"github.com/openshift/microshift/test/pkg/compose/helpers"
	"github.com/openshift/microshift/test/pkg/compose/templatingdata"
	"github.com/openshift/microshift/test/pkg/util"

	"k8s.io/klog/v2"
)

type SourceConfigurerOpts struct {
	Composer  helpers.Composer
	TplData   *templatingdata.Shim
	Events    util.EventManager
	SourcesFS fs.FS
}

type SourceConfigurer struct {
	Opts *SourceConfigurerOpts
}

func (sc *SourceConfigurer) ConfigureSources() error {
	existingSources, err := sc.Opts.Composer.ListSources()
	if err != nil {
		return err
	}

	entries, err := fs.ReadDir(sc.Opts.SourcesFS, ".")
	if err != nil {
		return err
	}

	errs := []error{}

	for _, entry := range entries {
		if entry.IsDir() {
			klog.InfoS("Ignoring unexpected dir in package-sources/", "name", entry.Name())
			continue
		}

		if err := sc.processSource(entry.Name(), existingSources); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (sc *SourceConfigurer) processSource(filename string, existingSources []string) error {
	start := time.Now()

	dataBytes, err := fs.ReadFile(sc.Opts.SourcesFS, filename)
	if err != nil {
		return fmt.Errorf("failed to read %q: %w", filename, err)
	}

	data := string(dataBytes)

	// Get source name/id directly from the TOML file to not operate on assumption
	// that filename without extension is name of the composer Source.
	name, err := util.GetTOMLFieldValue(data, "id")
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
			sc.Opts.Events.AddEvent(&util.SkippedEvent{
				Event: util.Event{
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
		sc.Opts.Events.AddEvent(&util.FailedEvent{
			Event: util.Event{
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

	sc.Opts.Events.AddEvent(&util.Event{
		Name:      name,
		Suite:     "sources",
		ClassName: "source",
		Start:     start,
		End:       time.Now(),
	})

	return nil
}
