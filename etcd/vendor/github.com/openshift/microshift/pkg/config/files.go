package config

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/openshift/microshift/pkg/util"
	"sigs.k8s.io/yaml"
)

const (
	ConfigFile      = "/etc/microshift/config.yaml"
	DataDir         = "/var/lib/microshift"
	BackupsDir      = "/var/lib/microshift-backups"
	ConfigDropInDir = "/etc/microshift/config.d"
)

func getActiveConfigFromYAMLDropins(yamlDropins [][]byte) (*Config, error) {
	var mergedUserConfigPatch []byte

	// Convert YAMLs to JSONs and merge them together to get a single configuration from the user.
	for _, dropin := range yamlDropins {
		if strings.TrimSpace(string(dropin)) == "" {
			continue
		}

		jsonDropin, err := yaml.YAMLToJSON(dropin)
		if err != nil {
			return nil, fmt.Errorf("failed to convert config yaml (%q) to json: %w", string(dropin), err)
		}

		if mergedUserConfigPatch == nil {
			mergedUserConfigPatch = jsonDropin
			continue
		}

		patched, err := jsonpatch.MergePatch(mergedUserConfigPatch, jsonDropin)
		if err != nil {
			return nil, fmt.Errorf("failed to merge dropin (%q) into the config patch (%q): %w", string(jsonDropin), string(mergedUserConfigPatch), err)
		}
		mergedUserConfigPatch = patched
	}

	cfg := &Config{}
	if err := cfg.fillDefaults(); err != nil {
		return nil, fmt.Errorf("failed to fill config's defaults: %w", err)
	}

	if len(mergedUserConfigPatch) != 0 {
		userSettings := &Config{}
		if err := json.Unmarshal(mergedUserConfigPatch, userSettings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user cfg json to config: %w", err)
		}
		cfg.incorporateUserSettings(userSettings)
	}

	if err := cfg.updateComputedValues(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// collectUserProvidedConfigs loads all the user provided yaml config files:
// - main MicroShift config (/etc/microshift/config.yaml), and
// - YAML files from config drop-in directory (/etc/microshift/config.d)
func collectUserProvidedConfigs() ([][]byte, error) {
	dropins := [][]byte{}

	if exists, err := util.PathExists(ConfigFile); err != nil {
		return nil, err
	} else if exists {
		contents, err := os.ReadFile(ConfigFile)
		if err != nil {
			return nil, fmt.Errorf("error reading config file %q: %v", ConfigFile, err)
		}
		dropins = append(dropins, contents)
	}

	dropInDirExists, err := util.PathExistsAndIsNotEmpty(ConfigDropInDir)
	if err != nil {
		return nil, err
	}

	if !dropInDirExists {
		return dropins, nil
	}

	err = filepath.WalkDir(ConfigDropInDir, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(info.Name()) == ".yaml" {
			contents, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("error reading config file %q: %v", path, err)
			}
			dropins = append(dropins, contents)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk the config drop-in dir %q: %w", ConfigDropInDir, err)
	}

	return dropins, nil
}

// ActiveConfig returns the active configuration which is default config with overrides
// from user provided config files.
func ActiveConfig() (*Config, error) {
	dropins, err := collectUserProvidedConfigs()
	if err != nil {
		return nil, err
	}

	return getActiveConfigFromYAMLDropins(dropins)
}
