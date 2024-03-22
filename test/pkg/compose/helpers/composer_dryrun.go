package helpers

import (
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

func (c *dryrunComposer) StartOSTreeCompose(blueprint, composeType, ref, parent string) (string, error) {
	klog.InfoS("DRYRUN: Starting ostree compose", "blueprint", blueprint, "type", composeType, "ref", ref, "parent", parent)
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

func (c *dryrunComposer) StartCompose(blueprint string, composeType string) (string, error) {
	klog.InfoS("DRYRUN: Starting compose", "blueprint", blueprint, "type", composeType)
	return "dummy-dry-run-id", nil
}

func (c *dryrunComposer) WaitForCompose(id, friendlyName string, timeout time.Duration) error {
	klog.InfoS("DRYRUN: Waiting for compose", "id", id, "timeout", timeout)
	time.Sleep(5 * time.Second)
	klog.InfoS("DRYRUN: Waited for compose", "id", id, "timeout", timeout)
	return nil
}
