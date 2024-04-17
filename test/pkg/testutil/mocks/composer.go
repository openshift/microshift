package mocks

import (
	"context"
	"time"

	"github.com/openshift/microshift/test/pkg/compose/helpers"
	"github.com/stretchr/testify/mock"
)

var _ helpers.Composer = (*ComposerMock)(nil)

type ComposerMock struct {
	mock.Mock
}

func (c *ComposerMock) AddBlueprint(toml string) error {
	args := c.Called(toml)
	return args.Error(0)
}

func (c *ComposerMock) AddSource(toml string) error {
	args := c.Called(toml)
	return args.Error(0)
}

func (c *ComposerMock) DeleteSource(id string) error {
	args := c.Called(id)
	return args.Error(0)
}

func (c *ComposerMock) DepsolveBlueprint(name string) error {
	args := c.Called(name)
	return args.Error(0)
}

func (c *ComposerMock) ListSources() ([]string, error) {
	args := c.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (c *ComposerMock) SaveComposeImage(id string, friendlyName string, ext string) (string, error) {
	args := c.Called(id, friendlyName, ext)
	return args.String(0), args.Error(1)
}

func (c *ComposerMock) SaveComposeLogs(id string, friendlyName string) error {
	args := c.Called(id, friendlyName)
	return args.Error(0)
}

func (c *ComposerMock) SaveComposeMetadata(id string, friendlyName string) error {
	args := c.Called(id, friendlyName)
	return args.Error(0)
}

func (c *ComposerMock) StartCompose(blueprint string, composeType string) (string, error) {
	args := c.Called(blueprint, composeType)
	return args.String(0), args.Error(1)
}

func (c *ComposerMock) StartOSTreeCompose(blueprint string, composeType string, ref string, parent string) (string, error) {
	args := c.Called(blueprint, composeType, ref, parent)
	return args.String(0), args.Error(1)
}

func (c *ComposerMock) WaitForCompose(ctx context.Context, id string, friendlyName string, timeout time.Duration) error {
	args := c.Called(ctx, id, friendlyName, timeout)
	return args.Error(0)
}
