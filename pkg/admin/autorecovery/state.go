package autorecovery

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/openshift/microshift/pkg/admin/data"
	"github.com/openshift/microshift/pkg/util"
	"k8s.io/klog/v2"
)

const (
	stateFilename = "state.json"
)

type state struct {
	LastBackup       data.BackupName `json:"LastBackup"`
	storageFS        fsConstraint    `json:"-"`
	storagePath      string          `json:"-"`
	intermediatePath string          `json:"-"`
}

func (s *state) Serialize() ([]byte, error) {
	return json.Marshal(s)
}

func (s *state) SaveToIntermediate() error {
	contents, err := s.Serialize()
	if err != nil {
		return err
	}
	path, err := data.GenerateUniqueTempPath(filepath.Join(s.storagePath, stateFilename))
	if err != nil {
		return err
	}
	s.intermediatePath = path
	klog.InfoS("Saving intermediate state", "state", contents, "path", path)
	return os.WriteFile(path, contents, 0600)
}

func (s *state) MoveToFinal() error {
	path := filepath.Join(s.storagePath, stateFilename)
	klog.InfoS("Moving state file to final path", "intermediatePath", s.intermediatePath, "finalPath", path)
	return os.Rename(s.intermediatePath, path)
}

func NewState(storagePath data.StoragePath, lastBackup data.BackupName) *state {
	return &state{
		LastBackup:  lastBackup,
		storagePath: string(storagePath),
	}
}

type fsConstraint interface {
	fs.StatFS
	fs.ReadFileFS
}

func GetState(storagePath data.StoragePath) (*state, error) {
	sp := string(storagePath)
	storageFS, ok := os.DirFS(sp).(fsConstraint)
	if !ok {
		// This should never happen.
		// Concrete type returned from os.DirFS() implements interfaces specified in fsConstaint,
		// but the function actually returns fs.FS interface which cannot be used
		// as more specialized interface without casting.
		err := fmt.Errorf("failed to cast os.DirFS to more specialized type")
		klog.ErrorS(err, "Something went terribly wrong")
		return nil, err
	}

	state, err := getStateFS(storageFS)
	if err != nil {
		return nil, err
	}
	if state != nil {
		klog.InfoS("Read state from the disk", "state", state)
		state.storagePath = sp
	}

	return state, nil
}

func getStateFS(fsys fsConstraint) (*state, error) {
	if exists, err := util.PathExistsFS(fsys, stateFilename); err != nil {
		return nil, err
	} else if !exists {
		klog.InfoS("State file doesn't exists", "storage", fsys)
		// State is optional (file will not exist on when 1st restore happens).
		// Using pointer as optional.
		var state *state
		return state, nil
	}

	contents, err := fsys.ReadFile(stateFilename)
	if err != nil {
		return nil, err
	}

	s := &state{storageFS: fsys}
	if err := json.Unmarshal(contents, s); err != nil {
		return nil, err
	}

	return s, nil
}
