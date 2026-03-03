package schemapatch

import (
	"fmt"

	"sigs.k8s.io/yaml"
)

// formatData formats the given YAML data.
// We use indentation of 2 spaces, and the yaml.v3 library as this is what
// other kube generators use.
func formatData(in []byte) ([]byte, error) {
	node := make(map[string]interface{})
	if err := yaml.Unmarshal(in, &node); err != nil {
		return nil, fmt.Errorf("could not unmarshal YAML: %v", err)
	}

	data, err := yaml.Marshal(node)
	if err != nil {
		return nil, fmt.Errorf("could not encode YAML: %v", err)
	}

	return data, nil
}
