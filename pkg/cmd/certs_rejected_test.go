package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2"

	"github.com/openshift/microshift/pkg/admin/certificates"
	certificatesv1alpha1 "github.com/openshift/microshift/pkg/apis/certificates/v1alpha1"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util/cryptomaterial/certchains"
)

func corruptPendingFixture(t *testing.T, contents string) *certificates.Transaction {
	t.Helper()
	tx := &certificates.Transaction{DataDir: t.TempDir()}
	directory := filepath.Join(tx.DataDir, ".cert-renewal")
	require.NoError(t, os.MkdirAll(filepath.Join(directory, "stage-test"), 0700))
	require.NoError(t, os.WriteFile(filepath.Join(directory, "pending.json"), []byte(contents), 0600))
	require.NoError(t, os.Mkdir(filepath.Join(tx.DataDir, "certs"), 0700))
	require.NoError(t, os.WriteFile(filepath.Join(tx.DataDir, "certs", "active"), []byte("unchanged active material"), 0600))
	return tx
}

func TestCorruptPendingRenewalAllowsStartup(t *testing.T) {
	for name, record := range map[string]string{
		"truncated JSON":    `{"credential-bearing-corrupt-json`,
		"empty record":      "",
		"null record":       `null`,
		"invalid directory": `{"directory":"../outside","metadata":{}}`,
		"invalid metadata":  `{"directory":"stage-test","metadata":"credential-bearing-data"}`,
		"missing fields":    `{"directory":"stage-test","metadata":{"version":1}}`,
		"missing stage":     `{"directory":"stage-missing","metadata":{"version":1}}`,
	} {
		t.Run(name, func(t *testing.T) {
			tx := corruptPendingFixture(t, record)
			before, err := certificates.DigestTree(filepath.Join(tx.DataDir, "certs"))
			require.NoError(t, err)
			var logs bytes.Buffer
			logState := klog.CaptureState()
			t.Cleanup(logState.Restore)
			klog.LogToStderr(false)
			klog.SetOutput(&logs)
			activated, err := activateCertificateRenewal(nil, tx, func() error {
				t.Fatal("corrupt metadata must not attempt activation")
				return nil
			}, time.Now())
			require.NoError(t, err)
			require.False(t, activated, "startup must use normal certificate initialization")
			klog.Flush()
			require.Contains(t, logs.String(), rejectedPendingRenewalWarning)
			require.Contains(t, logs.String(), "continuing normal certificate initialization")
			require.NotContains(t, logs.String(), "credential-bearing")
			after, err := certificates.DigestTree(filepath.Join(tx.DataDir, "certs"))
			require.NoError(t, err)
			require.Equal(t, before, after)
			pending, warnings, err := loadCertificateRenewalStatus(tx.DataDir)
			require.NoError(t, err)
			require.Nil(t, pending)
			require.Equal(t, []string{rejectedPendingRenewalWarning}, warnings)
			activated, err = activateCertificateRenewal(nil, tx, nil, time.Now())
			require.NoError(t, err, "a subsequent startup must not retry rejected metadata")
			require.False(t, activated)
		})
	}
}

