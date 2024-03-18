package compose

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"github.com/osbuild/weldr-client/v2/weldr"
	"k8s.io/klog/v2"
)

type SourceConfigurer struct {
	composer weldr.Client
	tplData  *TemplatingData
}

func NewSourceConfigurer(composer weldr.Client, tplData *TemplatingData) *SourceConfigurer {
	return &SourceConfigurer{
		composer: composer,
		tplData:  tplData,
	}
}

func (sc *SourceConfigurer) ConfigureSources() error {
	existingSources, err := sc.getComposerSources()
	if err != nil {
		return err
	}

	sourcesDir := filepath.Join(sc.tplData.MicroShiftRepoPath, "test", "package-sources")
	err = filepath.Walk(sourcesDir, func(path string, fileInfo fs.FileInfo, _ error) error {
		if fileInfo.IsDir() {
			return nil
		}

		dataBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %q: %w", path, err)
		}
		data := string(dataBytes)

		name, err := getTOMLFieldValue(data, "id")
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
		err = tpl.Execute(b, sc.tplData)
		if err != nil {
			klog.ErrorS(err, "Executing template failed", "template", path)
			return err
		}
		result := b.String()
		klog.InfoS("Template templatized", "name", name, "result", result)

		if len(result) == 0 {
			if slices.Contains(existingSources, name) {
				klog.InfoS("Template is empty but exists in composer - removing", "name", name)
				if err := sc.deleteComposerSource(name); err != nil {
					klog.ErrorS(err, "Deleting composer source failed")
					return err
				}
			} else {
				klog.InfoS("Template is empty - not adding", "name", name)
			}
			return nil
		}

		klog.InfoS("Adding source", "name", name)
		if err := sc.addComposerSource(result); err != nil {
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

func (sc *SourceConfigurer) getComposerSources() ([]string, error) {
	existingSources, rsp, err := sc.composer.ListSources()
	if err != nil {
		return nil, fmt.Errorf("listing composer sources failed: %w", err)
	}
	if rsp != nil {
		return nil, fmt.Errorf("listing composer sources returned wrong response: %v", rsp)
	}
	klog.InfoS("Existing sources of the composer-cli", "sources", existingSources)
	return existingSources, nil
}

func (sc *SourceConfigurer) deleteComposerSource(id string) error {
	rsp, err := sc.composer.DeleteSource(id)
	if err != nil {
		return fmt.Errorf("deleting composer source failed: %w", err)
	}
	if rsp != nil && !rsp.Status {
		return fmt.Errorf("deleting composer source returned wrong response: %v", rsp)
	}
	return nil
}

func (sc *SourceConfigurer) addComposerSource(source string) error {
	rsp, err := sc.composer.NewSourceTOML(source)
	if err != nil {
		return fmt.Errorf("adding composer source failed: %w", err)
	}
	if rsp != nil && !rsp.Status {
		return fmt.Errorf("adding composer source returned wrong response: %v", rsp)
	}
	return nil
}
