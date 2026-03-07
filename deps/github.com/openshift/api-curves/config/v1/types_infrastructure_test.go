package v1

import (
	"os"
	"path"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"sigs.k8s.io/yaml"
)

const (
	infraCRDDefaultFilePath     = "0000_10_config-operator_01_infrastructures-Default.crd.yaml"
	infraCRDTestPreviewFilePath = "0000_10_config-operator_01_infrastructures-TechPreviewNoUpgrade.crd.yaml"
)

// TestInfrastructureStatusDefault verifies that the infrastructure CR status does not have default value
// The admission code under https://github.com/openshift/kubernetes/pull/877 is expecting that the infrastructure status
// field will not have a default value.
// It allows separating between clean installation and the roll-back to the previous version of the cluster
func TestInfrastructureStatusDefault(t *testing.T) {

	filePaths := []string{infraCRDDefaultFilePath, infraCRDTestPreviewFilePath}

	for _, filepath := range filePaths {
		infraCRDBytes, err := os.ReadFile(path.Join("zz_generated.crd-manifests", filepath))
		if err != nil {
			t.Fatalf("failed to read infrastructure CRD file %q: %v", filepath, err)
		}

		var infraCRD map[string]interface{}
		if err := yaml.Unmarshal(infraCRDBytes, &infraCRD); err != nil {
			t.Fatalf("failed to unmarshal infra CRD: %v", err)
		}
		infraCRDSpec := infraCRD["spec"].(map[string]interface{})
		infraCRDVersions := infraCRDSpec["versions"].([]interface{})
		for _, v := range infraCRDVersions {
			infraCRDVersion := v.(map[string]interface{})
			status, exists, err := unstructured.NestedMap(infraCRDVersion, "schema", "openAPIV3Schema", "properties", "status")
			if err != nil {
				t.Fatalf("failed to get nested map: %v", err)
			}

			if !exists {
				t.Fatalf("one of fields does not exist under the CRD")
			}

			if _, ok := status["default"]; ok {
				t.Fatalf("expected no default for the infrastructure CRD status")
			}
		}
	}
}
