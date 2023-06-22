package prerun

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"

	embedded "github.com/openshift/microshift/assets"
	"k8s.io/klog/v2"
)

func IsUpgradeBlocked(execVersion versionMetadata, dataVersion versionMetadata) (bool, error) {
	buf, err := getBlockedUpgradesAsset()
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("failed to load embedded blocked upgrades asset: %w", err)
	}

	m, err := unmarshalBlockedUpgrades(buf)
	if err != nil {
		return false, err
	}

	return isBlocked(m, execVersion.String(), dataVersion.String()), nil
}

func getBlockedUpgradesAsset() ([]byte, error) {
	return embedded.Asset("release/upgrade-blocks.json")
}

func unmarshalBlockedUpgrades(data []byte) (map[string][]string, error) {
	var blockedEdges map[string][]string
	err := json.Unmarshal(data, &blockedEdges)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal %q: %w", string(data), err)
	}
	return blockedEdges, nil
}

func isBlocked(blockedUpgrades map[string][]string, execVersion, dataVersion string) bool {
	klog.InfoS("Checking if upgrade is allowed", "existing-data-version", dataVersion, "new-binary-version", execVersion, "blocked-upgrades", blockedUpgrades)

	for targetVersion, fromVersions := range blockedUpgrades {
		if targetVersion == execVersion {
			for _, from := range fromVersions {
				if from == dataVersion {
					klog.ErrorS(nil, "Detected an attempt of unsupported upgrade", "existing-data-version", dataVersion, "new-binary-version", execVersion)
					return true
				}
			}
		}
	}

	klog.InfoS("Upgrade is allowed", "existing-data-version", dataVersion, "new-binary-version", execVersion)
	return false
}
