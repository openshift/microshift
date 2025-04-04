package prerun

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/openshift/microshift/pkg/admin/data"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"github.com/openshift/microshift/pkg/version"
	"k8s.io/klog/v2"
)

var (
	versionFilePath = filepath.Join(config.DataDir, "version")

	errDataVersionDoesNotExist = errors.New("version file for MicroShift data does not exist")
)

type versionFile struct {
	Version      versionMetadata `json:"version"`
	DeploymentID string          `json:"deployment_id,omitempty"`
	BootID       string          `json:"boot_id"`
}

const MAX_VERSION_SKEW = 2

func (hi *versionFile) BackupName() data.BackupName {
	return data.BackupName(fmt.Sprintf("%s_%s", hi.DeploymentID, hi.BootID))
}

func VersionMetadataManagement() error {
	klog.InfoS("START version metadata management")
	if err := versionMetadataManagement(); err != nil {
		klog.ErrorS(err, "FAIL version metadata management")
		return err
	}
	klog.InfoS("END version metadata management")
	return nil
}

func versionMetadataManagement() error {
	klog.InfoS("START getting versions")
	ver, err := getVersions()
	if err != nil {
		klog.ErrorS(err, "FAIL getting versions")
		return err
	}
	klog.InfoS("END getting versions", "exec", ver.exec, "data", ver.data)

	if ver.data == nil {
		klog.InfoS("SKIP version compatibility checks - data does not exist")
	} else {
		klog.InfoS("START version compatibility checks")
		if err := checkVersionCompatibility(ver.exec, *ver.data); err != nil {
			klog.ErrorS(err, "FAIL version compatibility checks")
			return err
		}
		klog.InfoS("END version compatibility checks")

		klog.InfoS("START checking if version upgrade is blocked")
		if err := isUpgradeBlocked(ver.exec, *ver.data); err != nil {
			klog.ErrorS(err, "FAIL checking if version upgrade is blocked")
			return err
		}
		klog.InfoS("END checking if version upgrade is blocked")
	}

	klog.InfoS("START updating version file")
	if err := updateVersionFile(ver.exec); err != nil {
		klog.ErrorS(err, "FAIL updating version file")
		return err
	}
	klog.InfoS("END updating version file")

	return nil
}

type versions struct {
	exec versionMetadata
	data *versionMetadata
}

// getVersions obtains and returns versions of executable and data dir.
// Version of data will be nil if the MicroShift data does not exist yet.
func getVersions() (versions, error) {
	execVer, err := GetVersionOfExecutable()
	if err != nil {
		return versions{}, fmt.Errorf("failed to get version of MicroShift executable: %w", err)
	}

	vs := versions{
		exec: execVer,
		data: nil,
	}

	dataVer, err := getVersionOfData()
	if err == nil {
		vs.data = &dataVer
		return vs, nil
	}

	if !errors.Is(err, errDataVersionDoesNotExist) {
		// error is something else than "file does not exist", like permissions
		return versions{}, fmt.Errorf("failed to get version of existing MicroShift data: %w", err)
	}

	// Ignoring .nodename to not get false positives from mere existence of the path
	dataExists, err := util.PathExistsAndIsNotEmpty(config.DataDir, ".nodename")
	if err != nil {
		return versions{}, err
	}

	if !dataExists {
		// Data directory does not exist so it's first run of MicroShift
		klog.InfoS("Version file does not exist yet - assuming first run of MicroShift")
		vs.data = nil // repeated for clarity
		return vs, nil
	}

	// Data exists but without version file, let's assume 4.13 and compare versions
	klog.InfoS("MicroShift data directory exists, but doesn't contain version file" +
		" - assuming 4.13.0 and proceeding with version compatibility checks")
	vs.data = &versionMetadata{Major: 4, Minor: 13, Patch: 0}
	return vs, nil
}

func updateVersionFile(ver versionMetadata) error {
	currentDeploymentID := ""
	isOstree, err := util.IsOSTree()
	if err != nil {
		return fmt.Errorf("failed to check if system is ostree: %w", err)
	} else if isOstree {
		currentDeploymentID, err = GetCurrentDeploymentID()
		if err != nil {
			return fmt.Errorf("failed to get current deployment ID: %w", err)
		}
	}

	currentBootID, err := getCurrentBootID()
	if err != nil {
		return fmt.Errorf("failed to get current boot ID: %w", err)
	}

	v := versionFile{
		Version:      ver,
		DeploymentID: currentDeploymentID,
		BootID:       currentBootID,
	}

	klog.InfoS("Version file contents to write", "data", v)

	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal %v: %w", v, err)
	}

	if err := os.WriteFile(versionFilePath, data, 0600); err != nil {
		return fmt.Errorf("writing %q to %q failed: %w", string(data), versionFilePath, err)
	}

	if isOstree {
		if err := updateHealthInfo(v); err != nil {
			return fmt.Errorf("failed to update health.json: %w", err)
		}
	}

	return nil
}

type versionMetadata struct {
	Major, Minor, Patch int
}

