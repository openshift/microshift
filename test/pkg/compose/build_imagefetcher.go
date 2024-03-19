package compose

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"k8s.io/klog/v2"
)

type ImageFetcher struct {
	build
	Url string
}

func NewImageFetcher(path string, opts *BuildOpts) (*ImageFetcher, error) {
	filename := filepath.Base(path)
	withoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))

	dataBytes, err := fs.ReadFile(opts.Filesys, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	tpl, err := template.New(withoutExt).Parse(string(dataBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", "", err)
	}
	templatedData := strings.Builder{}
	if err := tpl.Execute(&templatedData, opts.TplData); err != nil {
		return nil, fmt.Errorf("failed to execute template %q: %w", path, err)
	}

	return &ImageFetcher{
		build: build{
			Name: withoutExt,
			Path: path,
			Opts: opts,
		},
		Url: templatedData.String(),
	}, nil
}

func (i *ImageFetcher) Execute() error {
	klog.InfoS("Downloading image", "name", i.Name)
	time.Sleep(1 * time.Second)
	klog.InfoS("Image downloaded", "name", i.Name)
	return nil
}

func fileExistsInDir(filesys fs.FS, dir, filename string) (bool, error) {
	entries, err := fs.ReadDir(filesys, dir)
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		if entry.Name() == filename {
			return true, nil
		}
	}
	return false, nil
}
