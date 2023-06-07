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
	execVer, err := getVersionOfExecutable()
	if err != nil {
		return err
	}

	dataVer, err := getVersionOfData()
	if err != nil {
		if errors.Is(err, errDataVersionDoesNotExist) {
			// First run of MicroShift, create version file shortly after creating DataDir
			execVerS := execVer.String()
			klog.InfoS("Version file in data directory does not exist - creating", "version", execVerS)

			if err := os.WriteFile(versionFilePath, []byte(execVerS), 0600); err != nil {
				return fmt.Errorf("writing '%s' to %s failed: %w", execVerS, versionFilePath, err)
			}
			return nil
		}
		return err
	}
	klog.InfoS("Comparing versions of MicroShift data on disk and executable", "data", dataVer, "exec", execVer)

	if execVer != dataVer {
		return fmt.Errorf("data version (%s) does not match binary version (%s) - missing migration?", dataVer, execVer)
	}

	return nil
}

type versionMetadata struct {
	X, Y int
}

func (v versionMetadata) String() string {
	return fmt.Sprintf("%d.%d", v.X, v.Y)
}

// versionMetadataFromString creates versionMetadata object from "X.Y" string where X and Y are integers
func versionMetadataFromString(xy string) (versionMetadata, error) {
	xy = strings.TrimSpace(xy)
	xys := strings.Split(xy, ".")
	if len(xys) != 2 {
		return versionMetadata{}, fmt.Errorf("invalid version string (%s): expected X.Y", xy)
	}

	x, err := strconv.Atoi(xys[0])
	if err != nil {
		return versionMetadata{}, fmt.Errorf("converting '%s' to an int failed: %w", xys[0], err)
	}

	y, err := strconv.Atoi(xys[1])
	if err != nil {
		return versionMetadata{}, fmt.Errorf("converting '%s' to an int failed: %w", xys[1], err)
	}

	return versionMetadata{X: x, Y: y}, nil
}

func getVersionOfExecutable() (versionMetadata, error) {
	ver := version.Get()
	return versionMetadataFromString(fmt.Sprintf("%s.%s", ver.Major, ver.Minor))
}

func getVersionOfData() (versionMetadata, error) {
	exists, err := util.PathExists(versionFilePath)
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
