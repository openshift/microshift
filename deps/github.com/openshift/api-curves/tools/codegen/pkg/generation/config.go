package generation

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
)

// getAPIGroupConfigs populates the API group context with the
// configuration for the API group if it exists.
func getAPIGroupConfigs(apiGroups []APIGroupContext) error {
	for i := range apiGroups {
		if err := getAPIGroupConfig(&apiGroups[i]); err != nil {
			return fmt.Errorf("could not get API group config for %s: %w", apiGroups[i].Name, err)
		}
	}

	return nil
}

// getAPIGroupConfig populates the API group context with the
// configuration for the API group if it exists.
func getAPIGroupConfig(apiGroup *APIGroupContext) error {
	configPath, err := getConfigFilePath(apiGroup.Versions)
	if err != nil {
		return fmt.Errorf("could not get config file path: %w", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil
	}

	config, err := readAPIGroupConfig(configPath)
	if err != nil {
		return fmt.Errorf("could not read API group config: %w", err)
	}

	apiGroup.Config = config
	return nil
}

// getConfigFilePath returns the expected path to the API group config file.
// All versions of the API group must be contained with the same group directory.
// The config file is expected to be in the group directory.
func getConfigFilePath(versions []APIVersionContext) (string, error) {
	if len(versions) == 0 {
		return "", fmt.Errorf("no versions found")
	}

	groupDir := ""
	for _, version := range versions {
		baseDir := filepath.Dir(version.Path)
		if groupDir == "" {
			groupDir = baseDir
		}

		if baseDir != groupDir {
			return "", fmt.Errorf("versions found in different directories")
		}
	}

	return filepath.Join(groupDir, ".codegen.yaml"), nil
}

// readAPIGroupConfig reads the API group config file into a Config struct.
func readAPIGroupConfig(path string) (*Config, error) {
	config := &Config{}
	if err := readYAML(path, config); err != nil {
		return nil, fmt.Errorf("could not read API group config: %w", err)
	}

	if err := resolveLinks(path, config); err != nil {
		return nil, fmt.Errorf("could not resolve links: %w", err)
	}

	return config, nil
}

// readYAML reads a YAML file from a path into a struct.
func readYAML(path string, out interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not read file: %w", err)
	}

	if err := yaml.Unmarshal(data, out); err != nil {
		return fmt.Errorf("could not unmarshal YAML: %w", err)
	}

	return nil
}

// resolveLinks resolves any relative links from configuration files
// into absolute paths.
// This allows users to set links relative to the configuration file
// independent of the working directory of the generator.
func resolveLinks(filePath string, config *Config) error {
	if config == nil {
		return nil
	}

	if config.Deepcopy != nil {
		if err := resolveLink(filePath, &config.Deepcopy.HeaderFilePath); err != nil {
			return fmt.Errorf("could not resolve deepcopy header link: %w", err)
		}
	}

	return nil
}

// resolveLink resolves a relative link into an absolute path.
func resolveLink(filePath string, link *string) error {
	if link == nil || *link == "" {
		return nil
	}

	if filepath.IsAbs(*link) {
		return nil
	}

	absPath, err := filepath.Abs(filepath.Join(filepath.Dir(filePath), *link))
	if err != nil {
		return fmt.Errorf("could not get absolute path for link: %w", err)
	}

	*link = absPath
	return nil
}
