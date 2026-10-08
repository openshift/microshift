package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/openshift/microshift/pkg/admin/certificates"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
	"github.com/openshift/microshift/pkg/util/cryptomaterial/certchains"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func TestConfiguredCertificateIssuance(t *testing.T) {
	for _, policy := range []config.Certificates{{}, {
		ServingValidity:                  &metav1.Duration{Duration: 1008 * time.Hour},
		CAValidity:                       &metav1.Duration{Duration: 17520 * time.Hour},
		ForceRestartOnExpirationImminent: ptr.To(false),
	}} {
		t.Run(policy.ServingDuration().String(), func(t *testing.T) {
			cfg := &config.Config{
				Certificates: policy,
				Network:      config.Network{ServiceNetwork: []string{"10.43.0.0/16"}},
				Node:         config.Node{HostnameOverride: "test-node", NodeIP: "192.0.2.10"},
				DNS:          config.DNS{BaseDomain: "example.test"},
				ApiServer:    config.ApiServer{AdvertiseAddress: "10.44.0.0"},
			}
			dir := t.TempDir()
			builder, err := certificateChainsSetup(cfg, dir)
			require.NoError(t, err)
			chains, err := builder.Complete()
			require.NoError(t, err)
			inventory, err := builder.ValidateRenewal(true, time.Now())
			require.NoError(t, err, "initial issuance must bound every descendant by its signing chain")
			for _, entry := range inventory {
				expected := cryptomaterial.ShortLivedCertificateValidity
				switch entry.Role {
				case certchains.CertificateRoleCA:
					expected = policy.CADuration()
				case certchains.CertificateRoleServing:
					expected = policy.ServingDuration()
				case certchains.CertificateRoleClient, certchains.CertificateRolePeer:
					if entry.RotationPolicy == certchains.RotationPolicyExtended {
						expected = cryptomaterial.LongLivedCertificateValidity
					}
					expected = min(expected, policy.CADuration())
				case certchains.CertificateRoleUnknown:
					t.Fatalf("unknown role for %s", entry.Name)
				}
				require.InDelta(t, expected.Seconds(), entry.Certificate.NotAfter.Sub(entry.Certificate.NotBefore).Seconds(), (24*time.Hour + time.Minute).Seconds(), entry.Name)
			}
			paths, err := certsToRegenerate(chains.Inventory(), time.Now())
			require.NoError(t, err)
			require.Empty(t, paths, "new custom lifetimes must not immediately trigger startup renewal")
			// A later startup must reuse healthy certificates despite changed policy.
			cfg.Certificates.ServingValidity = &metav1.Duration{Duration: 336 * time.Hour}
			cfg.Certificates.CAValidity = &metav1.Duration{Duration: 8760 * time.Hour}
			builder, err = certificateChainsSetup(cfg, dir)
			require.NoError(t, err)
			loaded, err := builder.Complete()
			require.NoError(t, err)
			after := loaded.Inventory()
			for i, entry := range chains.Inventory() {
				// Check the configurable roles; clients keep their existing
				// EnsureClientCertificate behavior, including identity repair.
				if entry.Role == certchains.CertificateRoleCA || entry.Role == certchains.CertificateRoleServing {
					require.True(t, bytes.Equal(entry.Certificate.Raw, after[i].Certificate.Raw), entry.Name)
				}
			}
		})
	}
}

func TestConfiguredCertificateRenewal(t *testing.T) {
	for _, renewCAs := range []bool{false, true} {
		t.Run(fmt.Sprintf("renewCAs=%t", renewCAs), func(t *testing.T) {
			cfg, dir := pendingRenewalFixture(t)
			before := activeCertificateFileDigests(t, dir)
			cfg.Certificates = config.Certificates{
				ServingValidity:                  &metav1.Duration{Duration: 1008 * time.Hour},
				CAValidity:                       &metav1.Duration{Duration: 17520 * time.Hour},
				ForceRestartOnExpirationImminent: ptr.To(false),
			}
			builder, err := certificateChainsSetup(cfg, dir)
			require.NoError(t, err)
			// A config change alone must not rewrite existing certificates.
			chains, err := builder.Load()
			require.NoError(t, err)
			require.Equal(t, before, activeCertificateFileDigests(t, dir))
			paths, err := certsToRegenerate(chains.Inventory(), time.Now())
			require.NoError(t, err)
			require.Empty(t, paths)
			plan, err := builder.PlanRenewal(renewCAs, time.Now())
			require.NoError(t, err)
			renewed, err := applyCertificateRenewal(cfg, dir, renewCAs)
			require.NoError(t, err)
			require.Equal(t, before, activeCertificateFileDigests(t, dir), "online preparation must preserve active PKI")
			for _, planned := range plan {
				for _, entry := range renewed {
					if strings.Join(planned.Path, "/") == strings.Join(entry.Path, "/") {
						require.WithinDuration(t, planned.NewNotAfter, entry.Certificate.NotAfter, 5*time.Second)
					}
				}
			}
			activated, err := activateCertificateRenewal(cfg, &certificates.Transaction{DataDir: dir}, func() error { return nil }, time.Now())
			require.NoError(t, err)
			require.True(t, activated)
			inventory, err := builder.LoadInventory()
			require.NoError(t, err)
			parents := map[string]time.Time{}
			for _, entry := range inventory {
				cert := entry.Certificate
				parents[strings.Join(entry.Path, "/")] = cert.NotAfter
				if len(entry.Path) > 1 {
					require.False(t, cert.NotAfter.After(parents[strings.Join(entry.Path[:len(entry.Path)-1], "/")]))
				}
				var expected time.Duration
				if entry.Role == certchains.CertificateRoleServing {
					expected = cfg.Certificates.ServingDuration()
				} else if renewCAs && entry.Role == certchains.CertificateRoleCA {
					expected = cfg.Certificates.CADuration()
				}
				if expected > 0 {
					require.InDelta(t, expected.Seconds(), cert.NotAfter.Sub(cert.NotBefore).Seconds(), (24*time.Hour + time.Minute).Seconds())
				}
				state, err := entry.StatusAt(time.Now())
				require.NoError(t, err)
				require.Equal(t, certchains.CertificateStatusHealthy, state)
			}
		})
	}
}

func TestCertificateStatusConfiguredPolicy(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	cfg := &config.Config{Certificates: config.Certificates{
		ServingValidity:                  &metav1.Duration{Duration: 90 * time.Minute},
		CAValidity:                       &metav1.Duration{Duration: 17520 * time.Hour},
		ForceRestartOnExpirationImminent: ptr.To(false),
	}}
	entry := certificateInventoryEntry("test", "old", certchains.CertificateRoleServing, certchains.RotationPolicyStandard, now.Add(-500*time.Hour), now.Add(500*time.Hour))
	status, err := newCertificateStatusList(certchains.CertificateInventory{entry}, cfg, now)
	require.NoError(t, err)
	require.False(t, status.Config.ForceRestartOnExpirationImminent)
	require.Equal(t, "1h30m0s", status.Config.ServingValidity)
	require.Equal(t, "17520h", status.Config.CAValidity)
	require.EqualValues(t, certchains.CertificateStatusExpiresSoon, status.Items[0].Status, "classify the encoded lifetime, not new configuration")
	var output bytes.Buffer
	require.NoError(t, writeCertificateStatusTable(&output, status))
	require.Contains(t, output.String(), "Force restart on expiration imminent: false")
}
