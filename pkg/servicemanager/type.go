package servicemanager

import (
	"context"
)

type Runner interface {
	Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error
}

type Service interface {
	Name() string
	Dependencies() []string
	Runner
}
