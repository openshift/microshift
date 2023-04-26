package ostree

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type action string

const (
	actionBackup  action = "backup"
	actionRestore action = "restore"

	nextBootFile = ".next_boot"
	filePerm     = os.FileMode(0644)
)

var getFileWriter = func() (io.Writer, error) {
	if err := EnsureAuxDirExists(); err != nil {
		return nil, err
	}
	return os.OpenFile(filepath.Join(auxDir, nextBootFile), os.O_RDWR|os.O_CREATE|os.O_TRUNC, filePerm)
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
