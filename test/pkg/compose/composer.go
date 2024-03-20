package compose

import (
	"archive/tar"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/osbuild/weldr-client/v2/weldr"
	"k8s.io/klog/v2"
)

type Composer interface {
	ListSources() ([]string, error)
	DeleteSource(id string) error
	AddSource(toml string) error

	AddBlueprint(toml string) error
	DepsolveBlueprint(name string) error

	StartOSTreeCompose(blueprint, composeType, ref, parent, url string, size uint) (string, error)
	StartCompose(blueprint, composeType string, size uint) (string, error)

	WaitForCompose(id, friendlyName string, timeout time.Duration) error

	SaveComposeLogs(id, friendlyName string) error
	SaveComposeMetadata(id, friendlyName string) error
	SaveComposeImage(id, friendlyName, ext string) (string, error)
}

type composer struct {
	client       weldr.Client
	artifactsDir string
}

func NewComposer(testDirPath string) Composer {
	return &composer{
		client:       weldr.InitClientUnixSocket(context.Background(), 1, "/run/weldr/api.socket"),
		artifactsDir: filepath.Join(testDirPath, "..", "_output", "test-images"),
	}
}

func (c *composer) ListSources() ([]string, error) {
	klog.InfoS("Listing Composer Sources")
	sources, apiResponse, err := c.client.ListSources()
	if err != nil {
		return nil, fmt.Errorf("listing composer sources failed: %w", err)
	}
	if apiResponse != nil && !apiResponse.Status {
		return nil, fmt.Errorf("listing composer sources returned wrong response: %v", apiResponse)
	}
	klog.InfoS("Listed Composer Sources", "sources", sources)
	return sources, nil
}

func (c *composer) DeleteSource(id string) error {
	klog.InfoS("Deleting Composer Source", "id", id)
	apiResponse, err := c.client.DeleteSource(id)
	if err != nil {
		return fmt.Errorf("deleting composer source failed: %w", err)
	}
	if apiResponse != nil && !apiResponse.Status {
		return fmt.Errorf("deleting composer source returned wrong response: %v", apiResponse)
	}
	klog.InfoS("Deleted Composer Source", "id", id)
	return nil
}

func (c *composer) AddSource(toml string) error {
	short := strings.ReplaceAll(toml[:50], "\n", "") + "..."
	klog.InfoS("Adding Composer Source", "toml", short)
	apiResponse, err := c.client.NewSourceTOML(toml)
	if err != nil {
		return fmt.Errorf("adding composer source failed: %w", err)
	}
	if apiResponse != nil && !apiResponse.Status {
		return fmt.Errorf("adding composer source returned wrong response: %v", apiResponse)
	}
	klog.InfoS("Added Composer Source", "toml", short)
	return nil
}

func (c *composer) AddBlueprint(toml string) error {
	short := strings.ReplaceAll(toml[:50], "\n", "") + "..."
	klog.InfoS("Adding Composer Blueprint", "toml", short)
	apiResponse, err := c.client.PushBlueprintTOML(toml)
	if err != nil {
		return fmt.Errorf("adding composer blueprint failed: %w", err)
	}
	if apiResponse != nil && !apiResponse.Status {
		return fmt.Errorf("adding composer blueprint returned wrong response: %v", apiResponse)
	}
	klog.InfoS("Added Composer Blueprint", "toml", short)
	return nil
}

func (c *composer) DepsolveBlueprint(name string) error {
	klog.InfoS("Depsolving blueprint", "name", name)
	blueprints, apiErrors, err := c.client.DepsolveBlueprints([]string{name})

	// TODO: Write to file
	_ = blueprints

	if err != nil {
		return fmt.Errorf("error depsolving blueprint %q: %w", name, err)
	}
	if len(apiErrors) != 0 {
		return fmt.Errorf("unsuccessful blueprint depsolve: %+v", apiErrors)
	}

	klog.InfoS("Depsolved blueprint", "name", name)

	return nil
}

func (c *composer) StartOSTreeCompose(blueprint, composeType, ref, parent, url string, size uint) (string, error) {
	klog.InfoS("Starting ostree compose", "blueprint", blueprint, "type", composeType, "ref", ref, "parent", parent, "url", url)
	id, apiResponse, err := c.client.StartOSTreeCompose(blueprint, composeType, ref, parent, url, size)
	if err != nil {
		return "", fmt.Errorf("error starting ostree compose: %w", err)
	}
	if apiResponse != nil && !apiResponse.Status {
		return "", fmt.Errorf("unsuccessful compose start: %+v", apiResponse)
	}
	klog.InfoS("Started ostree compose", "blueprint", blueprint, "id", id)

	return id, nil
}

