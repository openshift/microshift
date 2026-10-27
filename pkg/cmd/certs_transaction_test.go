package cmd

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/openshift/microshift/pkg/admin/certificates"
	certificatesv1alpha1 "github.com/openshift/microshift/pkg/apis/certificates/v1alpha1"
	"github.com/openshift/microshift/pkg/config"
)

func TestCertificateStartupLockWaitsForOperation(t *testing.T) {
	for _, exclusive := range []bool{false, true} {
		t.Run(map[bool]string{false: "reader", true: "writer"}[exclusive], func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "lock")
			operation, err := certificates.Lock(path, exclusive)
			require.NoError(t, err)
			defer func() { _ = operation.Close() }()
			released := make(chan error, 1)
			timer := time.AfterFunc(50*time.Millisecond, func() { released <- operation.Close() })
			defer timer.Stop()
			startup, err := waitForCertificateStartupLock(path, 5*time.Second)
			require.NoError(t, err)
			defer func() { _ = startup.Close() }()
			require.NoError(t, <-released)
			_, err = certificates.Lock(path, false)
			require.ErrorIs(t, err, certificates.ErrBusy, "startup must hold an exclusive lock")
			require.NoError(t, startup.Close())
			next, err := certificates.Lock(path, true)
			require.NoError(t, err, "startup cleanup must release the lock")
			require.NoError(t, next.Close())
		})
	}
}

func TestCertificateStartupLockTimeout(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lock")
	operation, err := certificates.Lock(path, true)
	require.NoError(t, err)
	defer func() { _ = operation.Close() }()
	startup, err := waitForCertificateStartupLock(path, 20*time.Millisecond)
	require.Nil(t, startup)
	require.ErrorIs(t, err, certificates.ErrBusy)
	require.ErrorContains(t, err, "timed out after 20ms")
	_, err = certificates.Lock(path, true)
	require.ErrorIs(t, err, certificates.ErrBusy, "timeout must not release the other operation's lock")
	require.NoError(t, operation.Close())
	startup, err = waitForCertificateStartupLock(path, time.Second)
	require.NoError(t, err)
	require.NoError(t, startup.Close())
}

func TestCertificateStartupLockRejectsInvalidPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "not-a-directory")
	require.NoError(t, os.WriteFile(path, nil, 0600))
	startup, err := waitForCertificateStartupLock(filepath.Join(path, "lock"), time.Second)
	require.Nil(t, startup)
	require.ErrorContains(t, err, "cannot create certificate lock directory")
	require.NotErrorIs(t, err, certificates.ErrBusy)
}

func TestStopCertificateStartupEtcd(t *testing.T) {
	for _, tc := range []struct {
		name       string
		state      string
		stopScript string
		wantStop   bool
		wantErr    string
	}{
		{name: "inactive", state: "inactive"},
		{name: "failed", state: "failed"},
		{name: "active", state: "active", stopScript: `printf 'inactive\n' > "$CERTIFICATE_TEST_STATE_FILE"`, wantStop: true},
		{name: "deactivating", state: "deactivating", stopScript: `sleep 0.05; printf 'inactive\n' > "$CERTIFICATE_TEST_STATE_FILE"`, wantStop: true},
		{name: "scope disappeared", state: "active", stopScript: `printf 'inactive\n' > "$CERTIFICATE_TEST_STATE_FILE"; exit 5`, wantStop: true},
		{name: "query failure", state: "query-error", wantErr: "cannot determine whether microshift-etcd.scope is stopped"},
		{name: "stop failure", state: "active", stopScript: "exit 1", wantStop: true, wantErr: "state: active"},
		{name: "still active", state: "active", stopScript: "exit 0", wantStop: true, wantErr: "state: active"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stopFile := fakeCertificateSystemctl(t, tc.state, tc.stopScript)
			err := stopCertificateStartupEtcd(time.Second)
			if tc.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.wantErr)
			}
			if tc.wantStop {
				require.FileExists(t, stopFile)
			} else {
				require.NoFileExists(t, stopFile)
			}
		})
	}
}

func TestStopCertificateStartupEtcdTimeout(t *testing.T) {
	fakeCertificateSystemctl(t, "active", "exec sleep 10")
	err := stopCertificateStartupEtcd(100 * time.Millisecond)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.ErrorContains(t, err, "timed out after 100ms waiting for microshift-etcd.scope to stop")
	require.ErrorContains(t, certificateUnitsStopped("microshift-etcd.scope"), "state: active")
}

func TestCertificateCLIRecoveryDoesNotStopEtcd(t *testing.T) {
	stopFile := fakeCertificateSystemctl(t, "active", "exit 0")
	require.ErrorContains(t, certificateServicesStopped(), "microshift-etcd.scope to be stopped (state: active)")
	require.NoFileExists(t, stopFile)
}

func fakeCertificateSystemctl(t *testing.T, state, stopScript string) string {
	t.Helper()
	dir := t.TempDir()
	stateFile := filepath.Join(dir, "state")
	stopFile := filepath.Join(dir, "stopped")
	require.NoError(t, os.WriteFile(stateFile, []byte(state), 0600))
	script := `#!/bin/sh
case "$1" in
show)
    if [ "$5" = microshift.service ]; then
        printf 'inactive\n'
        exit 0
    fi
    state=$(cat "$CERTIFICATE_TEST_STATE_FILE") || exit 1
    [ "$state" != query-error ] || exit 1
    printf '%s\n' "$state"
    ;;
stop)
    [ "$#" -eq 2 ] && [ "$2" = microshift-etcd.scope ] || exit 2
    touch "$CERTIFICATE_TEST_STOP_FILE"
    ` + stopScript + `
    ;;
*) exit 1 ;;
esac
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "systemctl"), []byte(script), 0700))
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("CERTIFICATE_TEST_STATE_FILE", stateFile)
	t.Setenv("CERTIFICATE_TEST_STOP_FILE", stopFile)
	return stopFile
}

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
