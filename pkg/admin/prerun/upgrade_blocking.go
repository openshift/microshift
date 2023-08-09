package prerun

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	embedded "github.com/openshift/microshift/assets"
	"k8s.io/klog/v2"
)

func isUpgradeBlocked(execVersion versionMetadata, dataVersion versionMetadata) error {
	klog.InfoS("START obtaining list of blocked upgrades")
	buf, err := getBlockedUpgradesAsset()
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			klog.InfoS("SKIP obtaining list of blocked upgrades - embedded blocked upgrades asset does not exist - skipping check if upgrade is blocked")
			return nil
		}
		klog.ErrorS(err, "FAIL obtaining list of blocked upgrades")
		return err
	}
	m, err := unmarshalBlockedUpgrades(buf)
	if err != nil {
		klog.ErrorS(err, "FAIL unmarshal blocked upgrades asset", "asset", strings.ReplaceAll(string(buf), "\n", ""))
		return err
	}
	klog.InfoS("END obtaining list of blocked upgrades", "blocked-upgrades", m)

	klog.InfoS("START checking if upgrade is blocked", "existing-data-version", dataVersion, "new-binary-version", execVersion)
	if err := isBlocked(m, execVersion.String(), dataVersion.String()); err != nil {
		klog.ErrorS(err, "FAIL upgrade is blocked")
		return err
	}
	klog.InfoS("END upgrade is not blocked", "existing-data-version", dataVersion, "new-binary-version", execVersion)
	return nil
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

func isBlocked(blockedUpgrades map[string][]string, execVersion, dataVersion string) error {
	for targetVersion, fromVersions := range blockedUpgrades {
		if targetVersion == execVersion {
			for _, from := range fromVersions {
				if from == dataVersion {
					return fmt.Errorf("upgrade from %q to %q is blocked", dataVersion, execVersion)
				}
			}
		}
	}
	return nil
}
