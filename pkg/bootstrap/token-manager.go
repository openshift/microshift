package bootstrap

import (
	"context"
	"path/filepath"

	"github.com/openshift/microshift/pkg/config"
)

type Token struct {
}

type TokenManager struct {
	path string
}

func NewTokenManager(cfg *config.MicroshiftConfig) *TokenManager {
	return &TokenManager{
		path: filepath.Join(cfg.DataDir, "resources", "microshift-bootstrap-token"),
	}
}

func (s *TokenManager) Name() string           { return "token-manager" }
func (s *TokenManager) Dependencies() []string { return []string{} }

func (s *TokenManager) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)
	defer close(ready)

	CreateTokenFile(s.path)
	return nil
}
