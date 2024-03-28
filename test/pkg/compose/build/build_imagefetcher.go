package build

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/openshift/microshift/pkg/util"
	"github.com/openshift/microshift/test/pkg/testutil"
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

	dataBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	templatedData, err := opts.TplData.Template(filename, string(dataBytes))
	if err != nil {
		return nil, err
	}

	return &ImageFetcher{
		build: build{
			Name: withoutExt,
			Path: path,
		},
		Url:         templatedData,
		Destination: filepath.Join(opts.ArtifactsMainDir, "vm-storage", withoutExt+".iso"),
	}, nil
}

func (i *ImageFetcher) Prepare(opts *Opts) error {
	return nil
}

func (i *ImageFetcher) Execute(opts *Opts) error {
	s := time.Now()
	skipped, err := i.execute(opts)

	if skipped && err == nil {
		opts.Events.AddEvent(&testutil.SkippedEvent{
			Event: testutil.Event{
				Name:      i.Name,
				Suite:     "download",
				ClassName: "iso-download",
				Start:     s,
				End:       time.Now(),
			},
			Message: fmt.Sprintf("Image already exists in %s", i.Destination),
		})
		return nil
	}

	if err != nil {
		opts.Events.AddEvent(&testutil.FailedEvent{
			Event: testutil.Event{
				Name:      i.Name,
				Suite:     "download",
				ClassName: "iso-download",
				Start:     s,
				End:       time.Now(),
			},
			Message: "Downloading image failed",
			Content: err.Error(),
		})
		return err
	}

	opts.Events.AddEvent(&testutil.Event{
		Name:      i.Name,
		Suite:     "download",
		ClassName: "iso-download",
		Start:     s,
		End:       time.Now(),
	})
	return nil
}

func (i *ImageFetcher) execute(opts *Opts) (bool, error) {
	if opts.DryRun {
		klog.InfoS("DRY RUN: Downloaded image", "name", i.Name)
		return false, nil
	}

	tmpDest := i.Destination + ".download"

	if exists, err := util.PathExistsAndIsNotEmpty(i.Destination); err != nil {
		return false, err
	} else if exists {
		klog.InfoS("Image for download already exists", "path", i.Destination)
		if !opts.Force {
			return true, nil
		}
		klog.InfoS("Force mode: removing and re-downloading file", "path", i.Destination)
		if err := os.RemoveAll(i.Destination); err != nil {
			return false, fmt.Errorf("failed to remove %q: %w", i.Destination, err)
		}
	}

	err := testutil.Retry(3, func() error { return i.download(tmpDest) })
	if err != nil {
		klog.ErrorS(err, "Failed to download image", "destination", tmpDest, "url", i.Url)
		return false, err
	}

	err = os.Rename(tmpDest, i.Destination)
	if err != nil {
		return false, fmt.Errorf("failed to rename %q to %q: %w", tmpDest, i.Destination, err)
	}
	klog.InfoS("Renamed image", "destination", i.Destination, "source", tmpDest)

	return false, nil
}

func (i *ImageFetcher) download(dest string) error {
	timeout := 20 * time.Minute
	klog.InfoS("Downloading image", "name", i.Name, "destination", i.Destination, "timeout", timeout)
	start := time.Now()

	client := http.Client{Timeout: timeout}
	resp, err := client.Get(i.Url)
	if err != nil {
		return fmt.Errorf("client.Get(%q) failed: %w", i.Url, err)
	}
	defer resp.Body.Close()

	f, err := os.Create(dest)
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

	if err != nil {
		return err
	}

	klog.InfoS("Downloaded image", "destination", dest, "url", i.Url, "sizeMB", n/1_048_576, "duration", time.Since(start))

	return nil
}

func fileExistsInDir(dir, filename string) (bool, error) {
	entries, err := os.ReadDir(dir)
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
