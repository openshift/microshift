package assets

import (
	"context"
	"testing"
)

func TestApply(t *testing.T) {

	err := ApplyGeneric(context.Background(),
		[]string{
			"components/lvms/lvms_default-lvmcluster.yaml",
		},
		nil,
		map[string]interface{}{},
		nil,
		"/home/jmoller/Projects/microshift/kubeconfig")

	if err != nil {
		t.Errorf("Failed to apply lvmcluster: %v", err)
	}
}
