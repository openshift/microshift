package mocks

import (
	"github.com/openshift/microshift/test/pkg/util"
	"github.com/stretchr/testify/mock"
)

var _ util.EventManager = (*EventManagerMock)(nil)

type EventManagerMock struct {
	mock.Mock
}

func (em *EventManagerMock) AddEvent(e util.IEvent) {
	em.Called(e)
}

func (em *EventManagerMock) WriteToFiles(intervalsFile string, timelinesFile string) error {
	panic("unimplemented")
}

func (em *EventManagerMock) GetJUnit() *util.JUnitTestSuites {
	panic("unimplemented")
}
