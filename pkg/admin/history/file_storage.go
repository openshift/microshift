package history

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
)

const (
	historyFilename = "history.json"
)

var (
	historyFilepath = path.Join(config.BackupsDir, historyFilename)
)

type HistoryStorage interface {
	Load() (*History, error)
	Save(*History) error
}

var _ HistoryStorage = (*HistoryFileStorage)(nil)

type HistoryFileStorage struct{}

func (hfs *HistoryFileStorage) Load() (*History, error) {
	if exists, err := util.PathExists(historyFilepath); err != nil {
		return nil, fmt.Errorf("checking if file %s exists failed: %w", historyFilepath, err)
	} else if !exists {
		return nil, ErrNoHistory
	}

	file, err := os.Open(historyFilepath)
	if err != nil {
		return nil, fmt.Errorf("opening file %s failed: %w", historyFilepath, err)
	}
	defer file.Close()

	buf, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("reading from file %s failed: %w", historyFilepath, err)
	}

	h := &History{}
	err = json.Unmarshal(buf, h)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling history from json failed: %w", err)
	}

	return h, nil
}

func (hfs *HistoryFileStorage) Save(history *History) error {
	b, err := json.Marshal(history)
	if err != nil {
		return fmt.Errorf("marshaling history to json failed: %w", err)
	}

	if err := util.MakeDir(config.BackupsDir); err != nil {
		return fmt.Errorf("making directory %s failed: %w", config.BackupsDir, err)
	}

	file, err := os.CreateTemp(config.BackupsDir, "history")
	if err != nil {
		return fmt.Errorf("creating temporary history file failed: %w", err)
	}
	defer file.Close()

	n, err := file.Write(b)
	if err != nil {
		return fmt.Errorf("writing to file %s failed: %w", file.Name(), err)
	}
	if n != len(b) {
		return fmt.Errorf("writing to file %s failed: wrote %d bytes, expected %d", file.Name(), n, len(b))
	}
	file.Close()

	if err := os.Rename(file.Name(), historyFilepath); err != nil {
		return fmt.Errorf("renaming %s to %s failed: %w", file.Name(), historyFilename, err)
	}

	return nil
}