func (v *versionMetadata) String() string {
	if v == nil {
		return "nil"
	}
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v versionMetadata) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

func (v *versionMetadata) UnmarshalJSON(data []byte) error {
	str := ""
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	ver, err := versionMetadataFromString(str)
	if err != nil {
		return err
	}

	v.Major = ver.Major
	v.Minor = ver.Minor
	v.Patch = ver.Patch
	return nil
}

// versionMetadataFromString creates versionMetadata object from "major.minor.patch" string where major, minor, and patch are integers
func versionMetadataFromString(majorMinorPatch string) (versionMetadata, error) {
	majorMinorPatch = strings.TrimSpace(majorMinorPatch)
	split := strings.Split(majorMinorPatch, ".")
	if len(split) != 3 {
		return versionMetadata{}, fmt.Errorf("invalid version string (%s): expected Major.Minor.Patch", majorMinorPatch)
	}

	major, err := strconv.Atoi(split[0])
	if err != nil {
		return versionMetadata{}, fmt.Errorf("converting %q to an int failed: %w", split[0], err)
	}

	minor, err := strconv.Atoi(split[1])
	if err != nil {
		return versionMetadata{}, fmt.Errorf("converting %q to an int failed: %w", split[1], err)
	}

	patch, err := strconv.Atoi(split[2])
	if err != nil {
		return versionMetadata{}, fmt.Errorf("converting %q to an int failed: %w", split[2], err)
	}

	return versionMetadata{Major: major, Minor: minor, Patch: patch}, nil
}

func GetVersionOfExecutable() (versionMetadata, error) {
	ver := version.Get()
	return versionMetadataFromString(fmt.Sprintf("%s.%s.%s", ver.Major, ver.Minor, ver.Patch))
}

func getVersionOfData() (versionMetadata, error) {
	klog.InfoS("START reading version file")
	verFile, err := getVersionFile()
	if err != nil {
		klog.ErrorS(err, "FAIL reading version file")
		return versionMetadata{}, err
	}
	klog.InfoS("END reading version file", "contents", verFile)
	return verFile.Version, nil
}

func getVersionFile() (versionFile, error) {
	exists, err := util.PathExistsAndIsNotEmpty(versionFilePath)
	if err != nil {
		return versionFile{}, fmt.Errorf("checking if path exists failed: %w", err)
	}

	if !exists {
		return versionFile{}, errDataVersionDoesNotExist
	}

	versionFileContents, err := os.ReadFile(versionFilePath)
	if err != nil {
		return versionFile{}, fmt.Errorf("reading %q failed: %w", versionFilePath, err)
	}
	return parseVersionFile(versionFileContents)
}

func parseVersionFile(contents []byte) (versionFile, error) {
	verFile := versionFile{}
	jsonErr := json.Unmarshal(contents, &verFile)
	if jsonErr == nil {
		return verFile, nil
	}

	// Unmarshalling version as json failed - fallback to using previous version file schema
	verMetadata, fallbackErr := versionMetadataFromString(string(contents))
	if fallbackErr != nil {
		return versionFile{},
			fmt.Errorf("parsing %q failed: %w", string(contents), errors.Join(jsonErr, fallbackErr))
	}

	return versionFile{
		Version:      verMetadata,
		BootID:       "",
		DeploymentID: "",
	}, nil
}

func GetVersionStringOfData() string {
	versionMetadata, err := getVersionOfData()
	if err != nil {
		if errors.Is(err, errDataVersionDoesNotExist) {
			dataExists, err := util.PathExistsAndIsNotEmpty(config.DataDir, ".nodename")
			if err == nil && dataExists {
				// version does not exists, but data exists
				return "4.13"
			}
		}
		klog.ErrorS(err, "Failed to read version - ignoring error")
		// couldn't access the file for whatever reason
		return "unknown"
	}
	return versionMetadata.String()
}

// checkVersionCompatibility compares versions of executable and existing data
// to detect unsupported version changes
func checkVersionCompatibility(execVer, dataVer versionMetadata) error {
	if execVer == dataVer {
		klog.InfoS("Executable and data versions are the same - continuing")
		return nil
	}

	if execVer.Major != dataVer.Major {
		return fmt.Errorf("major versions are different: %d and %d", dataVer.Major, execVer.Major)
	}

	if execVer.Minor < dataVer.Minor {
		return fmt.Errorf("executable (%s) is older than existing data (%s): migrating data to older version is not supported", execVer.String(), dataVer.String())
	}

	if execVer.Minor > dataVer.Minor {
		versionSkew := execVer.Minor - dataVer.Minor
		if versionSkew <= MAX_VERSION_SKEW {
			klog.Infof("Executable is newer than data by %d minor versions, continuing", versionSkew)
			return nil
		} else {
			return fmt.Errorf("executable (%s) is too recent compared to existing data (%s): minor version difference is %d, maximum allowed difference is %d",
				execVer.String(), dataVer.String(), versionSkew, MAX_VERSION_SKEW)
		}
	}

	klog.InfoS("All version checks passed")
	return nil
}
