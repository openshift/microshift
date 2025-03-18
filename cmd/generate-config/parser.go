package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
	v1ext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	kubeYaml "k8s.io/apimachinery/pkg/util/yaml"
)

const (
	jsonTypeString  = "string"
	jsonTypeNumber  = "number"
	jsonTypeInteger = "integer"
	jsonTypeObject  = "object"
	jsonTypeArray   = "array"
)

type crdParser struct {
	NoComments bool
	NoDefaults bool
}

func (p crdParser) parseToJsonSchema(data []byte) (v1ext.JSONSchemaProps, error) {
	crd := v1ext.CustomResourceDefinition{}
	err := kubeYaml.Unmarshal(data, &crd)
	if err != nil {
		return v1ext.JSONSchemaProps{}, fmt.Errorf("failed to unmarshal custom resource config: %w", err)
	}

	if len(crd.Spec.Versions) != 1 {
		return v1ext.JSONSchemaProps{}, fmt.Errorf("expected length of crd.spec.versions to be 1 but got %d", len(crd.Spec.Versions))
	}

	configData := crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["config"]
	return configData, nil
}

func (p crdParser) parseToJsonOpenAPI(data []byte) ([]byte, error) {
	configData, err := p.parseToJsonSchema(data)
	if err != nil {
		return nil, err
	}

	openAPIJsonData, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal openapi into json: %w", err)
	}
	return openAPIJsonData, nil
}

func (p crdParser) parseToYamlNode(data []byte) (*yaml.Node, error) {
	configData, err := p.parseToJsonSchema(data)
	if err != nil {
		return nil, err
	}

	node := p.toYamlNodeObject(configData.Properties)
	return node, nil
}

func (p crdParser) toYamlNodeObject(val map[string]v1ext.JSONSchemaProps) *yaml.Node {
	node := &yaml.Node{
		Kind: yaml.MappingNode,
	}

	orderedKeyNameArray := schemaKeyToOrderedArray(val)

	for _, schemaKeyName := range orderedKeyNameArray {
		field, ok := val[schemaKeyName]
		if !ok {
			// This should never happen since the ordered key array is created from the keys in val.
			// This would definitely mean it's time to panic.
			panic(fmt.Errorf("failed to find %s in the map of JSONSchemaProps: \nData ===\n%+v\n==", schemaKeyName, val))
		}

		keyNode := &yaml.Node{
			Value: schemaKeyName,
			Kind:  yaml.ScalarNode,
		}

		if !p.NoComments {
			keyNode.HeadComment = strings.ReplaceAll(field.Description, "\n\n\n", "\n#\n")
		}

		var valueNode *yaml.Node
		switch field.Type {
		case jsonTypeArray:

			valueNode = p.toYamlNodeArray(field.Items)
			if nodes := parseArrayJSONValue(field.Default); nodes != nil && !p.NoDefaults {
				valueNode.Content = nodes
			}

			if exampleValue := parseArrayJSONExample(field.Example); exampleValue != "" && !p.NoComments {
				keyNode.HeadComment = fmt.Sprintf("%s\nexample:\n  %s", keyNode.HeadComment, exampleValue)
			}

		case jsonTypeObject:

			valueNode = p.toYamlNodeObject(field.Properties)

			if exampleValue := parseMapJSONExample(field.Example); exampleValue != "" && !p.NoComments {
				keyNode.HeadComment = fmt.Sprintf("%s\nexample:\n  %s", keyNode.HeadComment, exampleValue)
			}

		default:

			valueNode = p.toYamlNodeValue(field)

			if exampleValue := parseScalarJSONExample(field.Example); exampleValue != "" && !p.NoComments {
				keyNode.HeadComment = fmt.Sprintf("%s\nexample:\n  %s", keyNode.HeadComment, exampleValue)
			}
		}
		node.Content = append(node.Content, keyNode, valueNode)
	}
	return node
}

func (p crdParser) toYamlNodeArray(val *v1ext.JSONSchemaPropsOrArray) *yaml.Node {
	node := &yaml.Node{
		Kind: yaml.SequenceNode,
	}
	if val == nil {
		return node
	}

	if val.Schema != nil {
		var valueNode *yaml.Node
		switch val.Schema.Type {
		case jsonTypeObject:
			valueNode = p.toYamlNodeObject(val.Schema.Properties)
		case jsonTypeArray:
			valueNode = p.toYamlNodeArray(val.Schema.Items)
		default:
			// No default to avoid arrays ending up [""].
			// Instead, they'll appear as [].
		}
		if valueNode != nil {
			node.Content = append(node.Content, valueNode)
		}

		return node
	}

	for _, field := range val.JSONSchemas {
		var valueNode *yaml.Node
		switch field.Type {
		case jsonTypeObject:
			valueNode = p.toYamlNodeObject(field.Properties)
		case jsonTypeArray:
			valueNode = p.toYamlNodeArray(field.Items)
		default:
			valueNode = p.toYamlNodeValue(field)
		}
		node.Content = append(node.Content, valueNode)
	}

	return node
}

func (p crdParser) toYamlNodeValue(val v1ext.JSONSchemaProps) *yaml.Node {
	node := &yaml.Node{
		Kind: yaml.ScalarNode,
	}

	if !p.NoDefaults {
		node.Value = parseScalarJSONValue(val.Default)
	}

	if node.Value == "" {
		switch val.Type {
		case jsonTypeString:
			node.SetString("")
		case jsonTypeInteger:
			node.Value = "0"
		case jsonTypeNumber:
			node.Value = "0.0"
		}
	}

	return node
}
