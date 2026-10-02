package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/openshift/microshift/pkg/admin/certificates"
	certificatesv1alpha1 "github.com/openshift/microshift/pkg/apis/certificates/v1alpha1"
	"github.com/openshift/microshift/pkg/config"
)

func TestCertificateLockSurvivesDataReplacement(t *testing.T) {
	for _, replaced := range []string{filepath.Join(config.DataDir, "certs"), config.DataDir} {
		t.Run(filepath.Base(replaced), func(t *testing.T) {
			// Mirror the production layout below a temporary root.
			root := t.TempDir()
			replacedPath := filepath.Join(root, strings.TrimPrefix(replaced, "/"))
			lockPath := filepath.Join(root, strings.TrimPrefix(certificateLockPath, "/"))
			require.NoError(t, os.MkdirAll(replacedPath, 0700))
			lock, err := certificates.Lock(lockPath, true)
			require.NoError(t, err)
			defer func() { _ = lock.Close() }()
			original, err := lock.Stat()
			require.NoError(t, err)

			require.NoError(t, os.Rename(replacedPath, replacedPath+".saved"))
			require.NoError(t, os.MkdirAll(replacedPath, 0700))
			info, err := os.Stat(lockPath)
			require.NoError(t, err)
			require.True(t, os.SameFile(original, info))
			for _, exclusive := range []bool{false, true} {
				_, err := certificates.Lock(lockPath, exclusive)
				require.ErrorIs(t, err, certificates.ErrBusy)
			}
		})
	}
}

func TestPrepareCertificateTransaction(t *testing.T) {
	for _, writing := range []bool{false, true} {
		t.Run(map[bool]string{false: "read", true: "write"}[writing], func(t *testing.T) {
			tx := &certificates.Transaction{DataDir: t.TempDir()}
			lockPath := filepath.Join(t.TempDir(), "lock")
			stoppedChecks := 0
			release, err := prepareCertificateTransaction(tx, lockPath, writing, func() error { stoppedChecks++; return nil })
			require.NoError(t, err)
			require.Zero(t, stoppedChecks, "preparing renewal, status and dry-run do not require a stopped service")
			_, err = certificates.Lock(lockPath, true)
			require.ErrorIs(t, err, certificates.ErrBusy)
			release()
			lock, err := certificates.Lock(lockPath, true)
			require.NoError(t, err)
			require.NoError(t, lock.Close())
		})
	}
}

func TestPrepareCertificateRecovery(t *testing.T) {
	for _, stopped := range []bool{false, true} {
		t.Run(map[bool]string{false: "running", true: "stopped"}[stopped], func(t *testing.T) {
			tx := &certificates.Transaction{DataDir: t.TempDir()}
			transactionDir := filepath.Join(tx.DataDir, ".cert-renewal")
			for _, name := range []string{"certs", "resources"} {
				backup := filepath.Join(transactionDir, "previous", name)
				require.NoError(t, os.MkdirAll(backup, 0700))
				require.NoError(t, os.WriteFile(filepath.Join(backup, "original"), []byte("old"), 0600))
				require.NoError(t, os.Mkdir(filepath.Join(tx.DataDir, name), 0700))
			}
			require.NoError(t, os.WriteFile(filepath.Join(transactionDir, "committing"), nil, 0600))
			require.ErrorContains(t, recoverCertificateStartup(tx, func() error { return errors.New("etcd still running") }), "etcd still running")
			lockPath := filepath.Join(t.TempDir(), "lock")
			release, err := prepareCertificateTransaction(tx, lockPath, false, func() error {
				if !stopped {
					return errors.New("service is running")
				}
				return nil
			})
			if !stopped {
				var commandError *certificateCommandError
				require.ErrorAs(t, err, &commandError)
				require.Equal(t, certificatesv1alpha1.ErrorCodeRecoveryFailed, commandError.code)
				require.Nil(t, release)
				require.FileExists(t, filepath.Join(transactionDir, "committing"))
			} else {
				require.NoError(t, err)
				release()
				for _, name := range []string{"certs", "resources"} {
					contents, err := os.ReadFile(filepath.Join(tx.DataDir, name, "original"))
					require.NoError(t, err)
					require.Equal(t, "old", string(contents))
				}
				require.NoDirExists(t, transactionDir)
				require.NoError(t, recoverCertificateStartup(tx, func() error { return nil }))
			}
			lock, err := certificates.Lock(lockPath, true)
			require.NoError(t, err, "failed recovery must release its lock")
			require.NoError(t, lock.Close())
		})
	}
}

func TestPrepareCertificateAccessRefusesConcurrentWriter(t *testing.T) {
	tx := &certificates.Transaction{DataDir: t.TempDir()}
	lockPath := filepath.Join(t.TempDir(), "lock")
	startup, err := certificates.Lock(lockPath, true)
	require.NoError(t, err)
	defer func() { _ = startup.Close() }()
	for _, writing := range []bool{false, true} {
		release, err := prepareCertificateTransaction(tx, lockPath, writing, func() error { return nil })
		require.ErrorIs(t, err, certificates.ErrBusy)
		require.Nil(t, release)
	}
}

func TestAdministratorKubeconfigCannotEscapeStaging(t *testing.T) {
	root := filepath.Join(t.TempDir(), "stage", "resources", "kubeadmin")
	for _, name := range []string{"", ".", "..", "../outside", "/outside", "a/../../outside"} {
		path, err := adminKubeconfigPath(root, name)
		require.Error(t, err)
		require.Empty(t, path)
	}
	for _, name := range []string{"node.example.test", "192.0.2.1", "2001:db8::1", "*.example.test"} {
		path, err := adminKubeconfigPath(root, name)
		require.NoError(t, err)
		require.Equal(t, filepath.Join(root, name, "kubeconfig"), path)
	}
}
