package servicemanager

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/openshift/microshift/pkg/util/sigchannel"
)

type serviceTest struct {
	service Service
	out     error
}
type serviceGraphTest []serviceTest

func TestAddService(t *testing.T) {
	var tests = []serviceGraphTest{
		{
			serviceTest{service: nil, out: errors.New("service must not be <nil>")},
		},
		{
			serviceTest{service: NewGenericService("foo", nil, nil), out: nil},
		},
		{
			serviceTest{service: NewGenericService("foo", nil, nil), out: nil},
			serviceTest{service: NewGenericService("bar", []string{"foo"}, nil), out: nil},
		},
		{
			serviceTest{service: NewGenericService("foo", nil, nil), out: nil},
			serviceTest{service: NewGenericService("foo", nil, nil), out: errors.New("service 'foo' added more than once")},
		},
		{
			serviceTest{service: NewGenericService("bar", []string{"foo"}, nil), out: errors.New("dependecy 'foo' of service 'bar' not yet defined")},
			serviceTest{service: NewGenericService("foo", nil, nil), out: nil},
		},
	}

	for _, test := range tests {
		m := NewServiceManager()
		for _, servicetest := range test {
			got := "<nil>"
			if err := m.AddService(servicetest.service); err != nil {
				got = err.Error()
			}
			want := "<nil>"
			if err := servicetest.out; err != nil {
				want = err.Error()
			}
			if want != got {
				t.Errorf("got: %v; want: %v", got, want)
			}
		}
	}
}

func TestRunToCompletion(t *testing.T) {
	var wg sync.WaitGroup
	defer wg.Wait()

	var runToCompletionFunc = func(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
		defer close(stopped)

		<-time.After(time.Second)
		close(ready)
		<-time.After(time.Second)
		wg.Done()

		return nil
	}

	m := NewServiceManager()
	m.AddService(NewGenericService("foo", nil, runToCompletionFunc))
	m.AddService(NewGenericService("bar", []string{"foo"}, runToCompletionFunc))
	wg.Add(2)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ready, stopped := make(chan struct{}), make(chan struct{})
	if err := m.Run(ctx, ready, stopped); err != nil {
		t.Errorf("error running %s: %v", m.Name(), err)
	}
	if !sigchannel.IsClosed(ready) {
		t.Errorf("ready channel not closed after completing service manager")
	}
	if !sigchannel.IsClosed(stopped) {
		t.Errorf("stopped channel not closed after completing service manager")
	}
}

func TestRunCancellation(t *testing.T) {
	var wg sync.WaitGroup
	defer wg.Wait()

	var runToCompletionFunc = func(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
		defer close(stopped)

		<-time.After(time.Second)
		close(ready)
		<-ctx.Done()
		wg.Done()

		return nil
	}

	m := NewServiceManager()
	m.AddService(NewGenericService("foo", nil, runToCompletionFunc))
	m.AddService(NewGenericService("bar", []string{"foo"}, runToCompletionFunc))
	wg.Add(2)

	ctx, cancel := context.WithCancel(context.Background())
	ready, stopped := make(chan struct{}), make(chan struct{})
	go func() {
		m.Run(ctx, ready, stopped)
	}()

	select {
	case <-ready:
	case <-time.After(time.Second * 5):
		t.Fatalf("timeout waiting for %s to become ready", m.Name())
	}

	cancel()
}
