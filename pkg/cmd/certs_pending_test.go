package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/openshift/microshift/pkg/admin/certificates"
	certificatesv1alpha1 "github.com/openshift/microshift/pkg/apis/certificates/v1alpha1"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util/cryptomaterial/certchains"
)

func pendingRenewalFixture(t *testing.T) (*config.Config, string) {
	t.Helper()
	cfg := &config.Config{
		Network:   config.Network{ServiceNetwork: []string{"10.43.0.0/16"}},
		Node:      config.Node{HostnameOverride: "test-node", NodeIP: "192.0.2.10"},
		DNS:       config.DNS{BaseDomain: "example.test"},
		ApiServer: config.ApiServer{AdvertiseAddress: "10.44.0.0", URL: "https://127.0.0.1:6443", Port: 6443},
	}
	dataDir := t.TempDir()
	builder, err := certificateChainsSetup(cfg, dataDir)
	require.NoError(t, err)
	chains, err := builder.Complete()
	require.NoError(t, err)
	require.NoError(t, initKubeconfigs(cfg, chains, dataDir))
	_, err = applyCertificateRenewal(cfg, dataDir, false)
	require.NoError(t, err)
	return cfg, dataDir
}

func TestPendingRenewalChangedInputs(t *testing.T) {
	for _, changed := range []string{"configuration", "active", "pending"} {
		t.Run(changed, func(t *testing.T) {
			cfg, dir := pendingRenewalFixture(t)
			tx := &certificates.Transaction{DataDir: dir}
			_, err := tx.LoadPending()
			require.NoError(t, err)
			switch changed {
			case "configuration":
				cfg.ApiServer.SubjectAltNames = []string{"new.example.test"}
			case "active":
				require.NoError(t, os.WriteFile(filepath.Join(dir, "certs", "external-change"), []byte("changed"), 0600))
			case "pending":
				require.NoError(t, os.WriteFile(filepath.Join(tx.Stage(), "certs", "external-change"), []byte("changed"), 0600))
			}
			before := activeCertificateFileDigests(t, dir)
			activated, err := activateCertificateRenewal(cfg, tx, func() error { return nil }, time.Now())
			require.False(t, activated)
			if changed == "pending" {
				require.ErrorContains(t, err, "changed")
			} else {
				require.NoError(t, err)
				require.NoDirExists(t, filepath.Join(dir, ".cert-renewal"))
			}
			require.Equal(t, before, activeCertificateFileDigests(t, dir))
			pending, err := loadPendingCertificateRenewal(dir)
			require.NoError(t, err)
			if changed == "pending" {
				require.NotNil(t, pending, "corrupt material must remain available for troubleshooting")
			} else {
				require.Nil(t, pending)
			}
		})
	}
}

func TestPendingRenewalCanBeReplaced(t *testing.T) {
	cfg, dir := pendingRenewalFixture(t)
	before := activeCertificateFileDigests(t, dir)
	tx := &certificates.Transaction{DataDir: dir}
	original, err := readPendingCertificateRenewal(tx)
	require.NoError(t, err)
	_, err = applyCertificateRenewal(cfg, dir, true)
	require.NoError(t, err)
	replacement, err := readPendingCertificateRenewal(tx)
	require.NoError(t, err)
	require.NotEqual(t, original.RenewedHash, replacement.RenewedHash)
	require.Equal(t, certificatesv1alpha1.RenewalModeCA, replacement.Result.Mode)
	require.Equal(t, before, activeCertificateFileDigests(t, dir))
	activated, err := activateCertificateRenewal(cfg, tx, func() error { return nil }, time.Now())
	require.NoError(t, err)
	require.True(t, activated)
	pend, err := loadPendingCertificateRenewal(dir)
	require.NoError(t, err)
	require.Nil(t, pend)
}

func TestCertStatusShowsActiveAndPendingSeparately(t *testing.T) {
	for _, format := range []string{"", certificateOutputJSON, certificateOutputYAML} {
		t.Run(format, func(t *testing.T) {
			now := time.Now().UTC().Truncate(time.Second)
			entry := certificateInventoryEntry("test", "client", certchains.CertificateRoleClient, certchains.RotationPolicyStandard, now.Add(-time.Hour), now.Add(time.Hour))
			var out, errOut bytes.Buffer
			options := &certStatusOptions{
				IOStreams: genericclioptions.IOStreams{Out: &out, ErrOut: &errOut},
				output:    format, now: func() time.Time { return now },
				loadConfig: func() (*config.Config, error) { return &config.Config{}, nil },
				loadInventory: func(*config.Config) (certchains.CertificateInventory, error) {
					return certchains.CertificateInventory{entry}, nil
				},
				loadPending: func() (*certificatesv1alpha1.CertificateRenewalResult, error) {
					return &certificatesv1alpha1.CertificateRenewalResult{Status: certificatesv1alpha1.RenewalStatusPending}, nil
				},
			}
			require.NoError(t, options.run())
			if format == "" {
				require.Contains(t, errOut.String(), "pending activation")
				return
			}
			require.Empty(t, errOut.String())
			object, _, err := certificateCodecs.UniversalDeserializer().Decode(out.Bytes(), nil, nil)
			require.NoError(t, err)
			status := object.(*certificatesv1alpha1.CertificateStatusList)
			require.True(t, entry.Certificate.NotAfter.Equal(status.Items[0].NotAfter.Time))
			require.Equal(t, certificatesv1alpha1.RenewalStatusPending, status.PendingRenewal.Status)
			statusCopy := status.DeepCopy()
			statusCopy.PendingRenewal.Status = certificatesv1alpha1.RenewalStatusValidated
			require.Equal(t, certificatesv1alpha1.RenewalStatusPending, status.PendingRenewal.Status)
		})
	}
}

