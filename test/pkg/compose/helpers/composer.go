package helpers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/openshift/microshift/test/pkg/util"
	"github.com/osbuild/weldr-client/v2/weldr"
	"k8s.io/klog/v2"
)

type Composer interface {
	ListSources() ([]string, error)
	DeleteSource(id string) error
	AddSource(toml string) error
}

var _ Composer = (*composer)(nil)

type composer struct {
	client        weldr.Client
	ostreeRepoURL string

	logsDir   string
	buildsDir string
}

func NewComposer(paths *util.Paths, ostreeRepoURL string) (Composer, error) {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(ostreeRepoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %q - is webserver up?", ostreeRepoURL)
	}
	resp.Body.Close()

	return &composer{
		client:        weldr.InitClientUnixSocket(context.Background(), 1, "/run/weldr/api.socket"),
		ostreeRepoURL: ostreeRepoURL,

		logsDir:   paths.BuildLogsDir,
		buildsDir: paths.BuildsDir,
	}, nil
}

func (c *composer) ListSources() ([]string, error) {
	klog.InfoS("Listing Composer Sources")
	sources, apiResponse, err := c.client.ListSources()
	if err != nil {
		return nil, fmt.Errorf("listing composer sources failed: %w", err)
	}
	if apiResponse != nil && !apiResponse.Status {
		return nil, fmt.Errorf("ListSources() - wrong api response: %+v", apiResponse)
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
		return fmt.Errorf("DeleteSource(%q) - wrong api response: %+v", id, apiResponse)
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
		return fmt.Errorf("NewSourceTOML(...) - wrong api response: %+v", apiResponse)
	}
	klog.InfoS("Added Composer Source", "toml", short)
	return nil
}
