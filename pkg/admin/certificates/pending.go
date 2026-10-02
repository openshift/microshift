package certificates

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type pendingGeneration struct {
	Directory string          `json:"directory"`
	Metadata  json.RawMessage `json:"metadata"`
}

func (t *Transaction) readyPath() string { return filepath.Join(t.directory(), "pending.json") }

// LoadPending selects the published generation without activating it. A missing
// generation returns nil. Callers must hold the certificate operation lock.
func (t *Transaction) LoadPending() ([]byte, error) {
	if err := t.validatePaths(); err != nil {
		return nil, err
	}
	info, err := os.Lstat(t.readyPath())
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("pending certificate record must be a regular file")
	}
	contents, err := os.ReadFile(t.readyPath())
	if err != nil {
		return nil, err
	}
	var pending pendingGeneration
	if err := json.Unmarshal(contents, &pending); err != nil {
		return nil, fmt.Errorf("invalid pending certificate record")
	}
	if !strings.HasPrefix(pending.Directory, "stage-") || filepath.Base(pending.Directory) != pending.Directory || len(pending.Metadata) == 0 {
		return nil, fmt.Errorf("invalid pending certificate generation")
	}
	t.stage = filepath.Join(t.directory(), pending.Directory)
	if err := regularTree(t.Stage()); err != nil {
		return nil, err
	}
	return pending.Metadata, nil
}

// Publish durably records a validated generation. Replacing the small pointer
// file is atomic: failed generation never destroys the previously pending set.
func (t *Transaction) Publish(metadata []byte) error {
	if err := t.validatePaths(); err != nil {
		return err
	}
	if err := syncTree(t.Stage()); err != nil {
		return err
	}
	contents, err := json.Marshal(pendingGeneration{filepath.Base(t.Stage()), metadata})
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(t.directory(), "pending-")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(file.Name()) }() //nolint:gosec // CreateTemp chooses the name inside the validated transaction directory.
	_, err = file.Write(contents)
	err = errors.Join(err, file.Sync(), file.Close())
	if err != nil {
		return err
	}
	// Persist the generation's directory entry before publishing its name.
	if err := syncDirectory(t.directory()); err != nil {
		return err
	}
	if err := os.Rename(file.Name(), t.readyPath()); err != nil { //nolint:gosec // CreateTemp chooses the source name inside the validated transaction directory.
		return err
	}
	if err := syncDirectory(t.directory()); err != nil {
		return err
	}
	return syncDirectory(t.DataDir)
}

func (t *Transaction) discardUnpublished() error {
	// LoadPending changes Stage, so preserve the caller's working generation.
	working := t.stage
	metadata, err := t.LoadPending()
	keep := filepath.Base(t.Stage())
	t.stage = working
	if err != nil {
		return err // Preserve everything if the published record cannot be read.
	}
	if metadata == nil {
		return os.RemoveAll(t.directory())
	}
	entries, err := os.ReadDir(t.directory())
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.Name() != keep && (strings.HasPrefix(entry.Name(), "stage-") || strings.HasPrefix(entry.Name(), "pending-")) {
			if err := os.RemoveAll(filepath.Join(t.directory(), entry.Name())); err != nil {
				return err
			}
		}
	}
	return syncDirectory(t.directory())
}

// RefreshResources is only used during activation, with the service stopped.
// Never replace current unrelated resources with the snapshot taken at renewal.
func (t *Transaction) RefreshResources() error {
	if err := regularTree(t.Stage()); err != nil {
		return err
	}
	source := filepath.Join(t.DataDir, "resources")
	if err := regularTree(source); err != nil {
		return err
	}
	destination := filepath.Join(t.Stage(), "resources")
	if err := os.RemoveAll(destination); err != nil {
		return err
	}
	return exec.Command("cp", "--archive", "--reflink=auto", "--", source, destination).Run()
}

// DigestTree binds a generation to exact file names and contents, including keys
// and bundles, without recording those contents in transaction metadata.
func DigestTree(root string) (string, error) {
	if err := regularTree(root); err != nil {
		return "", err
	}
	hash := sha256.New()
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		name, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(hash, "%s\x00%x\n", name, sha256.Sum256(contents))
		return nil
	})
	return fmt.Sprintf("%x", hash.Sum(nil)), err
}
