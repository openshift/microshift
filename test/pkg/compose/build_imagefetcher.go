package compose

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
	dest := filepath.Join(i.Opts.ComposeOpts.ArtifactsMainDir, "vm-storage", i.Name+".iso")
	tmpDest := dest + ".download"

	if exists, err := util.PathExistsAndIsNotEmpty(dest); err != nil {
		return err
	} else if exists {
		klog.InfoS("Image for download already exists", "path", dest)
		if !i.Opts.ComposeOpts.Force {
			return nil
		}
		klog.InfoS("Force mode: removing and re-downloading file", "path", dest)
		if err := os.RemoveAll(dest); err != nil {
			return fmt.Errorf("failed to remove %q: %w", dest, err)
		}
	}

	timeout := 20 * time.Minute
	klog.InfoS("Downloading image", "name", i.Name, "destination", dest, "timeout", timeout)
	start := time.Now()

	client := http.Client{Timeout: timeout}
	resp, err := client.Get(i.Url)
	if err != nil {
		return fmt.Errorf("client.Get(%q) failed: %w", i.Url, err)
	}
	defer resp.Body.Close()

	f, err := os.Create(tmpDest)
	if err != nil {
		return fmt.Errorf("failed to create file %q for downloading: %w", dest, err)
	}

	n, err := io.Copy(f, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to copy %q to %q: %w", i.Url, dest, err)
	}

	err = os.Rename(tmpDest, dest)
	if err != nil {
		return fmt.Errorf("failed to rename %q to %q: %w", tmpDest, dest, err)
	}

	klog.InfoS("Downloaded image", "destination", dest, "url", i.Url, "sizeMB", n/1_048_576, "duration", time.Since(start))

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
