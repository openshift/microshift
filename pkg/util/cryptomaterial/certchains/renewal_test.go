package certchains

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apiserver/pkg/authentication/user"

	"github.com/openshift/library-go/pkg/crypto"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
)

func TestPlanRenewal(t *testing.T) {
	dir := t.TempDir()
	builder := NewCertificateChains(
		NewCertificateSigner("root", dir, 24*time.Hour).WithService("test").WithClientCertificates(
			&ClientCertificateSigningRequestInfo{
				CSRMeta:  CSRMeta{Name: "client", Service: "test", Validity: 48 * time.Hour, RotationPolicy: RotationPolicyStandard},
				UserInfo: &user.DefaultInfo{Name: "test-client"},
			},
		),
	)
	chains, err := builder.Complete()
	require.NoError(t, err)
	before := chains.Inventory()
	filesBefore := renewalFiles(t, dir)
	now := time.Now()
	for _, renewCAs := range []bool{false, true} {
		plan, err := builder.PlanRenewal(renewCAs, now)
		require.NoError(t, err)
		require.Equal(t, "client", plan[0].Name)
		require.Equal(t, "root", plan[0].ParentCA)
		if renewCAs {
			require.Len(t, plan, 2)
			require.Equal(t, "root", plan[1].Name)
			require.Equal(t, plan[1].NewNotAfter, plan[0].NewNotAfter)
		} else {
			require.Len(t, plan, 1)
			require.Equal(t, before[0].Certificate.NotAfter, plan[0].NewNotAfter)
		}
	}
	require.Equal(t, filesBefore, renewalFiles(t, dir), "planning must not write any files")

	expiredTime := before[0].Certificate.NotAfter.Add(time.Hour)
	_, err = builder.PlanRenewal(false, expiredTime)
	require.ErrorContains(t, err, "use --ca")
	require.ErrorIs(t, err, ErrCertificateExpired)
	_, err = builder.PlanRenewal(true, expiredTime)
	require.NoError(t, err, "CA renewal must support recovery from expired certificates")

	keyPath := cryptomaterial.ClientKeyPath(filepath.Join(dir, "client"))
	require.NoError(t, os.WriteFile(keyPath, []byte("invalid test key"), 0600))
	_, err = builder.PlanRenewal(false, now)
	require.EqualError(t, err, `invalid certificate/key pair for "client"`)
	require.NoError(t, os.Remove(keyPath))
	_, err = builder.PlanRenewal(true, now)
	require.EqualError(t, err, `cannot read private key for certificate "client"`)
}

func TestValidateRenewalDistinguishesExpiry(t *testing.T) {
	builder := NewCertificateChains(NewCertificateSigner("root", t.TempDir(), 24*time.Hour).
		WithService("test").WithClientCertificates(&ClientCertificateSigningRequestInfo{
		CSRMeta:  CSRMeta{Name: "client", Service: "test", Validity: time.Hour},
		UserInfo: &user.DefaultInfo{Name: "test-client"},
	}))
	chains, err := builder.Complete()
	require.NoError(t, err)
	inventory := chains.Inventory()
	for _, entry := range inventory {
		t.Run(string(entry.Role), func(t *testing.T) {
			_, err := builder.ValidateRenewal(false, entry.Certificate.NotAfter)
			require.ErrorIs(t, err, ErrCertificateExpired)
			_, err = builder.ValidateRenewal(false, entry.Certificate.NotBefore.Add(-time.Second))
			require.Error(t, err)
			require.NotErrorIs(t, err, ErrCertificateExpired)
		})
	}
}

func TestRenewalBoundsDescendantsAndReloadsPeers(t *testing.T) {
	dir := t.TempDir()
	bundle := filepath.Join(dir, "bundle.crt")
	child := NewCertificateSigner("intermediate", filepath.Join(dir, "intermediate"), 48*time.Hour).
		WithService("test").WithPeerCertificiates(&PeerCertificateSigningRequestInfo{
		CSRMeta:  CSRMeta{Name: "peer", Service: "test", Validity: 72 * time.Hour, RotationPolicy: RotationPolicyStandard},
		UserInfo: &user.DefaultInfo{Name: "peer"}, Hostnames: []string{"localhost"},
	})
	builder := NewCertificateChains(NewCertificateSigner("root", dir, 24*time.Hour).WithService("test").WithSubCAs(child)).
		WithCABundle(bundle, []string{"root"}, []string{"root", "intermediate"})
	chains, err := builder.Complete()
	require.NoError(t, err)
	require.NoError(t, chains.RegenerateForRenewal("root"))
	before, err := builder.ValidateRenewal(true, time.Now())
	require.NoError(t, err)
	require.Len(t, before, 3)
	for _, entry := range before {
		require.Equal(t, before[0].Certificate.NotAfter, entry.Certificate.NotAfter)
	}
	// Reload existing files: peers must stay in the in-memory inventory so a
	// cascade regenerates them along with their signer, rather than losing them.
	chains, err = builder.Complete()
	require.NoError(t, err)
	require.Len(t, chains.Inventory(), 3)
	require.NoError(t, chains.RegenerateForRenewal("root"))
	after, err := builder.ValidateRenewal(true, time.Now())
	require.NoError(t, err)
	for i := range before {
		require.NotEqual(t, before[i].Certificate.Raw, after[i].Certificate.Raw)
		require.Equal(t, after[0].Certificate.NotAfter, after[i].Certificate.NotAfter)
	}
	require.NoError(t, os.WriteFile(bundle, []byte("invalid bundle"), 0600))
	_, err = builder.ValidateRenewal(true, time.Now())
	require.ErrorContains(t, err, "invalid CA bundle")
}

