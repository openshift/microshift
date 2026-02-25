package prerun

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

var (
	featureGateLockFilePath = filepath.Join(config.DataDir, "no-upgrade")
	errLockFileDoesNotExist = errors.New("feature gate lock file does not exist")
	// getExecutableVersion is a function variable that allows tests to override the version
	getExecutableVersion = GetVersionOfExecutable
)

// featureGateLockFile represents the structure of the lock file
// that tracks custom feature gate configuration and prevents changes/upgrades
type featureGateLockFile struct {
	FeatureSet      string                 `json:"featureSet"`
	CustomNoUpgrade config.CustomNoUpgrade `json:"customNoUpgrade"`
	Version         versionMetadata        `json:"version"`
}

// FeatureGateLockManagement manages the feature gate lock file
// that prevents upgrades and config changes when custom feature gates are configured
func FeatureGateLockManagement(cfg *config.Config) error {
	klog.InfoS("START feature gate lock management")
	if err := featureGateLockManagement(cfg); err != nil {
		klog.ErrorS(err, "FAIL feature gate lock management")
		return err
	}
	klog.InfoS("END feature gate lock management")
	return nil
}

func featureGateLockManagement(cfg *config.Config) error {
	// If a lock file exists, it must be validated regardless of current config
	// This prevents users from removing feature gates from config in order to block upgrades and configuration changes
	lockExists, err := util.PathExists(featureGateLockFilePath)
	if err != nil {
		return fmt.Errorf("failed to check if lock file exists: %w", err)
	}
	// Lock file exists - validate configuration
	if lockExists {
		return runValidationsChecks(cfg)
	}
	// No lock file exists yet and custom feature gates are configured, so this is the first time configuring custom feature gates
	if cfg.ApiServer.FeatureGates.FeatureSet != "" {
		return createFeatureGateLockFile(cfg)
	}
	// No lock file and no custom feature gates - normal operation
	return nil
}

// createFeatureGateLockFile creates the lock file with current configuration
func createFeatureGateLockFile(cfg *config.Config) error {
	klog.InfoS("Creating feature gate lock file - this cluster can no longer be upgraded",
		"path", featureGateLockFilePath)

	// Get current version from executable
	currentVersion, err := getExecutableVersion()
	if err != nil {
		return fmt.Errorf("failed to get version: %w", err)
	}

	lockFile := featureGateLockFile{
		FeatureSet:      cfg.ApiServer.FeatureGates.FeatureSet,
		CustomNoUpgrade: cfg.ApiServer.FeatureGates.CustomNoUpgrade,
		Version:         currentVersion,
	}

	if err := writeFeatureGateLockFile(featureGateLockFilePath, lockFile); err != nil {
		return fmt.Errorf("failed to write lock file: %w", err)
	}

	klog.InfoS("Feature gate lock file created successfully",
		"featureSet", lockFile.FeatureSet,
		"version", lockFile.Version.String())

	return nil
}

// runValidationsChecks validates the feature gate lock file and the current configuration
// It returns an error if the configuration is invalid or if an x or y stream version upgrade has occurred.
func runValidationsChecks(cfg *config.Config) error {
	klog.InfoS("Validating feature gate lock file", "path", featureGateLockFilePath)

	lockFile, err := readFeatureGateLockFile(featureGateLockFilePath)
	if err != nil {
		return fmt.Errorf("failed to read lock file: %w", err)
	}

	// Check if feature gate configuration has changed
	if err := configValidationChecksPass(lockFile, cfg.ApiServer.FeatureGates); err != nil {
		return fmt.Errorf("detected invalid changes in feature gate configuration: %w\n\n"+
			"To restore MicroShift to a supported state, you must:\n"+
			"1. Run: sudo microshift-cleanup-data --all\n"+
			"2. Remove custom feature gates from /etc/microshift/config.yaml\n"+
			"3. Restart MicroShift: sudo systemctl restart microshift", err)
	}

	// Check if version has changed (upgrade attempted)
	currentExecutableVersion, err := getExecutableVersion()
	if err != nil {
		return fmt.Errorf("failed to get current executable version: %w", err)
	}

	if lockFile.Version.Major != currentExecutableVersion.Major || lockFile.Version.Minor != currentExecutableVersion.Minor {
		return fmt.Errorf("version upgrade detected with custom feature gates: locked version %s, current version %s\n\n"+
			"Upgrades are not supported when custom feature gates are configured.\n"+
			"Custom feature gates (%s) were configured in version %s.\n"+
			"To restore MicroShift to a supported state, you must:\n"+
			"1. Roll back to version %s, OR\n"+
			"2. Run: sudo microshift-cleanup-data --all\n"+
			"3. Remove custom feature gates from /etc/microshift/config.yaml\n"+
			"4. Restart MicroShift: sudo systemctl restart microshift",
			lockFile.Version.String(), currentExecutableVersion.String(),
			lockFile.FeatureSet, lockFile.Version.String(), lockFile.Version.String())
	}

	klog.InfoS("Feature gate lock file validation successful")
	return nil
}

func configValidationChecksPass(prev featureGateLockFile, current config.FeatureGates) error {
	if prev.FeatureSet != "" && current.FeatureSet == "" {
		// Disallow changing from feature set to no feature set
		return fmt.Errorf("cannot unset feature set. Previous config had feature set %q, current config has no feature set configured", prev.FeatureSet)
	}
	if prev.FeatureSet == config.FeatureSetCustomNoUpgrade && current.FeatureSet != config.FeatureSetCustomNoUpgrade {
		// Disallow changing from custom feature gates to any other feature set
		return fmt.Errorf("cannot change CustomNoUpgrade feature set. Previous feature set was %q, current feature set is %q", prev.FeatureSet, current.FeatureSet)
	}
	return nil
}

// writeFeatureGateLockFile writes the lock file to disk in YAML format
func writeFeatureGateLockFile(path string, lockFile featureGateLockFile) error {
	data, err := yaml.Marshal(lockFile)
	if err != nil {
		return fmt.Errorf("failed to marshal lock file: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write lock file to %q: %w", path, err)
	}

	return nil
}

// readFeatureGateLockFile reads the lock file from disk in YAML format
func readFeatureGateLockFile(path string) (featureGateLockFile, error) {
	exists, err := util.PathExists(path)
	if err != nil {
		return featureGateLockFile{}, fmt.Errorf("failed to check if lock file exists: %w", err)
	}

	if !exists {
		return featureGateLockFile{}, errLockFileDoesNotExist
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return featureGateLockFile{}, fmt.Errorf("failed to read lock file from %q: %w", path, err)
	}

	var lockFile featureGateLockFile
	if err := yaml.Unmarshal(data, &lockFile); err != nil {
		return featureGateLockFile{}, fmt.Errorf("failed to unmarshal lock file: %w", err)
	}

	return lockFile, nil
}
