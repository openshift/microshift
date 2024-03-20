package sources

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"github.com/openshift/microshift/test/pkg/compose/helpers"
	"github.com/openshift/microshift/test/pkg/compose/templatingdata"
	"github.com/openshift/microshift/test/pkg/testutil"

	"k8s.io/klog/v2"
)

type SourceConfigurerOpts struct {
	Composer    helpers.Composer
	TplData     *templatingdata.TemplatingData
	TestDirPath string
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

		funcs := map[string]any{"hasPrefix": strings.HasPrefix}
		tpl, err := template.New(name).Funcs(funcs).Parse(data)
		if err != nil {
			klog.ErrorS(err, "Failed to parse template file", "template", name, "filepath", path)
			return err
		}

		b := &strings.Builder{}
		err = tpl.Execute(b, sc.Opts.TplData)
		if err != nil {
			klog.ErrorS(err, "Executing template failed", "template", path)
			return err
		}
		result := b.String()
		klog.InfoS("Template templatized", "name", name, "result", result)

		if len(result) == 0 {
			if slices.Contains(existingSources, name) {
				klog.InfoS("Template is empty but exists in composer - removing", "name", name)
				if err := sc.Opts.Composer.DeleteSource(name); err != nil {
					klog.ErrorS(err, "Deleting composer source failed")
					return err
				}
			} else {
				klog.InfoS("Template is empty - not adding", "name", name)
			}
			return nil
		}

		klog.InfoS("Adding source to the composer", "name", name)
		if err := sc.Opts.Composer.AddSource(result); err != nil {
			klog.ErrorS(err, "Adding composer source failed")
			return err
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to add sources to the composer: %w", err)
	}

	return nil
}
