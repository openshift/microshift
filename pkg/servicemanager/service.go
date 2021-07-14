package servicemanager

import (
	"context"
	"fmt"
)

type RunFunc func(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error

type GenericService struct {
	name string
	deps []string

	run RunFunc
}

func NewGenericService(name string, dependencies []string, run RunFunc) *GenericService {
	return &GenericService{
		name: name,
		deps: dependencies,
		run:  run,
	}
}
func (s *GenericService) Name() string           { return s.name }
func (s *GenericService) Dependencies() []string { return s.deps }

func (s *GenericService) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	if s.run == nil {
		defer close(stopped)
		defer close(ready)
		return fmt.Errorf("no run function defined for '%s'", s.Name())
	}

	return s.run(ctx, ready, stopped)
}
