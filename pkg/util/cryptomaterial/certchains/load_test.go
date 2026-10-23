package certchains

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/openshift/microshift/pkg/util/cryptomaterial"
	"github.com/stretchr/testify/require"
	"k8s.io/apiserver/pkg/authentication/user"
)

func TestLoadChainsDoesNotReissueOrRepair(t *testing.T) {
	dir := t.TempDir()
	rootDir := filepath.Join(dir, "root")
	childDir := filepath.Join(rootDir, "child")
	builder := NewCertificateChains(NewCertificateSigner("root", rootDir, 24*time.Hour).
		WithSubCAs(NewCertificateSigner("child", childDir, 12*time.Hour).
			WithClientCertificates(&ClientCertificateSigningRequestInfo{
				CSRMeta:  CSRMeta{Name: "client", Service: "test", Validity: time.Hour},
				UserInfo: &user.DefaultInfo{Name: "client-without-groups"},
			}))).WithCABundle(filepath.Join(dir, "bundle.crt"), []string{"root"}, []string{"root", "child"})
	created, err := builder.Complete()
	require.NoError(t, err)
	certBefore, keyBefore, err := created.GetCertKey("root", "child", "client")
	require.NoError(t, err)
	serialPath := cryptomaterial.CASerialsPath(rootDir)
	require.NoError(t, os.WriteFile(serialPath, []byte("0"), 0600))
	loaded, err := builder.Load()
	require.NoError(t, err)
	serial, err := os.ReadFile(serialPath)
	require.NoError(t, err)
	require.Equal(t, "0", string(serial), "read-only loading must not repair the issuance counter")
	certAfter, keyAfter, err := loaded.GetCertKey("root", "child", "client")
	require.NoError(t, err)
	require.True(t, bytes.Equal(certBefore, certAfter))
	require.True(t, bytes.Equal(keyBefore, keyAfter), "loading must preserve the key without disclosing its contents")
	for i, entry := range loaded.Inventory() {
		require.True(t, bytes.Equal(created.Inventory()[i].Certificate.Raw, entry.Certificate.Raw))
	}
	keyPath := filepath.Join(childDir, "client", "client.key")
	require.NoError(t, os.Remove(keyPath))
	_, err = builder.Load()
	require.Error(t, err)
	require.NoFileExists(t, keyPath, "read-only loading must not repair missing material")
}
