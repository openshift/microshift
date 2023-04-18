package main

import (
	"encoding/json"
	"sort"
	"strconv"

	"gopkg.in/yaml.v3"
	v1ext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

var (
	defaultTemplateFuncs = map[string]any{
		"parseToConfigYaml":     parseToConfigYaml,
		"parseToConfigYamlOpts": parseToConfigYamlOpts,
		// this is an always false flag to give explicit intent
		// to blocks such as {{ with deleteCurrent }}.
		"deleteCurrent": func() bool { return false },
	}
)

func parseToConfigYaml(data []byte) string {
	parser := crdParser{}
	yamlNode, err := parser.parseToYamlNode(data)
	if err != nil {
		panic(err)
	}
	return toYaml(yamlNode)
}

func parseToConfigYamlOpts(data []byte, noComments, noDefaults bool) string {
	parser := crdParser{NoComments: noComments, NoDefaults: noDefaults}
	yamlNode, err := parser.parseToYamlNode(data)
	if err != nil {
		panic(err)
	}
	return toYaml(yamlNode)
}

func toYaml(val interface{}) string {
	result, err := yaml.Marshal(val)
	if err != nil {
		panic(err)
	}
	return string(result)
}

// These util functions are here to primarily deal with the default and example options,
// they take the raw json and create a yaml node out of them.

// Functions to help spit out string versions of examples to be used in comments
func parseArrayJSONExample(rawJson *v1ext.JSON) string {
	results := parseArrayJSONValue(rawJson)
	if len(results) == 0 {
		return ""
	}
	return toYaml(results)
}
func parseMapJSONExample(rawJson *v1ext.JSON) string {
	results := parseMapJSONValue(rawJson)
	if len(results) == 0 {
		return ""
	}
	return toYaml(results)
}
func parseScalarJSONExample(rawJson *v1ext.JSON) string {
	results := parseScalarJSONValue(rawJson)
	if len(results) == 0 {
		return ""
	}
	return toYaml(results)
}

// Best effort to parse raw json examples, controller-gen will break if the examples
func parseArrayJSONValue(rawJson *v1ext.JSON) []*yaml.Node {
	if rawJson == nil {
		return nil
	}
	var arrayContainer []interface{}
	if err := json.Unmarshal(rawJson.Raw, &arrayContainer); err != nil {
		panic(err)
	}
	_, node := parseJSONValue(arrayContainer)
	return node
}

func parseMapJSONValue(rawJson *v1ext.JSON) (node []*yaml.Node) {
	if rawJson == nil {
		return nil
	}
	var mapContainer map[string]interface{}
	if err := json.Unmarshal(rawJson.Raw, &mapContainer); err != nil {
		panic(err)
	}
	_, node = parseJSONValue(mapContainer)
	return node
}

func parseScalarJSONValue(rawJson *v1ext.JSON) string {
	if rawJson == nil {
		return ""
	}
	var scalarContainer interface{}
	if err := json.Unmarshal(rawJson.Raw, &scalarContainer); err != nil {
		panic(err)
	}
	scalar, _ := parseJSONValue(scalarContainer)
	return scalar
}

// Wil return either a string or a yaml node that could contain an array of objects
func parseJSONValue(jsonType interface{}) (value string, node []*yaml.Node) {
	if jsonType == nil {
		return
	}

	// Json supported types bool, int64, float64, string, []interface{}, map[string]interface{} and nil
	switch jsonVal := jsonType.(type) {
	case nil, bool, int64, float64, string:
		value = primativeToString(jsonVal)
	case []interface{}:
		for _, k := range jsonVal {
			value := primativeToString(k)
			node = append(node, &yaml.Node{Kind: yaml.ScalarNode, Value: value})
		}
	case map[string]interface{}:
		// TODO: parse objects and merge into a yaml node
		yamlBytes, err := yaml.Marshal(jsonVal)
		if err != nil {
			// Something has gone very wrong if we get to this point
			// crd generation will fail if either example/default values are not complient during
			// the controller-gen action.
			panic(err)
		}
		value = string(yamlBytes)
	}
	return
}

func schemaKeyToOrderedArray[K string | int, V any](schemaProperties map[K]V) []K {
	var ordered = []K{}
	for k, _ := range schemaProperties {
		ordered = append(ordered, k)
	}
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i] < ordered[j]
	})
	return ordered
}

func primativeToString(v interface{}) (value string) {
	switch jsonVal := v.(type) {
	case nil:
		value = "null"
	case bool:
		value = strconv.FormatBool(jsonVal)
	case int64:
		value = strconv.FormatInt(jsonVal, 10)
	case float64:
		value = strconv.FormatFloat(jsonVal, 'g', -1, 64)
	case string:
		value = jsonVal
	}
	return
}