func TestRejectedRenewalWarningInAllStatusFormats(t *testing.T) {
	tx := corruptPendingFixture(t, `{"credential-bearing-corrupt-json`)
	now := time.Now().UTC().Truncate(time.Second)
	entry := certificateInventoryEntry("test", "client", certchains.CertificateRoleClient, certchains.RotationPolicyStandard, now.Add(-time.Hour), now.Add(24*time.Hour))
	for _, quarantined := range []bool{false, true} {
		wantWarning := invalidPendingRenewalWarning
		if quarantined {
			require.NoError(t, tx.QuarantinePending())
			wantWarning = rejectedPendingRenewalWarning
		}
		for _, format := range []string{"", certificateOutputJSON, certificateOutputYAML} {
			var out, errOut bytes.Buffer
			options := &certStatusOptions{
				IOStreams: genericclioptions.IOStreams{Out: &out, ErrOut: &errOut},
				output:    format, now: func() time.Time { return now },
				loadConfig: func() (*config.Config, error) { return &config.Config{}, nil },
				loadInventory: func(*config.Config) (certchains.CertificateInventory, error) {
					return certchains.CertificateInventory{entry}, nil
				},
				loadPending: func() (*certificatesv1alpha1.CertificateRenewalResult, []string, error) {
					return loadCertificateRenewalStatus(tx.DataDir)
				},
			}
			require.NoError(t, options.run())
			require.NotContains(t, out.String()+errOut.String(), "credential-bearing")
			if format == "" {
				require.Contains(t, out.String(), "Healthy")
				require.Contains(t, errOut.String(), wantWarning)
				continue
			}
			require.Empty(t, errOut.String())
			object, _, err := certificateCodecs.UniversalDeserializer().Decode(out.Bytes(), nil, nil)
			require.NoError(t, err)
			status := object.(*certificatesv1alpha1.CertificateStatusList)
			require.Len(t, status.Items, 1)
			require.Equal(t, certificatesv1alpha1.CertificateStatusHealthy, status.Items[0].Status)
			require.Nil(t, status.PendingRenewal)
			require.Equal(t, []string{wantWarning}, status.Warnings)
		}
	}
}

func TestPendingRenewalFallbackDoesNotHideOtherFailures(t *testing.T) {
	for _, failure := range []string{"future version", "record symlink", "incomplete activation"} {
		t.Run(failure, func(t *testing.T) {
			tx := corruptPendingFixture(t, `{"directory":"stage-test","metadata":{"version":2}}`)
			record := filepath.Join(tx.DataDir, ".cert-renewal", "pending.json")
			switch failure {
			case "record symlink":
				require.NoError(t, os.Rename(record, record+".saved"))
				require.NoError(t, os.Symlink(record+".saved", record))
			case "incomplete activation":
				require.NoError(t, os.WriteFile(record, []byte("invalid JSON"), 0600))
				require.NoError(t, os.WriteFile(filepath.Join(tx.DataDir, ".cert-renewal", "committing"), nil, 0600))
			}
			activated, err := activateCertificateRenewal(nil, tx, nil, time.Now())
			require.Error(t, err)
			require.False(t, activated)
			require.FileExists(t, record)
			require.NoDirExists(t, filepath.Join(tx.DataDir, ".cert-renewal-rejected"))
			if failure != "incomplete activation" {
				_, _, err = loadCertificateRenewalStatus(tx.DataDir)
				require.Error(t, err, "status must not hide filesystem or unsupported-version errors")
			}
		})
	}
}

func TestRejectedRenewalWarningClearsOnlyAfterActivation(t *testing.T) {
	cfg, dir := pendingRenewalFixture(t)
	tx := &certificates.Transaction{DataDir: dir}
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".cert-renewal", "pending.json"), []byte("invalid JSON"), 0600))
	activated, err := activateCertificateRenewal(cfg, tx, nil, time.Now())
	require.NoError(t, err)
	require.False(t, activated)
	_, err = applyCertificateRenewal(cfg, dir, false)
	require.NoError(t, err)
	pending, warnings, err := loadCertificateRenewalStatus(dir)
	require.NoError(t, err)
	require.NotNil(t, pending)
	require.Equal(t, []string{rejectedPendingRenewalWarning}, warnings)
	activated, err = activateCertificateRenewal(cfg, tx, func() error { return nil }, time.Now())
	require.NoError(t, err)
	require.True(t, activated)
	pending, warnings, err = loadCertificateRenewalStatus(dir)
	require.NoError(t, err)
	require.Nil(t, pending)
	require.Empty(t, warnings)
	require.NoDirExists(t, filepath.Join(dir, ".cert-renewal-rejected"))
}