func TestPendingRenewalExpiredBeforeActivation(t *testing.T) {
	cfg, dir := pendingRenewalFixture(t)
	before := activeCertificateFileDigests(t, dir)
	tx := &certificates.Transaction{DataDir: dir}
	activated, err := activateCertificateRenewal(cfg, tx, func() error { return nil }, time.Now().AddDate(20, 0, 0))
	require.NoError(t, err)
	require.False(t, activated)
	require.Equal(t, before, activeCertificateFileDigests(t, dir))
	pending, err := loadPendingCertificateRenewal(dir)
	require.NoError(t, err)
	require.Nil(t, pending)
	require.NoDirExists(t, filepath.Join(dir, ".cert-renewal"))
}

func TestPendingRenewalNotYetValidBeforeActivation(t *testing.T) {
	cfg, dir := pendingRenewalFixture(t)
	before := activeCertificateFileDigests(t, dir)
	tx := &certificates.Transaction{DataDir: dir}
	activated, err := activateCertificateRenewal(cfg, tx, func() error { return nil }, time.Now().AddDate(-20, 0, 0))
	require.ErrorContains(t, err, "pending certificates cannot be activated")
	require.False(t, activated)
	require.Equal(t, before, activeCertificateFileDigests(t, dir))
	pending, err := loadPendingCertificateRenewal(dir)
	require.NoError(t, err)
	require.NotNil(t, pending)
}

func TestPendingRenewalAllowsUnrelatedConfigurationChanges(t *testing.T) {
	cfg, dir := pendingRenewalFixture(t)
	cfg.Debugging.LogLevel = "Debug"
	cfg.ApiServer.AuditLog.MaxFiles = 20
	// This endpoint is embedded in kubeconfigs, which activation rebuilds.
	cfg.ApiServer.URL = "https://127.0.0.1:7443"
	cfg.ApiServer.Port = 7443
	activated, err := activateCertificateRenewal(cfg, &certificates.Transaction{DataDir: dir}, func() error { return nil }, time.Now())
	require.NoError(t, err)
	require.True(t, activated)
	kubeconfig, err := os.ReadFile(filepath.Join(dir, "resources", string(config.ObservabilityClient), "kubeconfig"))
	require.NoError(t, err)
	require.Contains(t, string(kubeconfig), cfg.ApiServer.URL)
}

func TestCertificateConfigHash(t *testing.T) {
	for name, change := range map[string]func(*config.Config){
		"service network":   func(cfg *config.Config) { cfg.Network.ServiceNetwork = []string{"10.44.0.0/16"} },
		"hostname":          func(cfg *config.Config) { cfg.Node.HostnameOverride = "new-node" },
		"node IP":           func(cfg *config.Config) { cfg.Node.NodeIP = "192.0.2.11" },
		"base domain":       func(cfg *config.Config) { cfg.DNS.BaseDomain = "new.example.test" },
		"advertise address": func(cfg *config.Config) { cfg.ApiServer.AdvertiseAddress = "192.0.2.12" },
		"SANs":              func(cfg *config.Config) { cfg.ApiServer.SubjectAltNames = []string{"new.example.test"} },
	} {
		t.Run(name, func(t *testing.T) {
			cfg := &config.Config{}
			before, err := certificateConfigHash(cfg)
			require.NoError(t, err)
			change(cfg)
			after, err := certificateConfigHash(cfg)
			require.NoError(t, err)
			require.NotEqual(t, before, after)
		})
	}
}

func TestCertificateActivationWithoutPendingRenewal(t *testing.T) {
	tx := &certificates.Transaction{DataDir: t.TempDir()}
	activated, err := activateCertificateRenewal(nil, tx, func() error {
		t.Fatal("no pending renewal should not require an etcd check")
		return nil
	}, time.Now())
	require.NoError(t, err)
	require.False(t, activated)
}

func TestCertRenewPreservesRootHookAndChecksPrivileges(t *testing.T) {
	var out, errOut bytes.Buffer
	rootRan := false
	root := newCertsCommand(&certStatusOptions{IOStreams: genericclioptions.IOStreams{Out: &out, ErrOut: &errOut}}, func() error { return nil })
	root.PersistentPreRun = func(_ *cobra.Command, _ []string) { rootRan = true }
	root.AddCommand(newCertsRenewCommand(&certRenewOptions{}, func() error { return errors.New("requires root") }))
	require.Equal(t, 1, RunCertsCommand(root, []string{"renew", "--ca", "-o", "json"}))
	require.True(t, rootRan)
	require.Empty(t, out.String())
	require.Contains(t, errOut.String(), "InsufficientPrivileges")
}
