package prerun

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"github.com/openshift/microshift/pkg/version"
	"k8s.io/klog/v2"
)

var (
	versionFilePath = filepath.Join(config.DataDir, "version")

	errDataVersionDoesNotExist = errors.New("version file for MicroShift data does not exist")
)

// CreateOrValidateDataVersion creates or compares data version against executable's version
//
// Function is intended to be invoked by main MicroShift run procedure, just before starting,
// to ensure that storage migration (which should update the version file) was performed.
func CreateOrValidateDataVersion() error {
	dataVer, err := getVersionOfData()
	if err != nil {
		if errors.Is(err, errDataVersionDoesNotExist) {
			// First run of MicroShift, create version file shortly after creating DataDir
			return writeExecVersionToData()
		}
		return err
	}

	execVer, err := getVersionOfExecutable()
	if err != nil {
		return err
	}
	klog.InfoS("Comparing versions of MicroShift data on disk and executable", "data", dataVer, "exec", execVer)

	if execVer != dataVer {
		return fmt.Errorf("data version (%s) does not match binary version (%s) - missing migration?", dataVer, execVer)
	}

	return nil
}

func writeExecVersionToData() error {
	execVer, err := getVersionOfExecutable()
	if err != nil {
		return err
	}
	version := execVer.String()
	klog.InfoS("Writing MicroShift version to the file in data directory", "version", version)

	if err := os.WriteFile(versionFilePath, []byte(version), 0600); err != nil {
		return fmt.Errorf("writing '%s' to %s failed: %w", version, versionFilePath, err)
	}
	return nil
}

type versionMetadata struct {
	Major, Minor, Patch int
}

func (v versionMetadata) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
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

func getVersionOfExecutable() (versionMetadata, error) {
	ver := version.Get()
	return versionMetadataFromString(fmt.Sprintf("%s.%s.%s", ver.Major, ver.Minor, ver.Patch))
}

func getVersionOfData() (versionMetadata, error) {
	exists, err := util.PathExistsAndIsNotEmpty(versionFilePath)
	if err != nil {
		return versionMetadata{}, fmt.Errorf("checking if path exists failed: %w", err)
	}

	if !exists {
		return versionMetadata{}, errDataVersionDoesNotExist
	}

	versionFileContents, err := os.ReadFile(versionFilePath)
	if err != nil {
		return versionMetadata{}, fmt.Errorf("reading %s failed: %w", versionFilePath, err)
	}

	return versionMetadataFromString(string(versionFileContents))
}

// checkVersionDiff compares versions of executable and existing data for purposes of data migration.
// It returns true if the migration should be performed.
func checkVersionDiff(execVer, dataVer versionMetadata) (bool, error) {
	if execVer == dataVer {
		return false, nil
	}

	if execVer.Major != dataVer.Major {
		return false, fmt.Errorf("major versions are different: %d and %d", dataVer.Major, execVer.Major)
	}

	if execVer.Minor < dataVer.Minor {
		return false, fmt.Errorf("executable (%s) is older than existing data (%s): migrating data to older version is not supported", execVer.String(), dataVer.String())
	}

	if execVer.Minor > dataVer.Minor {
		if execVer.Minor-1 == dataVer.Minor {
			return true, nil
		} else {
			return false, fmt.Errorf("executable (%s) is too recent compared to existing data (%s): maximum minor version difference is 1", execVer.String(), dataVer.String())
		}
	}

	return IsUpgradeBlocked(execVer, dataVer)
}
