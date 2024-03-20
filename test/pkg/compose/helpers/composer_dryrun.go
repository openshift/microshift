package helpers

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"k8s.io/klog/v2"
)

var _ Composer = (*dryrunComposer)(nil)

type dryrunComposer struct{}

func NewDryRunComposer() Composer {
	return &dryrunComposer{}
}

func (c *dryrunComposer) ListSources() ([]string, error) {
	sources := []string{}
	klog.InfoS("DRYRUN: Listing Sources", "sources", sources)
	return sources, nil
}

func (c *dryrunComposer) DeleteSource(id string) error {
	klog.InfoS("DRYRUN: Removing Source", "id", id)
	return nil
}

func (c *dryrunComposer) AddSource(toml string) error {
	short := strings.ReplaceAll(toml[:50], "\n", "") + "..."
	klog.InfoS("DRYRUN: Adding Source", "toml", short)
	return nil
}

func (c *dryrunComposer) AddBlueprint(toml string) error {
	short := strings.ReplaceAll(toml[:50], "\n", "") + "..."
	klog.InfoS("DRYRUN: Adding Blueprint", "toml", short)
	return nil
}

func (c *dryrunComposer) DepsolveBlueprint(name string) error {
	klog.InfoS("DRYRUN: Depsolving Blueprint", "name", name)
	return nil
}

func (c *dryrunComposer) StartOSTreeCompose(blueprint, composeType, ref, parent, url string, size uint) (string, error) {
	klog.InfoS("DRYRUN: Starting ostree compose", "blueprint", blueprint, "type", composeType, "ref", ref, "parent", parent, "url", url)
	return "dummy-dry-run-id", nil
}

func (c *dryrunComposer) SaveComposeImage(id, friendlyName, ext string) (string, error) {
	klog.InfoS("DRYRUN: Getting compose image", "id", id, "friendlyName", friendlyName)
	return "/dummy/dry/run/path", nil
}

func (c *dryrunComposer) SaveComposeLogs(id string, path string) error {
	klog.InfoS("DRYRUN: Saving compose logs", "id", id, "path", path)
	return nil
}

func (c *dryrunComposer) SaveComposeMetadata(id string, friendlyName string) error {
	klog.InfoS("DRYRUN: Getting compose metadata", "id", id, "friendlyName", friendlyName)
	return nil
}

func (c *dryrunComposer) StartCompose(blueprint string, composeType string, size uint) (string, error) {
	klog.InfoS("DRYRUN: Starting compose", "blueprint", blueprint, "type", composeType)
	return "dummy-dry-run-id", nil
}

func (c *dryrunComposer) WaitForCompose(id, friendlyName string, timeout time.Duration) error {
	klog.InfoS("DRYRUN: Waiting for compose", "id", id, "timeout", timeout)
	time.Sleep(1 * time.Second)
	klog.InfoS("DRYRUN: Waited for compose", "id", id, "timeout", timeout)
	return nil
}

// extractSingleFileFromTar extracts exactly one file from tar archive.
// Intended to be used with compose metadata or logs as the archives contain only single file.
func extractSingleFileFromTar(archivePath, filePath string) error {
	archiveFile, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open tar file %q: %w", archivePath, err)
	}
	tarFile := tar.NewReader(archiveFile)
	header, err := tarFile.Next()
	if err == io.EOF {
		return fmt.Errorf("tar file %q is empty, expected single file", archivePath)
	}
	if err != nil {
		return fmt.Errorf("error when reading tar file %q: %w", archivePath, err)
	}
	if header.Typeflag != tar.TypeReg {
		return fmt.Errorf("unexpected header type in tar file %q - type: %v", archivePath, header.Typeflag)
	}

	extractedFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
	if err != nil {
		return fmt.Errorf("failed to open file %q: %w", filePath, err)
	}
	defer extractedFile.Close()

	written, err := io.Copy(extractedFile, tarFile)
	if err != nil {
		return fmt.Errorf("failed to copy file to %q: %w", filePath, err)
	}

	if header.Size != written {
		return fmt.Errorf("incomplete file copy from the tar archive, size:%d, copied:%d", header.Size, written)
	}

	err = os.Remove(archivePath)
	if err != nil {
		return fmt.Errorf("failed to remove old tar file %q: %w", archivePath, err)
	}

	return nil
}
