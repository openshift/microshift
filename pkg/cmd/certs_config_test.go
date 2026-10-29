package cmd

import (
	"bytes"
	"context"
	"testing"
	"time"

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
			now := time.Now()
			nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
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
				require.WithinDuration(t, nextMidnight.Add(expected), entry.Certificate.NotAfter, time.Minute, entry.Name)
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

func TestCertificateRotationRestartPolicy(t *testing.T) {
	for _, forceRestart := range []bool{true, false} {
		t.Run(map[bool]string{true: "enabled", false: "disabled"}[forceRestart], func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			watchCertificateRotation(ctx, cancel, time.Now().Add(-time.Second), forceRestart)
			if forceRestart {
				require.ErrorIs(t, ctx.Err(), context.Canceled)
			} else {
				require.NoError(t, ctx.Err(), "critical expiry must not stop a warn-only service")
			}
		})
	}
}

func TestCertificateRotationWatcherStopsWithService(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	watchCertificateRotation(ctx, func() { t.Fatal("service cancellation must not trigger certificate rotation") }, time.Now().Add(time.Hour), true)
}
