//nolint:testpackage // Verify quarantine paths and interrupted transaction state.
package certificates

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQuarantinePendingPreservesRejectedMaterial(t *testing.T) {
	tx := newTestTransaction(t)
	stage := filepath.Base(tx.Stage())
	corrupt := []byte(`{"directory": "credential-bearing-corrupt-json`)
	require.NoError(t, os.WriteFile(tx.readyPath(), corrupt, 0600))
	_, err := tx.LoadPending()
	require.ErrorIs(t, err, ErrInvalidPendingMetadata)
	require.NoError(t, tx.QuarantinePending())
	assertTransactionContents(t, tx, "old")
	require.NoDirExists(t, tx.directory())
	entries, err := os.ReadDir(tx.rejectedDirectory())
	require.NoError(t, err)
	require.Len(t, entries, 1)
	quarantine := filepath.Join(tx.rejectedDirectory(), entries[0].Name(), "renewal")
	contents, err := os.ReadFile(filepath.Join(quarantine, "pending.json"))
	require.NoError(t, err)
	require.Equal(t, corrupt, contents)
	require.FileExists(t, filepath.Join(quarantine, stage, "certs", "original"))
	info, err := os.Stat(tx.rejectedDirectory())
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0700), info.Mode().Perm())

	// Normal cleanup, failed preparation, and recovery do not clear the warning.
	require.NoError(t, tx.Discard())
	require.NoError(t, tx.Prepare())
	require.NoError(t, tx.Discard())
	require.NoError(t, tx.Recover())
	rejected, err := tx.HasRejected()
	require.NoError(t, err)
	require.True(t, rejected)

	// More than one rejected attempt is retained until successful activation.
	require.NoError(t, tx.Prepare())
	require.NoError(t, tx.Publish([]byte(`{}`)))
	require.NoError(t, os.WriteFile(tx.readyPath(), corrupt, 0600))
	require.NoError(t, tx.QuarantinePending())
	entries, err = os.ReadDir(tx.rejectedDirectory())
	require.NoError(t, err)
	require.Len(t, entries, 2)
	require.NoError(t, tx.ClearRejected())
	require.NoError(t, tx.ClearRejected())
	rejected, err = tx.HasRejected()
	require.NoError(t, err)
	require.False(t, rejected)
}

func TestQuarantinePendingPreservesInterruptedActivation(t *testing.T) {
	tx := newTestTransaction(t)
	tx.rename = func(from, to string) error {
		require.NoError(t, os.Rename(from, to))
		panic("simulated process interruption")
	}
	require.Panics(t, func() { _ = tx.Commit(func() error { return nil }) })
	require.NoError(t, os.WriteFile(tx.readyPath(), []byte("invalid json"), 0600))
	require.ErrorContains(t, tx.QuarantinePending(), "incomplete activation")
	require.NoDirExists(t, tx.rejectedDirectory())
	require.FileExists(t, filepath.Join(tx.directory(), "previous", "certs", "original"))
	restarted := &Transaction{DataDir: tx.DataDir}
	require.NoError(t, restarted.Recover())
	assertTransactionContents(t, restarted, "old")
}

func TestQuarantinePendingRejectsSymlink(t *testing.T) {
	tx := newTestTransaction(t)
	outside := t.TempDir()
	require.NoError(t, os.Symlink(outside, tx.rejectedDirectory()))
	require.Error(t, tx.QuarantinePending())
	_, err := tx.HasRejected()
	require.Error(t, err)
	require.Error(t, tx.ClearRejected())
	require.FileExists(t, tx.readyPath())
	entries, err := os.ReadDir(outside)
	require.NoError(t, err)
	require.Empty(t, entries)
}

func TestQuarantinePendingRenameFailure(t *testing.T) {
	tx := newTestTransaction(t)
	failure := errors.New("cannot quarantine")
	tx.rename = func(_, _ string) error { return failure }
	require.ErrorIs(t, tx.QuarantinePending(), failure)
	require.FileExists(t, tx.readyPath())
	assertTransactionContents(t, tx, "old")
	rejected, err := tx.HasRejected()
	require.NoError(t, err)
	require.False(t, rejected)
}
