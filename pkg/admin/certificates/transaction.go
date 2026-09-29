// Package certificates provides recoverable, offline replacement of MicroShift PKI.
package certificates

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

const transactionDirectory = ".cert-renewal"

// RecoveryError means originals remain in the transaction directory and a
// later recovery is required before the data can be used.
type RecoveryError struct{ Err error }

func (e *RecoveryError) Error() string { return e.Err.Error() }
func (e *RecoveryError) Unwrap() error { return e.Err }

// Transaction replaces the PKI and generated resources together. Callers must
// hold the certificate write lock and keep MicroShift stopped until completion.
type Transaction struct {
	DataDir string
	// rename is injectable for commit/rollback fault-injection tests.
	rename func(string, string) error
}

func (t *Transaction) directory() string { return filepath.Join(t.DataDir, transactionDirectory) }
func (t *Transaction) Stage() string     { return filepath.Join(t.directory(), "stage") }
func (t *Transaction) marker() string    { return filepath.Join(t.directory(), "committing") }

func (t *Transaction) Pending() (bool, error) {
	if err := t.validatePaths(); err != nil {
		return false, err
	}
	info, err := os.Lstat(t.marker())
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !info.Mode().IsRegular() {
		return false, fmt.Errorf("certificate transaction marker must be a regular file")
	}
	return true, nil
}

func (t *Transaction) validatePaths() error {
	if !filepath.IsAbs(t.DataDir) || filepath.Clean(t.DataDir) != t.DataDir || t.DataDir == string(filepath.Separator) {
		return fmt.Errorf("certificate transaction requires an absolute data directory, not the filesystem root")
	}
	for _, path := range []string{t.DataDir, t.directory(), filepath.Join(t.directory(), "previous")} {
		info, err := os.Lstat(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return fmt.Errorf("certificate transaction path %q must be a directory, not a symlink", path)
		}
	}
	return nil
}

// Prepare copies only PKI and generated resources, never the etcd database.
// The original trees remain untouched until Commit. cp preserves ownership,
// permissions, timestamps, and SELinux labels, matching existing backup tooling.
func (t *Transaction) Prepare() error {
	pending, err := t.Pending()
	if err != nil {
		return err
	}
	if pending {
		return fmt.Errorf("an interrupted certificate renewal must be recovered first")
	}
	if err := t.Discard(); err != nil {
		return err
	}
	if err := os.MkdirAll(t.Stage(), 0700); err != nil {
		return err
	}
	for _, name := range []string{"certs", "resources"} {
		source := filepath.Join(t.DataDir, name)
		if err := regularTree(source); err != nil {
			return err
		}
		if err := exec.Command("cp", "--archive", "--reflink=auto", "--", source, filepath.Join(t.Stage(), name)).Run(); err != nil {
			return fmt.Errorf("failed to stage %s: %w", name, err)
		}
	}
	return nil
}

// Commit keeps the old trees until post-commit validation succeeds. A durable
// marker is written before the first rename. Recovery is idempotent, including
// when interrupted between directory renames or during rollback itself.
func (t *Transaction) Commit(validate func() error) error {
	if err := t.validatePaths(); err != nil {
		return err
	}
	if err := syncTree(t.Stage()); err != nil {
		return err
	}
	for _, name := range []string{"certs", "resources"} {
		if err := syncTree(filepath.Join(t.DataDir, name)); err != nil {
			return err
		}
	}
	previous := filepath.Join(t.directory(), "previous")
	if err := os.Mkdir(previous, 0700); err != nil {
		return err
	}
	marker, err := os.OpenFile(t.marker(), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	err = marker.Sync()
	err = errors.Join(err, marker.Close())
	if err != nil {
		return err
	}
	if err := syncDirectory(t.directory()); err != nil {
		return err
	}
	if err := syncDirectory(t.DataDir); err != nil {
		return err
	}
	if err := t.replace(previous); err != nil {
		return t.rollback(err)
	}
	if err := validate(); err != nil {
		return t.rollback(err)
	}
	return t.finish()
}

func (t *Transaction) rollback(cause error) error {
	if err := t.Recover(); err != nil {
		return &RecoveryError{errors.Join(cause, err)}
	}
	return cause
}

func (t *Transaction) replace(previous string) error {
	for _, name := range []string{"certs", "resources"} {
		active := filepath.Join(t.DataDir, name)
		if err := t.move(active, filepath.Join(previous, name)); err != nil {
			return err
		}
		if err := syncDirectory(previous); err != nil {
			return err
		}
		if err := syncDirectory(t.DataDir); err != nil {
			return err
		}
		if err := t.move(filepath.Join(t.Stage(), name), active); err != nil {
			return err
		}
		if err := syncDirectory(t.Stage()); err != nil {
			return err
		}
		if err := syncDirectory(t.DataDir); err != nil {
			return err
		}
	}
	return nil
}

func (t *Transaction) move(from, to string) error {
	if t.rename != nil {
		return t.rename(from, to)
	}
	return os.Rename(from, to)
}

// Recover restores the original material. No certificate parsing or current
// configuration is needed to roll back a transaction.
func (t *Transaction) Recover() error {
	pending, err := t.Pending()
	if err != nil || !pending {
		return err
	}
	previous := filepath.Join(t.directory(), "previous")
	for _, name := range []string{"certs", "resources"} {
		backup := filepath.Join(previous, name)
		if _, err := os.Lstat(backup); errors.Is(err, os.ErrNotExist) {
			if err := regularTree(filepath.Join(t.DataDir, name)); err != nil {
				return fmt.Errorf("neither original nor recoverable %s tree is available: %w", name, err)
			}
			continue // Not replaced yet, or already restored by an earlier recovery.
		} else if err != nil {
			return err
		}
		if err := regularTree(backup); err != nil {
			return err
		}
		active := filepath.Join(t.DataDir, name)
		if err := os.RemoveAll(active); err != nil {
			return err
		}
		if err := t.move(backup, active); err != nil {
			return err
		}
		if err := syncDirectory(t.DataDir); err != nil {
			return err
		}
		if err := syncDirectory(previous); err != nil {
			return err
		}
	}
	return t.finish()
}

func (t *Transaction) finish() error {
	if err := os.Remove(t.marker()); err != nil {
		return err
	}
	if err := syncDirectory(t.directory()); err != nil {
		return err
	}
	return t.Discard()
}

// Discard removes staging only when no committed replacement needs recovery.
func (t *Transaction) Discard() error {
	pending, err := t.Pending()
	if err != nil {
		return err
	}
	if pending {
		return nil // Never delete the only recoverable copy of active material.
	}
	return os.RemoveAll(t.directory())
}

func regularTree(root string) error {
	info, err := os.Lstat(root)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("certificate renewal requires a directory at %q", root)
	}
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && !entry.Type().IsRegular() {
			return fmt.Errorf("certificate renewal cannot replace non-regular path %q", path)
		}
		return nil
	})
}

func syncTree(root string) error {
	if err := regularTree(root); err != nil {
		return err
	}
	return filepath.WalkDir(root, func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		return syncDirectory(path)
	})
}

func syncDirectory(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	return file.Sync()
}
