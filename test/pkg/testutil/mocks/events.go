package mocks

import (
	"github.com/openshift/microshift/test/pkg/testutil"
	"github.com/stretchr/testify/mock"
)

var _ testutil.EventManager = (*EventManagerMock)(nil)

type EventManagerMock struct {
	mock.Mock
}

func (em *EventManagerMock) AddEvent(e testutil.IEvent) {
	em.Called(e)
}

func (em *EventManagerMock) WriteToFiles(intervalsFile string, timelinesFile string) error {
	panic("unimplemented")
}

func (em *EventManagerMock) GetJUnit() *testutil.JUnitTestSuites {
	panic("unimplemented")
}
