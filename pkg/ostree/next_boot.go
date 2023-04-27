package ostree

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/openshift/microshift/pkg/config"
	"k8s.io/klog/v2"
)

type action string

const (
	actionBackup  action = "backup"
	actionRestore action = "restore"
	actionMissing action = "missing"

	nextBootFile = ".next_boot"
	filePerm     = os.FileMode(0644)
)

var nextBootFilePath = filepath.Join(config.AuxDataDir, nextBootFile)

var getFileWriter = func() (io.Writer, error) {
	if err := config.EnsureAuxDirExists(); err != nil {
		return nil, err
	}
	return os.OpenFile(nextBootFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, filePerm)
}

var getFileReader = func() (io.Reader, error) {
	return os.Open(nextBootFilePath)
}

var fileExists = func() (bool, error) {
	if _, err := os.Stat(nextBootFilePath); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

type nextBoot struct {
	Action   action `json:"action"`
	OstreeID string `json:"ostree,omitempty"`
}

func (nb *nextBoot) Persist() error {
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
			"writing nextBoot was incomplete - wrote %d bytes, expected %d", n, len(b))
	}

	return nil
}

func (nb *nextBoot) RemoveFromDisk() error {
	exists, err := fileExists()

	if err != nil {
		return err
	}
	if exists {
		klog.Infof("Removing %s", nextBootFilePath)
		return os.Remove(nextBootFilePath)
	}
	return nil
}

func nextBootFromDisk() (*nextBoot, error) {
	if exists, err := fileExists(); err != nil {
		return nil, fmt.Errorf("problem with next boot file: %w", err)
	} else if !exists {
		return &nextBoot{Action: actionMissing}, nil
	}

	reader, err := getFileReader()
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	nb := &nextBoot{}
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

	klog.Infof("Loaded next boot action: %#v", nb)
	return nb, nil
}
