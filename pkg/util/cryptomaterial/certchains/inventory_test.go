package certchains

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCertificateChains_Inventory(t *testing.T) {
	chains := testChains(t, t.TempDir())

	type expectedEntry struct {
		path string
		role CertificateRole
	}
	expected := []expectedEntry{
		{"test-signer1", CertificateRoleCA},
		{"test-signer1/test-signer1-subca", CertificateRoleCA},
		{"test-signer1/test-signer1-subca/test-signer1-subca-too", CertificateRoleCA},
		{"test-signer1/test-signer1-subca/test-signer1-subca-too/test-signer1-subca-too-too", CertificateRoleCA},
		{"test-signer1/test-signer1-subca/test-signer1-subca-too/test-signer1-subca-too-too/subca-too-too-test-client2", CertificateRoleClient},
		{"test-signer1/test-signer1-subca/test-signer1-subca-too/test-signer1-subca-too-too2", CertificateRoleCA},
		{"test-signer1/test-signer1-subca/test-signer1-subca-too/subca-too-test-client1", CertificateRoleClient},
		{"test-signer1/test-signer1-subca/test-signer1-subca-server1", CertificateRoleServing},
		{"test-signer1/test-client1", CertificateRoleClient},
		{"test-signer1/test-client2", CertificateRoleClient},
		{"test-signer1/test-signer1-server1", CertificateRoleServing},
		{"test-signer2", CertificateRoleCA},
		{"test-signer2/test-signer2-server1", CertificateRoleServing},
		{"test-signer3", CertificateRoleCA},
		{"test-signer3/test-signer3-subca1", CertificateRoleCA},
		{"test-signer3/test-signer3-subca1/test-client1", CertificateRoleClient},
		{"test-signer3/test-peer1", CertificateRolePeer},
		{"test-signer3/test-signer3-server1", CertificateRoleServing},
	}

	inventory := chains.Inventory()
	require.Len(t, inventory, len(expected))
	for index, want := range expected {
		entry := inventory[index]
		require.Equal(t, want.path, strings.Join(entry.Path, "/"))
		require.Equal(t, want.role, entry.Role)
		require.False(t, entry.Certificate.NotAfter.IsZero())

		if entry.Role == CertificateRoleCA {
			require.NotNil(t, chains.GetSigner(entry.Path...))
			continue
		}
		_, _, err := chains.GetCertKey(entry.Path...)
		require.NoError(t, err)
	}
}

func TestCertificateInventory_ByRole(t *testing.T) {
	inventory := testChains(t, t.TempDir()).Inventory()

	tests := []struct {
		name  string
		role  CertificateRole
		paths []string
	}{
		{
			name: "CAs",
			role: CertificateRoleCA,
			paths: []string{
				"test-signer1",
				"test-signer1/test-signer1-subca",
				"test-signer1/test-signer1-subca/test-signer1-subca-too",
				"test-signer1/test-signer1-subca/test-signer1-subca-too/test-signer1-subca-too-too",
				"test-signer1/test-signer1-subca/test-signer1-subca-too/test-signer1-subca-too-too2",
				"test-signer2",
				"test-signer3",
				"test-signer3/test-signer3-subca1",
			},
		},
		{
			name: "serving certificates",
			role: CertificateRoleServing,
			paths: []string{
				"test-signer1/test-signer1-subca/test-signer1-subca-server1",
				"test-signer1/test-signer1-server1",
				"test-signer2/test-signer2-server1",
				"test-signer3/test-signer3-server1",
			},
		},
		{
			name:  "unknown role",
			role:  CertificateRoleUnknown,
			paths: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries := inventory.ByRole(tt.role)
			paths := make([]string, 0, len(entries))
			for _, entry := range entries {
				paths = append(paths, strings.Join(entry.Path, "/"))
			}
			require.Equal(t, tt.paths, paths)
		})
	}
}
