package bootstrap

import (
	"context"

	"github.com/openshift/microshift/pkg/config"
)

type Token struct {
}

type TokenManager struct {
}

func NewTokenManager(cfg *config.MicroshiftConfig) *TokenManager {
	return &TokenManager{}
}

func (s *TokenManager) Name() string           { return "token-manager" }
func (s *TokenManager) Dependencies() []string { return []string{"kube-apiserver"} }

func Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)
	defer close(ready)

}
