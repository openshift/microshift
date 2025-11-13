package controllers

import (
	"context"
	"testing"

	kubecontrolplanev1 "github.com/openshift/api/kubecontrolplane/v1"
	"github.com/openshift/microshift/pkg/config"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/yaml"
)

func TestKubeAPIServer_configureFeatureGates(t *testing.T) {

	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		ccfg *config.Config
		// Named input parameters for target function.
		cfg     *config.Config
		wantErr bool
	}{
		{
			name: "custom feature gates passed to kube-apiserver",
			ccfg: config.NewDefault(),
			cfg: &config.Config{
				ApiServer: config.ApiServer{
					AuditLog: config.AuditLog{
						Profile: "Default",
					},
					FeatureGates: config.FeatureGates{
						FeatureSet: config.FeatureSetCustomNoUpgrade,
						CustomNoUpgrade: config.CustomNoUpgrade{
							Enabled:  []string{"TestCustomFeatureGate"},
							Disabled: []string{"TestCustomFeatureGate2"},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewKubeAPIServer(tt.ccfg)
			gotErr := s.configure(context.Background(), tt.cfg)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("configure() failed: %v", gotErr)
				}
				return
			}

			runtimeKasConfig := &kubecontrolplanev1.KubeAPIServerConfig{}
			err := yaml.Unmarshal(s.kasConfigBytes, runtimeKasConfig)
			if err != nil {
				t.Errorf("yaml.Unmarshal() failed: %v", err)
			}
			if !sets.NewString(runtimeKasConfig.APIServerArguments["feature-gates"]...).HasAll("TestCustomFeatureGate", "TestCustomFeatureGate2") {
				t.Errorf("expected TestCustomFeatureGate=true, TestCustomFeatureGate2=false, got %s", runtimeKasConfig.APIServerArguments["feature-gates"])
			}
		})
	}
}
