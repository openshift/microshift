package mocks

import (
	"github.com/openshift/microshift/test/pkg/compose/helpers"
	"github.com/stretchr/testify/mock"
)

var _ helpers.Composer = (*ComposerMock)(nil)

type ComposerMock struct {
	mock.Mock
}

func (c *ComposerMock) AddSource(toml string) error {
	args := c.Called(toml)
	return args.Error(0)
}

func (c *ComposerMock) DeleteSource(id string) error {
	args := c.Called(id)
	return args.Error(0)
}

func (c *ComposerMock) ListSources() ([]string, error) {
	args := c.Called()
	return args.Get(0).([]string), args.Error(1)
}
