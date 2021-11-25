package bootstrap

import (
	"context"
	"os"
	"path/filepath"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
)

type TokenManager struct {
	path string
	cfg  *config.MicroshiftConfig
}

func NewTokenManager(cfg *config.MicroshiftConfig) *TokenManager {
	return &TokenManager{
		path: filepath.Join(cfg.DataDir, "resources", "microshift-bootstrap-token"),
		cfg:  cfg,
	}
}

func (s *TokenManager) Name() string           { return "token-manager" }
func (s *TokenManager) Dependencies() []string { return []string{} }

func (s *TokenManager) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)
	defer close(ready)

	CreateTokenFile(s.path)
	_, err := os.Stat(s.cfg.DataDir + "/resources/kubelet/bootstrap-kubeconfig")
	if os.IsNotExist(err) {
		if err := util.BootstrapKubeconfig(GetToken(s.path), s.cfg.DataDir+"/resources/kubelet/bootstrap-kubeconfig", "system:bootstrappers", []string{"system:bootstrappers"}, s.cfg.Cluster.URL); err != nil {
			return err
		}
	}

	return nil
}
