package certchains

import (
	"crypto/x509"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apiserver/pkg/authentication/user"

	"github.com/openshift/microshift/pkg/util/cryptomaterial"
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
		require.Equal(t, entry.Path[len(entry.Path)-1], entry.Name)
		require.Equal(t, want.role, entry.Role)
		require.False(t, entry.Certificate.NotAfter.IsZero())
		if len(entry.Path) > 1 {
			require.Equal(t, entry.Path[len(entry.Path)-2], entry.ParentCA)
		} else {
			require.Empty(t, entry.ParentCA)
		}

		if entry.Role == CertificateRoleCA {
			require.Equal(t, RotationPolicyExtended, entry.RotationPolicy)
			require.NotNil(t, chains.GetSigner(entry.Path...))
			continue
		}
		if entry.Role == CertificateRoleServing {
			require.Equal(t, RotationPolicyStandard, entry.RotationPolicy)
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

func TestCertificateInventoryEntry_ZoneAt(t *testing.T) {
	now := time.Date(2026, time.September, 8, 10, 30, 0, 0, time.UTC)
	tests := []struct {
		name      string
		policy    RotationPolicy
		notBefore time.Time
		notAfter  time.Time
		want      CertificateZone
		wantError bool
	}{
		{
			name:      "standard green",
			policy:    RotationPolicyStandard,
			notBefore: now.Add(-4169 * time.Hour),
			notAfter:  now.Add(5831 * time.Hour),
			want:      CertificateZoneGreen,
		},
		{
			name:      "standard green boundary is yellow",
			policy:    RotationPolicyStandard,
			notBefore: now.Add(-417 * time.Hour),
			notAfter:  now.Add(583 * time.Hour),
			want:      CertificateZoneYellow,
		},
		{
			name:      "standard yellow boundary is red",
			policy:    RotationPolicyStandard,
			notBefore: now.Add(-667 * time.Hour),
			notAfter:  now.Add(333 * time.Hour),
			want:      CertificateZoneRed,
		},
		{
			name:      "extended green boundary is yellow",
			policy:    RotationPolicyExtended,
			notBefore: now.Add(-850 * time.Hour),
			notAfter:  now.Add(150 * time.Hour),
			want:      CertificateZoneYellow,
		},
		{
			name:      "extended yellow boundary is red",
			policy:    RotationPolicyExtended,
			notBefore: now.Add(-900 * time.Hour),
			notAfter:  now.Add(100 * time.Hour),
			want:      CertificateZoneRed,
		},
		{
			name:      "not yet valid",
			policy:    RotationPolicyStandard,
			notBefore: now.Add(time.Hour),
			notAfter:  now.Add(1000 * time.Hour),
			want:      CertificateZoneRed,
		},
		{
			name:      "expired",
			policy:    RotationPolicyStandard,
			notBefore: now.Add(-1000 * time.Hour),
			notAfter:  now,
			want:      CertificateZoneRed,
		},
		{
			name:      "unknown policy",
			policy:    RotationPolicyUnknown,
			notBefore: now.Add(-500 * time.Hour),
			notAfter:  now.Add(500 * time.Hour),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := CertificateInventoryEntry{
				Name:           "test-certificate",
				RotationPolicy: tt.policy,
				Certificate: x509.Certificate{
					NotBefore: tt.notBefore,
					NotAfter:  tt.notAfter,
				},
			}
			got, err := entry.ZoneAt(now)
			if tt.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCertificateChainsBuilder_LoadInventory(t *testing.T) {
	tmpDir := t.TempDir()
	builder := NewCertificateChains(
		NewCertificateSigner("test-signer", filepath.Join(tmpDir, "test-signer"), 10*365*24*time.Hour).
			WithService("test-service").
			WithClientCertificates(&ClientCertificateSigningRequestInfo{
				CSRMeta: CSRMeta{
					Name:           "test-client",
					Service:        "test-client-service",
					Validity:       365 * 24 * time.Hour,
					RotationPolicy: RotationPolicyStandard,
				},
				UserInfo: &user.DefaultInfo{Name: "test-user"},
			}),
	)
	chains, err := builder.Complete()
	require.NoError(t, err)

	inventory, err := builder.LoadInventory()
	require.NoError(t, err)
	require.Equal(t, chains.Inventory(), inventory)
	require.Equal(t, "test-service", inventory[0].Service)
	require.Equal(t, RotationPolicyExtended, inventory[0].RotationPolicy)
	require.Equal(t, "test-client-service", inventory[1].Service)
	require.Equal(t, RotationPolicyStandard, inventory[1].RotationPolicy)

	clientCertPath := cryptomaterial.ClientCertPath(filepath.Join(tmpDir, "test-signer", "test-client"))
	require.NoError(t, os.Remove(clientCertPath))
	_, err = builder.LoadInventory()
	require.Error(t, err)
	_, statErr := os.Stat(clientCertPath)
	require.ErrorIs(t, statErr, os.ErrNotExist)
}
