package certchains

import (
	"crypto/x509"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apiserver/pkg/authentication/user"

	"github.com/openshift/microshift/pkg/util/cryptomaterial"
)

func TestSerialCounterFormat(t *testing.T) {
	for _, tc := range []struct {
		name, contents, want string
		serial               int64
	}{
		{name: "missing", want: "02\n", serial: 2},
		{name: "reserved zero", contents: "00\n", want: "02\n", serial: 2},
		{name: "hexadecimal padding", contents: "FF\n", want: "0100\n", serial: 256},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := cryptomaterial.CASerialsPath(t.TempDir())
			if tc.contents != "" {
				require.NoError(t, os.WriteFile(path, []byte(tc.contents), 0600))
			}
			g, err := ensureSerialGenerator(path)
			require.NoError(t, err)
			before, err := os.Stat(path)
			require.NoError(t, err)
			serial, err := g.Next(&x509.Certificate{})
			require.NoError(t, err)
			require.Equal(t, tc.serial, serial)
			contents, err := os.ReadFile(path)
			require.NoError(t, err)
			require.Equal(t, tc.want, string(contents))
			after, err := os.Stat(path)
			require.NoError(t, err)
			require.False(t, os.SameFile(before, after), "replace the counter instead of truncating it")
		})
	}
}

func TestSerialCounterRejectsCorruption(t *testing.T) {
	for _, contents := range []string{"", "invalid", "-01\n", "8000000000000000\n"} {
		path := cryptomaterial.CASerialsPath(t.TempDir())
		require.NoError(t, os.WriteFile(path, []byte(contents), 0600))
		_, err := ensureSerialGenerator(path)
		require.Error(t, err)
		after, err := os.ReadFile(path)
		require.NoError(t, err)
		require.Equal(t, contents, string(after))
	}
}

func TestSerialCounterRejectsStaleOrExhaustedGenerator(t *testing.T) {
	path := cryptomaterial.CASerialsPath(t.TempDir())
	g, err := ensureSerialGenerator(path)
	require.NoError(t, err)
	other, err := ensureSerialGenerator(path)
	require.NoError(t, err)
	_, err = g.Next(&x509.Certificate{})
	require.NoError(t, err)
	_, err = other.Next(&x509.Certificate{})
	require.ErrorContains(t, err, "out of sync")
	require.NoError(t, os.WriteFile(path, []byte("7FFFFFFFFFFFFFFF\n"), 0600))
	g, err = ensureSerialGenerator(path)
	require.NoError(t, err)
	_, err = g.Next(&x509.Certificate{})
	require.ErrorContains(t, err, "exhausted")
	contents, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "7FFFFFFFFFFFFFFF\n", string(contents))
}

func TestSerialCounterRetriesAfterDirectorySyncFailure(t *testing.T) {
	g, err := ensureSerialGenerator(cryptomaterial.CASerialsPath(t.TempDir()))
	require.NoError(t, err)

	syncErr := errors.New("directory sync failed")
	g.syncDir = func(*os.File) error { return syncErr }
	serial, err := g.Next(nil)
	require.ErrorIs(t, err, syncErr)
	require.Zero(t, serial)

	g.syncDir = (*os.File).Sync
	serial, err = g.Next(nil)
	require.NoError(t, err)
	require.Equal(t, int64(3), serial, "skip the serial reserved by the failed issuance")
}

func TestSerialCounterResumesStartupAfterInterruptedUpdate(t *testing.T) {
	for _, afterRename := range []bool{false, true} {
		name := "before rename"
		if afterRename {
			name = "after rename"
		}
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			subDir := filepath.Join(dir, "intermediate")
			child := NewCertificateSigner("intermediate", subDir, 24*time.Hour).
				WithClientCertificates(&ClientCertificateSigningRequestInfo{
					CSRMeta: CSRMeta{Name: "client", Validity: time.Hour}, UserInfo: &user.DefaultInfo{Name: "client", Groups: []string{"test"}},
				})
			builder := NewCertificateChains(NewCertificateSigner("root", dir, 24*time.Hour).WithSubCAs(child))
			chains, err := builder.Complete()
			require.NoError(t, err)
			path := cryptomaterial.CASerialsPath(dir)
			require.NoError(t, os.WriteFile(path, []byte("0100\n"), 0600))
			require.NoError(t, os.WriteFile(cryptomaterial.CASerialsPath(subDir), []byte("0200\n"), 0600))
			g, err := ensureSerialGenerator(path)
			require.NoError(t, err)
			interrupted := errors.New("interrupted counter update")
			if afterRename {
				g.syncDir = func(*os.File) error { return interrupted }
			} else {
				g.rename = func(string, string) error { return interrupted }
			}
			root := chains.GetSigner("root")
			root.signerConfig.SerialGenerator = g
			// Force the same intermediate reissue performed by startup recovery.
			require.NoError(t, os.WriteFile(cryptomaterial.CAKeyPath(subDir), nil, 0600))
			require.ErrorIs(t, root.SignSubCA(child), interrupted)
			want := int64(256)
			if afterRename {
				want++
			}
			resumed, err := ensureSerialGenerator(path)
			require.NoError(t, err, "the counter must retain either the previous or new value")
			require.Equal(t, want, resumed.serial)
			before, err := os.Stat(path)
			require.NoError(t, err)
			_, err = builder.Complete()
			require.NoError(t, err)
			inventory, err := builder.ValidateRenewal(true, time.Now())
			require.NoError(t, err)
			require.Equal(t, want+1, inventory[1].Certificate.SerialNumber.Int64())
			require.Greater(t, inventory[2].Certificate.SerialNumber.Int64(), int64(512))
			after, err := os.Stat(path)
			require.NoError(t, err)
			require.False(t, os.SameFile(before, after), "startup must also replace the counter atomically")
		})
	}
}
