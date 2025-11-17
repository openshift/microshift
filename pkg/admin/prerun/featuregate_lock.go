package prerun

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

var (
	featureGateLockFilePath = filepath.Join(config.DataDir, "no-upgrade")
	errLockFileDoesNotExist = errors.New("feature gate lock file does not exist")
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
	// Check if custom feature gates are configured
	if !isCustomFeatureGatesConfigured(cfg.ApiServer.FeatureGates) {
		klog.InfoS("No custom feature gates configured - skipping lock file management")
		return nil
	}

	klog.InfoS("Custom feature gates detected", "featureSet", cfg.ApiServer.FeatureGates.FeatureSet)

	// Check if lock file exists
	lockExists, err := util.PathExists(featureGateLockFilePath)
	if err != nil {
		return fmt.Errorf("failed to check if lock file exists: %w", err)
	}

	if !lockExists {
		// First time configuring custom feature gates - create lock file
		return createFeatureGateLockFile(cfg)
	}

	// Lock file exists - validate configuration hasn't changed
	return validateFeatureGateLockFile(cfg)
}

// isCustomFeatureGatesConfigured checks if any custom feature gates are configured
func isCustomFeatureGatesConfigured(fg config.FeatureGates) bool {
	// Empty feature set means no custom feature gates
	if fg.FeatureSet == "" {
		return false
	}

	// TechPreviewNoUpgrade and DevPreviewNoUpgrade are considered custom
	if fg.FeatureSet == config.FeatureSetTechPreviewNoUpgrade ||
		fg.FeatureSet == config.FeatureSetDevPreviewNoUpgrade {
		return true
	}

	// CustomNoUpgrade requires actual enabled or disabled features
	if fg.FeatureSet == config.FeatureSetCustomNoUpgrade {
		return len(fg.CustomNoUpgrade.Enabled) > 0 || len(fg.CustomNoUpgrade.Disabled) > 0
	}

	return false
}

// createFeatureGateLockFile creates the lock file with current configuration
func createFeatureGateLockFile(cfg *config.Config) error {
	klog.InfoS("Creating feature gate lock file - this cluster can no longer be upgraded",
		"path", featureGateLockFilePath)

	// Get current version from version file
	currentVersion, err := getVersionOfData()
	if err != nil {
		// If version file doesn't exist yet, get executable version
		klog.InfoS("Version file does not exist yet, using executable version")
		currentVersion, err = GetVersionOfExecutable()
		if err != nil {
			return fmt.Errorf("failed to get version: %w", err)
		}
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

// validateFeatureGateLockFile validates that the current configuration matches the lock file
// and that no version upgrade has occurred
func validateFeatureGateLockFile(cfg *config.Config) error {
	klog.InfoS("Validating feature gate lock file", "path", featureGateLockFilePath)

	lockFile, err := readFeatureGateLockFile(featureGateLockFilePath)
	if err != nil {
		return fmt.Errorf("failed to read lock file: %w", err)
	}

	// Check if feature gate configuration has changed
	if err := compareFeatureGates(lockFile, cfg.ApiServer.FeatureGates); err != nil {
		return fmt.Errorf("feature gate configuration has changed: %w\n\n"+
			"Custom feature gates cannot be modified or reverted once applied.\n"+
			"To restore MicroShift to a supported state, you must:\n"+
			"1. Run: sudo microshift-cleanup-data --all\n"+
			"2. Remove custom feature gates from /etc/microshift/config.yaml\n"+
			"3. Restart MicroShift: sudo systemctl restart microshift", err)
	}

	// Check if version has changed (upgrade attempted)
	currentVersion, err := getVersionOfData()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if lockFile.Version != currentVersion {
		return fmt.Errorf("version upgrade detected with custom feature gates: locked version %s, current version %s\n\n"+
			"Upgrades are not supported when custom feature gates are configured.\n"+
			"Custom feature gates (%s) were configured in version %s.\n"+
			"To restore MicroShift to a supported state, you must:\n"+
			"1. Roll back to version %s, OR\n"+
			"2. Run: sudo microshift-cleanup-data --all\n"+
			"3. Remove custom feature gates from /etc/microshift/config.yaml\n"+
			"4. Restart MicroShift: sudo systemctl restart microshift",
			lockFile.Version.String(), currentVersion.String(),
			lockFile.FeatureSet, lockFile.Version.String(), lockFile.Version.String())
	}

	klog.InfoS("Feature gate lock file validation successful")
	return nil
}

// compareFeatureGates compares the lock file with current configuration
func compareFeatureGates(lockFile featureGateLockFile, current config.FeatureGates) error {
	if lockFile.FeatureSet != current.FeatureSet {
		return fmt.Errorf("feature set changed: locked config has %q, current config has %q",
			lockFile.FeatureSet, current.FeatureSet)
	}

	if !reflect.DeepEqual(lockFile.CustomNoUpgrade, current.CustomNoUpgrade) {
		return fmt.Errorf("custom feature gates changed: locked config has %#v, current config has %#v",
			lockFile.CustomNoUpgrade, current.CustomNoUpgrade)
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
