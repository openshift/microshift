package compose

import (
	"context"
	"fmt"
	"strings"

	"github.com/osbuild/weldr-client/v2/weldr"
	"k8s.io/klog/v2"
)

type Composer interface {
	ListSources() ([]string, error)
	DeleteSource(id string) error
	AddSource(toml string) error
}

type composer struct {
	client weldr.Client
}

func NewComposer() Composer {
	return &composer{
		client: weldr.InitClientUnixSocket(context.Background(), 1, "/run/weldr/api.socket"),
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