func (c *composer) StartCompose(blueprint, composeType string, size uint) (string, error) {
	klog.InfoS("Starting compose", "blueprint", blueprint, "type", composeType)
	id, apiResponse, err := c.client.StartCompose(blueprint, composeType, size)
	if err != nil {
		return "", fmt.Errorf("error starting ostree compose: %w", err)
	}
	if apiResponse != nil && !apiResponse.Status {
		return "", fmt.Errorf("unsuccessful compose start: %+v", apiResponse)
	}
	klog.InfoS("Started compose", "blueprint", blueprint, "id", id)
	return id, nil
}

func (c *composer) WaitForCompose(id, friendlyName string, timeout time.Duration) error {
	klog.InfoS("Waiting for compose", "id", id, "timeout", timeout)

	aborted, info, apiResponse, err := c.client.ComposeWait(id, timeout, 30*time.Second)
	klog.InfoS("Wait for compose complete", "id", id)
	_ = info

	// info should always be set, even if compose failed
	infoJson, infoErr := json.MarshalIndent(info, "", "    ")
	if infoErr != nil {
		return fmt.Errorf("failed to marshal compose info: %w", err)
	}
	infoFilepath := filepath.Join(c.artifactsDir, "build-logs", friendlyName+"_info.log")
	infoErr = os.WriteFile(infoFilepath, infoJson, 0644)
	if infoErr != nil {
		return fmt.Errorf("failed to write compose info to file: %w", err)
	}

	if err != nil {
		return fmt.Errorf("failed to wait for the compose %q: %w", id, err)
	}
	if apiResponse != nil && !apiResponse.Status {
		return fmt.Errorf("unsuccessful compose wait: %+v", apiResponse)
	}
	if aborted {
		return fmt.Errorf("wait for compose %q timed out", id)
	}
	return nil
}

func (c *composer) SaveComposeLogs(id, friendlyName string) error {
	klog.InfoS("Saving compose logs archive", "id", id, "friendlyName", friendlyName)

	archiveFilepath := filepath.Join(c.artifactsDir, "build-logs", friendlyName+".tar")
	logFilepath := filepath.Join(c.artifactsDir, "build-logs", friendlyName+".log")

	err := os.RemoveAll(archiveFilepath)
	if err != nil {
		return fmt.Errorf("failed to remove existing %q before downloading it: %w", archiveFilepath, err)
	}

	filename, apiResponse, err := c.client.ComposeLogsPath(id, archiveFilepath)
	if err != nil {
		return fmt.Errorf("failed to get compose's %q logs: %w", id, err)
	}
	if apiResponse != nil && !apiResponse.Status {
		return fmt.Errorf("unsuccessful get compose's %q logs: %w", id, err)
	}

	klog.InfoS("Saved compose logs archive", "id", id, "path", filename)

	err = extractSingleFileFromTar(archiveFilepath, logFilepath)
	if err != nil {
		return err
	}

	return nil
}

func (c *composer) SaveComposeMetadata(id, friendlyName string) error {
	klog.InfoS("Getting compose metadata", "id", id, "friendlyName", friendlyName)

	archiveFilepath := filepath.Join(c.artifactsDir, "build-logs", friendlyName+"_metadata.tar")
	logFilepath := filepath.Join(c.artifactsDir, "build-logs", friendlyName+"_metadata.log")

	err := os.RemoveAll(archiveFilepath)
	if err != nil {
		return fmt.Errorf("failed to remove existing %q before downloading it: %w", archiveFilepath, err)
	}

	filename, apiResponse, err := c.client.ComposeMetadataPath(id, archiveFilepath)
	if err != nil {
		return fmt.Errorf("failed to get compose's %q metadata: %w", id, err)
	}
	if apiResponse != nil && !apiResponse.Status {
		return fmt.Errorf("unsuccessful get compose's %q metadata: %w", id, err)
	}

	klog.InfoS("Got compose metadata", "id", id, "filename", filename)

	err = extractSingleFileFromTar(archiveFilepath, logFilepath)
	if err != nil {
		return err
	}

	return nil
}

func (c *composer) SaveComposeImage(id, friendlyName, ext string) (string, error) {
	klog.InfoS("Getting compose image", "id", id, "friendlyName", friendlyName, "ext", ext)

	path := filepath.Join(c.artifactsDir, "builds", friendlyName+ext)
	err := os.RemoveAll(path)
	if err != nil {
		return "", fmt.Errorf("failed to remove existing %q before downloading it: %w", path, err)
	}

	filename, apiResponse, err := c.client.ComposeImagePath(id, path)
	if err != nil {
		return "", fmt.Errorf("failed to get compose's %q image: %w", id, err)
	}
	if apiResponse != nil && !apiResponse.Status {
		return "", fmt.Errorf("unsuccessful get compose's %q image: %w", id, err)
	}

	klog.InfoS("Got compose image", "id", id, "filename", filename)
	return filename, nil
}

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