func TestLeafRenewalPreservesLegacyCAs(t *testing.T) {
	dir := t.TempDir()
	child := NewCertificateSigner("intermediate", filepath.Join(dir, "intermediate"), 48*time.Hour).
		WithService("test").WithClientCertificates(&ClientCertificateSigningRequestInfo{
		CSRMeta:  CSRMeta{Name: "client", Service: "test", Validity: 72 * time.Hour, RotationPolicy: RotationPolicyStandard},
		UserInfo: &user.DefaultInfo{Name: "client"},
	})
	builder := NewCertificateChains(NewCertificateSigner("root", dir, 24*time.Hour).WithService("test").WithSubCAs(child))
	chains, err := builder.Complete()
	require.NoError(t, err)
	before := chains.Inventory()
	require.True(t, before[1].Certificate.NotAfter.After(before[0].Certificate.NotAfter), "simulate an existing unbounded intermediate")
	plan, err := builder.PlanRenewal(false, time.Now())
	require.NoError(t, err)
	require.Len(t, plan, 1)
	require.Equal(t, before[0].Certificate.NotAfter, plan[0].NewNotAfter)
	chains, err = builder.Complete()
	require.NoError(t, err)
	require.NoError(t, chains.RegenerateForRenewal("root", "intermediate", "client"))
	after, err := builder.ValidateRenewal(false, time.Now())
	require.NoError(t, err)
	require.Equal(t, before[0].Certificate.Raw, after[0].Certificate.Raw)
	require.Equal(t, before[1].Certificate.Raw, after[1].Certificate.Raw)
	require.Equal(t, before[0].Certificate.NotAfter, after[2].Certificate.NotAfter)
	_, err = builder.ValidateRenewal(true, time.Now())
	require.ErrorContains(t, err, "expires after its signing CA", "full CA renewal must validate all newly issued CAs too")
}

func TestRenewExpiredCAWithChangedServingNames(t *testing.T) {
	dir := t.TempDir()
	expiredCA, err := crypto.UnsafeMakeSelfSignedCAConfigForDurationAtTime("root",
		func() time.Time { return time.Now().Add(-48 * time.Hour) }, 24*time.Hour)
	require.NoError(t, err)
	require.NoError(t, expiredCA.WriteCertConfigFile(cryptomaterial.CACertPath(dir), cryptomaterial.CAKeyPath(dir)))
	require.NoError(t, os.WriteFile(cryptomaterial.CASerialsPath(dir), []byte("00\n"), 0600))
	serving := &ServingCertificateSigningRequestInfo{
		CSRMeta: CSRMeta{Name: "server", Service: "test", Validity: 48 * time.Hour}, Hostnames: []string{"old.example.test"},
	}
	builder := NewCertificateChains(NewCertificateSigner("root", dir, 24*time.Hour).WithService("test").WithServingCertificates(serving))
	_, err = builder.Complete()
	require.NoError(t, err)
	_, err = builder.PlanRenewal(false, time.Now())
	require.ErrorContains(t, err, "use --ca")
	_, err = builder.PlanRenewal(true, time.Now())
	require.NoError(t, err)
	// Loading a staged copy may issue a temporary leaf for changed names. It
	// must not prevent replacing the expired CA and issuing the final leaf.
	serving.Hostnames = []string{"new.example.test"}
	chains, err := builder.Complete()
	require.NoError(t, err)
	require.NoError(t, chains.RegenerateForRenewal("root"))
	after, err := builder.ValidateRenewal(true, time.Now())
	require.NoError(t, err)
	require.Equal(t, serving.Hostnames, after[1].Certificate.DNSNames)
	require.Equal(t, after[0].Certificate.NotAfter, after[1].Certificate.NotAfter)
}

func renewalFiles(t *testing.T, dir string) map[string]string {
	t.Helper()
	files := map[string]string{}
	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			files[path] = string(data)
		}
		return nil
	})
	require.NoError(t, err)
	return files
}
