package build

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/openshift/microshift/pkg/util"
	"k8s.io/klog/v2"
)

var _ Build = (*ImageFetcher)(nil)

type ImageFetcher struct {
	build
	Url         string
	Destination string
}

func NewImageFetcher(path string, opts *PlannerOpts) (*ImageFetcher, error) {
	klog.InfoS("Constructing ImageFetcher", "path", path)

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
		},
		Url:         templatedData.String(),
		Destination: filepath.Join(opts.ArtifactsMainDir, "vm-storage", withoutExt+".iso"),
	}, nil
}

func (i *ImageFetcher) Prepare(opts *Opts) error {
	return nil
}

func (i *ImageFetcher) Execute(opts *Opts) error {
	if opts.DryRun {
		klog.InfoS("DRY RUN: Downloaded image", "name", i.Name)
		return nil
	}

	tmpDest := i.Destination + ".download"

	if exists, err := util.PathExistsAndIsNotEmpty(i.Destination); err != nil {
		return err
	} else if exists {
		klog.InfoS("Image for download already exists", "path", i.Destination)
		if !opts.Force {
			return nil
		}
		klog.InfoS("Force mode: removing and re-downloading file", "path", i.Destination)
		if err := os.RemoveAll(i.Destination); err != nil {
			return fmt.Errorf("failed to remove %q: %w", i.Destination, err)
		}
	}

	timeout := 20 * time.Minute
	klog.InfoS("Downloading image", "name", i.Name, "destination", i.Destination, "timeout", timeout)
	start := time.Now()

	client := http.Client{Timeout: timeout}
	resp, err := client.Get(i.Url)
	if err != nil {
		return fmt.Errorf("client.Get(%q) failed: %w", i.Url, err)
	}
	defer resp.Body.Close()

	f, err := os.Create(tmpDest)
	if err != nil {
		return fmt.Errorf("failed to create file %q for downloading: %w", i.Destination, err)
	}

	var n int64
	done := make(chan struct{})
	go func() {
		n, err = io.Copy(f, resp.Body)
		done <- struct{}{}
	}()

	ticker := time.NewTicker(time.Second * 30)
outer:
	for {
		select {
		case <-done:
			break outer
		case <-ticker.C:
			klog.InfoS("Waiting for image download", "name", i.Name, "timeout", timeout, "elapsed", time.Since(start))
		}
	}
	ticker.Stop()

	err = os.Rename(tmpDest, i.Destination)
	if err != nil {
		return fmt.Errorf("failed to rename %q to %q: %w", tmpDest, i.Destination, err)
	}

	klog.InfoS("Downloaded image", "destination", i.Destination, "url", i.Url, "sizeMB", n/1_048_576, "duration", time.Since(start))

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
