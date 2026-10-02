//nolint:testpackage // Exercise publication and interrupted activation state.
package certificates

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPendingGenerationPublication(t *testing.T) {
	tx := newTestTransaction(t)
	originalStage := tx.Stage()
	require.NoError(t, tx.Discard())
	require.DirExists(t, originalStage, "published material survives CLI cleanup")
	assertTransactionContents(t, tx, "old")

	// A failed or interrupted second generation must preserve the first one.
	retry := &Transaction{DataDir: tx.DataDir}
	require.NoError(t, retry.Prepare())
	unpublished := retry.Stage()
	require.NotEqual(t, originalStage, unpublished)
	require.NoError(t, retry.Discard())
	require.NoDirExists(t, unpublished)
	metadata, err := retry.LoadPending()
	require.NoError(t, err)
	require.JSONEq(t, `{"validated":true}`, string(metadata))
	require.Equal(t, originalStage, retry.Stage())

	// A successful retry atomically publishes a new generation.
	require.NoError(t, retry.Prepare())
	replacement := retry.Stage()
	require.NoError(t, retry.Publish([]byte(`{"replacement":true}`)))
	require.NoError(t, retry.Discard())
	require.NoDirExists(t, originalStage)
	metadata, err = retry.LoadPending()
	require.NoError(t, err)
	require.JSONEq(t, `{"replacement":true}`, string(metadata))
	require.Equal(t, replacement, retry.Stage())
	assertTransactionContents(t, tx, "old")
}

func TestPendingGenerationRejectsUnsafeRecord(t *testing.T) {
	for _, record := range []string{
		`not json`,
		`{"directory":"../outside","metadata":{}}`,
		`{"directory":"/outside","metadata":{}}`,
		`{"directory":"stage-../../outside","metadata":{}}`,
	} {
		t.Run(record, func(t *testing.T) {
			tx := newTestTransaction(t)
			require.NoError(t, os.WriteFile(tx.readyPath(), []byte(record), 0600))
			_, err := tx.LoadPending()
			require.Error(t, err)
			require.Error(t, tx.Discard())
			assertTransactionContents(t, tx, "old")
		})
	}
	for _, target := range []string{"record", "generation"} {
		t.Run(target, func(t *testing.T) {
			tx := newTestTransaction(t)
			path := tx.readyPath()
			if target == "generation" {
				path = tx.Stage()
			}
			require.NoError(t, os.Rename(path, path+".saved"))
			require.NoError(t, os.Symlink(path+".saved", path))
			_, err := tx.LoadPending()
			require.Error(t, err)
			require.Error(t, tx.Discard())
		})
	}
}

func TestPendingGenerationRefreshesResources(t *testing.T) {
	tx := newTestTransaction(t)
	require.NoError(t, os.WriteFile(filepath.Join(tx.DataDir, "resources", "later"), []byte("latest"), 0600))
	require.NoError(t, tx.RefreshResources())
	contents, err := os.ReadFile(filepath.Join(tx.Stage(), "resources", "later"))
	require.NoError(t, err)
	require.Equal(t, "latest", string(contents))
}

func TestPendingActivationRetriesBeforeCommitMarker(t *testing.T) {
	tx := newTestTransaction(t)
	require.NoError(t, os.Mkdir(filepath.Join(tx.directory(), "previous"), 0700))
	restarted := &Transaction{DataDir: tx.DataDir}
	_, err := restarted.LoadPending()
	require.NoError(t, err)
	require.NoError(t, restarted.Commit(func() error { return nil }))
	assertTransactionContents(t, restarted, "new")
}
