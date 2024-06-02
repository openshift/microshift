package helpers

import (
	"strings"

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
