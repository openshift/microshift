package certificates

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func (t *Transaction) rejectedDirectory() string {
	return filepath.Join(t.DataDir, ".cert-renewal-rejected")
}

// QuarantinePending preserves a rejected preparation without parsing its
// untrusted metadata or touching active PKI. Callers hold the operation lock.
// The retained directory also records the rejection for subsequent status calls.
func (t *Transaction) QuarantinePending() error {
	committing, err := t.Pending()
	if err != nil {
		return err
	}
	if committing {
		return fmt.Errorf("cannot quarantine pending certificates during an incomplete activation")
	}
	if err := os.MkdirAll(t.rejectedDirectory(), 0700); err != nil {
		return err
	}
	destination, err := os.MkdirTemp(t.rejectedDirectory(), "generation-")
	if err != nil {
		return err
	}
	if err := t.move(t.directory(), filepath.Join(destination, "renewal")); err != nil {
		_ = os.Remove(destination) // Only remove the empty destination on failure.
		return err
	}
	t.stage = ""
	for _, path := range []string{destination, t.rejectedDirectory(), t.DataDir} {
		if err := syncDirectory(path); err != nil {
			return err
		}
	}
	return nil
}

// HasRejected reports a quarantined preparation without reading corrupt data.
// Ordinary staging cleanup must not clear this persistent warning.
func (t *Transaction) HasRejected() (bool, error) {
	if err := t.validatePaths(); err != nil {
		return false, err
	}
	entries, err := os.ReadDir(t.rejectedDirectory())
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return len(entries) > 0, err
}

// ClearRejected removes rejected preparations only after a successful activation,
// never after publishing a replacement or recovering an interrupted activation.
func (t *Transaction) ClearRejected() error {
	if err := t.validatePaths(); err != nil {
		return err
	}
	if _, err := os.Lstat(t.rejectedDirectory()); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	if err := os.RemoveAll(t.rejectedDirectory()); err != nil {
		return err
	}
	return syncDirectory(t.DataDir)
}
