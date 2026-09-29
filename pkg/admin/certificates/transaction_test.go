//nolint:testpackage // Inject failures at individual commit operations.
package certificates

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func newTestTransaction(t *testing.T) *Transaction {
	t.Helper()
	tx := &Transaction{DataDir: t.TempDir()}
	for _, name := range []string{"certs", "resources"} {
		require.NoError(t, os.Mkdir(filepath.Join(tx.DataDir, name), 0700))
		require.NoError(t, os.WriteFile(filepath.Join(tx.DataDir, name, "original"), []byte("old"), 0600))
	}
	require.NoError(t, tx.Prepare())
	for _, name := range []string{"certs", "resources"} {
		require.NoError(t, os.WriteFile(filepath.Join(tx.Stage(), name, "original"), []byte("new"), 0600))
	}
	return tx
}

func assertTransactionContents(t *testing.T, tx *Transaction, want string) {
	t.Helper()
	for _, name := range []string{"certs", "resources"} {
		contents, err := os.ReadFile(filepath.Join(tx.DataDir, name, "original"))
		require.NoError(t, err)
		require.Equal(t, want, string(contents))
		info, err := os.Stat(filepath.Join(tx.DataDir, name, "original"))
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0600), info.Mode().Perm())
	}
}

func TestTransactionCommit(t *testing.T) {
	tx := newTestTransaction(t)
	assertTransactionContents(t, tx, "old")
	require.NoError(t, tx.Commit(func() error {
		assertTransactionContents(t, tx, "new")
		return nil
	}))
	assertTransactionContents(t, tx, "new")
	_, err := os.Stat(tx.directory())
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestTransactionRollback(t *testing.T) {
	for failAt := 1; failAt <= 5; failAt++ {
		t.Run(fmt.Sprint(failAt), func(t *testing.T) {
			tx := newTestTransaction(t)
			failure := errors.New("injected failure")
			calls := 0
			tx.rename = func(from, to string) error {
				calls++
				if calls == failAt && failAt <= 4 {
					return failure
				}
				return os.Rename(from, to)
			}
			err := tx.Commit(func() error { return failure })
			require.ErrorIs(t, err, failure)
			assertTransactionContents(t, tx, "old")
			require.NoError(t, tx.Recover())
		})
	}
}

func TestTransactionFailedRollbackIsRecoverable(t *testing.T) {
	tx := newTestTransaction(t)
	calls := 0
	tx.rename = func(from, to string) error {
		calls++
		if calls == 5 {
			return errors.New("injected rollback failure")
		}
		return os.Rename(from, to)
	}
	var recoveryError *RecoveryError
	require.ErrorAs(t, tx.Commit(func() error { return errors.New("force rollback") }), &recoveryError)
	require.NoError(t, tx.Discard())
	require.FileExists(t, filepath.Join(tx.directory(), "previous", "certs", "original"))
	restarted := &Transaction{DataDir: tx.DataDir}
	require.NoError(t, restarted.Recover())
	assertTransactionContents(t, restarted, "old")
}

func TestTransactionInterruptedCommit(t *testing.T) {
	// Four commit moves followed by the two rollback moves.
	for interruptAt := 1; interruptAt <= 6; interruptAt++ {
		t.Run(fmt.Sprint(interruptAt), func(t *testing.T) {
			tx := newTestTransaction(t)
			calls := 0
			tx.rename = func(from, to string) error {
				require.NoError(t, os.Rename(from, to))
				calls++
				if calls == interruptAt {
					panic("simulated process interruption")
				}
				return nil
			}
			require.Panics(t, func() { _ = tx.Commit(func() error { return errors.New("force rollback") }) })
			require.NoError(t, tx.Discard(), "cleanup must retain a pending backup")
			pending, err := tx.Pending()
			require.NoError(t, err)
			require.True(t, pending)
			restarted := &Transaction{DataDir: tx.DataDir}
			require.NoError(t, restarted.Recover())
			require.NoError(t, restarted.Recover())
			assertTransactionContents(t, restarted, "old")
		})
	}
}

func TestCertificateLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "certificate.lock")
	reader, err := Lock(path, false)
	require.NoError(t, err)
	defer func() { _ = reader.Close() }()
	otherReader, err := Lock(path, false)
	require.NoError(t, err)
	require.NoError(t, otherReader.Close())
	_, err = Lock(path, true)
	require.ErrorIs(t, err, ErrBusy)
	require.NoError(t, reader.Close())
	writer, err := Lock(path, true)
	require.NoError(t, err)
	defer func() { _ = writer.Close() }()
	_, err = Lock(path, false)
	require.ErrorIs(t, err, ErrBusy)
	_, err = Lock(path, true)
	require.ErrorIs(t, err, ErrBusy)
	require.NoError(t, ShareLock(writer))
	reader, err = Lock(path, false)
	require.NoError(t, err)
	require.NoError(t, reader.Close())
}

func TestCertificateLockCreatesDirectory(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "microshift-backups")
	path := filepath.Join(directory, "certs.lock")
	lock, err := Lock(path, true)
	require.NoError(t, err)
	t.Cleanup(func() { _ = lock.Close() })
	info, err := os.Stat(directory)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0700), info.Mode().Perm())
	original, err := lock.Stat()
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0600), original.Mode().Perm())
	require.NoError(t, lock.Close())

	// A persistent file is not a stale lock: closing releases it without unlinking.
	reopened, err := Lock(path, true)
	require.NoError(t, err)
	defer func() { _ = reopened.Close() }()
	info, err = reopened.Stat()
	require.NoError(t, err)
	require.True(t, os.SameFile(original, info))
}

func TestCertificateLockRejectsInvalidDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "not-a-directory")
	require.NoError(t, os.WriteFile(path, nil, 0600))
	lock, err := Lock(filepath.Join(path, "certs.lock"), true)
	require.ErrorContains(t, err, "cannot create certificate lock directory")
	require.Nil(t, lock)
}

func TestTransactionRejectsUnsafePaths(t *testing.T) {
	for _, path := range []string{"", ".", "/", "relative/data"} {
		tx := &Transaction{DataDir: path}
		require.Error(t, tx.Discard())
		require.Error(t, tx.Prepare())
	}
	dataDir, outside := t.TempDir(), t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(outside, "keep"), []byte("unchanged"), 0600))
	tx := &Transaction{DataDir: dataDir}
	require.NoError(t, os.Symlink(outside, tx.directory()))
	require.Error(t, tx.Prepare())
	require.Error(t, tx.Discard())
	require.FileExists(t, filepath.Join(outside, "keep"))
}

func TestTransactionRejectsSymlinkMaterial(t *testing.T) {
	tx := newTestTransaction(t)
	require.NoError(t, tx.Discard())
	require.NoError(t, os.Symlink(t.TempDir(), filepath.Join(tx.DataDir, "certs", "link")))
	require.ErrorContains(t, tx.Prepare(), "non-regular")
	assertTransactionContents(t, tx, "old")
}

func TestTransactionKeepsUnrecoverableMarker(t *testing.T) {
	tx := newTestTransaction(t)
	require.NoError(t, os.WriteFile(tx.marker(), nil, 0600))
	require.NoError(t, os.Rename(filepath.Join(tx.DataDir, "certs"), filepath.Join(tx.DataDir, "missing-certs")))
	require.Error(t, tx.Recover())
	require.NoError(t, tx.Discard())
	pending, err := tx.Pending()
	require.NoError(t, err)
	require.True(t, pending, "incomplete recovery must continue blocking startup")
}
