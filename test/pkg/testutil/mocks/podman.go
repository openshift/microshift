package mocks

import (
	"context"

	"github.com/openshift/microshift/test/pkg/compose/helpers"
	"github.com/stretchr/testify/mock"
)

var _ helpers.Podman = (*PodmanMock)(nil)

type PodmanMock struct {
	mock.Mock
}

func (p *PodmanMock) BuildAndSave(ctx context.Context, tag string, containerfilePath string, contextDir string, output string) error {
	args := p.Called(ctx, tag, containerfilePath, contextDir, output)
	return args.Error(0)
}
