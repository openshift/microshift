package ostree

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"k8s.io/klog/v2"
)

type action string

const (
	actionBackup  action = "backup"
	actionRestore action = "restore"
	actionMissing action = "missing"

	persistenceFile = "pre_run_action"
	filePerm        = os.FileMode(0644)
)

var preRunActionFilepath = filepath.Join(config.AuxDataDir, persistenceFile)

var getFileWriter = func() (io.Writer, error) {
	if err := config.EnsureAuxDirExists(); err != nil {
		return nil, err
	}
	return os.OpenFile(preRunActionFilepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, filePerm)
}

var getFileReader = func() (io.Reader, error) {
	return os.Open(preRunActionFilepath)
}

var fileExists = func() (bool, error) {
	return util.CheckIfFileExists(preRunActionFilepath)
}

type preRunAction struct {
	Action   action `json:"action"`
	OstreeID string `json:"ostree,omitempty"`
}

func (nb *preRunAction) Persist() error {
	w, err := getFileWriter()
	if err != nil {
		return err
	}

	b, err := json.Marshal(nb)
	if err != nil {
		return err
	}
	n, err := w.Write(b)
	if err != nil {
		return err
	}

	if n != len(b) {
		return fmt.Errorf(
			"writing pre-run-action was incomplete - wrote %d bytes, expected %d", n, len(b))
	}

	klog.Infof("Persisted next pre-run action '%s' to %s", string(b), preRunActionFilepath)
	return nil
}

func (nb *preRunAction) RemoveFromDisk() error {
	exists, err := fileExists()

	if err != nil {
		return err
	}
	if exists {
		klog.Infof("Removing %s", preRunActionFilepath)
		return os.Remove(preRunActionFilepath)
	}
	return nil
}

func preRunActionFromDisk() (*preRunAction, error) {
	if exists, err := fileExists(); err != nil {
		return nil, fmt.Errorf("problem with pre-run-action file: %w", err)
	} else if !exists {
		return &preRunAction{Action: actionMissing}, nil
	}

	reader, err := getFileReader()
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	nb := &preRunAction{}
	err = json.Unmarshal(b, nb)
	if err != nil {
		return nil, err
	}

	switch nb.Action {
	case actionBackup, actionRestore:
	case actionMissing:
		return nil, fmt.Errorf("deserialized 'missing' which shouldn't happen")
	default:
		return nil, fmt.Errorf("unknown action deserialized: %s", nb.Action)
	}

	klog.Infof("Loaded pre-run-action: %#v", nb)
	return nb, nil
}
