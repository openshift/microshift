package mocks

import (
	"github.com/openshift/microshift/test/pkg/compose/helpers"
	"github.com/stretchr/testify/mock"
)

var _ helpers.Ostree = (*OstreeMock)(nil)

type OstreeMock struct {
	mock.Mock
}

func (o *OstreeMock) CreateAlias(origin string, aliases ...string) error {
	args := o.Called(origin, aliases)
	return args.Error(0)
}

func (o *OstreeMock) DoesRefExists(ref string) (bool, error) {
	args := o.Called(ref)
	return args.Bool(0), args.Error(1)
}

func (o *OstreeMock) ExtractCommit(path string) error {
	args := o.Called(path)
	return args.Error(0)
}
