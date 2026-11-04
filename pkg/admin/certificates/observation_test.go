package certificates_test

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/openshift/microshift/pkg/admin/certificates"
	certificatesv1alpha1 "github.com/openshift/microshift/pkg/apis/certificates/v1alpha1"
	"github.com/openshift/microshift/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestReadOnlyCertificateObservation(t *testing.T) {
	cfg := &config.Config{
		Network:   config.Network{ServiceNetwork: []string{"10.43.0.0/16"}},
		Node:      config.Node{HostnameOverride: "test-node", NodeIP: "192.0.2.10"},
		DNS:       config.DNS{BaseDomain: "example.test"},
		ApiServer: config.ApiServer{AdvertiseAddress: "10.44.0.0"},
	}
	dataDir, lockPath := t.TempDir(), filepath.Join(t.TempDir(), "certs.lock")
	now := time.Now().UTC().Truncate(time.Second)
	observation := certificates.Observe(cfg, dataDir, lockPath, now)
	require.Error(t, observation.InventoryError)
	require.Error(t, observation.PendingError)
	require.NoFileExists(t, lockPath)

	lock, err := certificates.Lock(lockPath, true)
	require.NoError(t, err)
	require.NoError(t, lock.Close())
	observation = certificates.Observe(cfg, dataDir, lockPath, now)
	require.Error(t, observation.InventoryError) // PKI has not been initialized.
	require.NoError(t, observation.PendingError)
	require.NoDirExists(t, filepath.Join(dataDir, "certs"))

	builder, err := certificates.NewChains(cfg, dataDir)
	require.NoError(t, err)
	chains, err := builder.Complete()
	require.NoError(t, err)
	// Unparseable private keys must not affect public certificate observations.
	require.NoError(t, filepath.WalkDir(dataDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(entry.Name(), ".key") {
			return os.WriteFile(path, []byte("not a private key"), 0600)
		}
		return nil
	}))
	before, err := certificates.DigestTree(dataDir)
	require.NoError(t, err)
	observation = certificates.Observe(cfg, dataDir, lockPath, now)
	require.NoError(t, observation.InventoryError)
	require.NoError(t, observation.PendingError)
	require.False(t, observation.RenewalPending)
	require.Equal(t, now, observation.ObservedAt)
	require.Equal(t, chains.Inventory(), observation.Inventory)
	after, err := certificates.DigestTree(dataDir)
	require.NoError(t, err)
	require.Equal(t, before, after)

	txDir := filepath.Join(dataDir, ".cert-renewal")
	require.NoError(t, os.Mkdir(txDir, 0700))
	metadata := certificates.PendingRenewal{
		Version: 1, ConfigHash: "config", OriginalHash: "active", RenewedHash: "pending",
		Result: &certificatesv1alpha1.CertificateRenewalResult{
			Status: certificatesv1alpha1.RenewalStatusPending, Mode: certificatesv1alpha1.RenewalModeServing,
		},
	}
	for _, version := range []int{1, 2} {
		metadata.Version = version
		record, err := json.Marshal(map[string]any{"directory": "stage-test", "metadata": metadata})
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(txDir, "pending.json"), record, 0600))
		observation = certificates.Observe(cfg, dataDir, lockPath, now)
		require.NoError(t, observation.InventoryError)
		require.Equal(t, chains.Inventory(), observation.Inventory)
		if version == 1 {
			require.NoError(t, observation.PendingError)
			require.True(t, observation.RenewalPending) // Metadata only; the stage is not read.
		} else {
			require.Error(t, observation.PendingError)
		}
	}
	require.NoError(t, os.WriteFile(filepath.Join(txDir, "pending.json"), []byte("{"), 0600))
	observation = certificates.Observe(cfg, dataDir, lockPath, now)
	require.NoError(t, observation.InventoryError)
	require.ErrorIs(t, observation.PendingError, certificates.ErrInvalidPendingMetadata)
	require.FileExists(t, filepath.Join(txDir, "pending.json"))
	require.NoDirExists(t, filepath.Join(dataDir, ".cert-renewal-rejected"))

	require.NoError(t, os.WriteFile(filepath.Join(txDir, "committing"), nil, 0600))
	observation = certificates.Observe(cfg, dataDir, lockPath, now)
	require.Error(t, observation.InventoryError)
	require.Error(t, observation.PendingError)
	require.FileExists(t, filepath.Join(txDir, "committing")) // No recovery.

	lock, err = certificates.Lock(lockPath, true)
	require.NoError(t, err)
	defer lock.Close()
	observation = certificates.Observe(cfg, dataDir, lockPath, now)
	require.ErrorIs(t, observation.InventoryError, certificates.ErrBusy)
	require.ErrorIs(t, observation.PendingError, certificates.ErrBusy)
}
