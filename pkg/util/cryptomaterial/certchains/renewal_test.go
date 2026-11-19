package certchains

import (
	"maps"
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
		if before[i].ParentCA != "" {
			require.Positive(t, after[i].Certificate.SerialNumber.Cmp(before[i].Certificate.SerialNumber))
		}
	}
	require.NoError(t, os.WriteFile(bundle, []byte("invalid bundle"), 0600))
	_, err = builder.ValidateRenewal(true, time.Now())
	require.ErrorContains(t, err, "invalid CA bundle")
}

func TestRenewalPreservesSerialCounters(t *testing.T) {
	dir := t.TempDir()
	subDir := filepath.Join(dir, "intermediate")
	child := NewCertificateSigner("intermediate", subDir, 24*time.Hour).
		WithClientCertificates(&ClientCertificateSigningRequestInfo{
			CSRMeta:  CSRMeta{Name: "client", Validity: time.Hour},
			UserInfo: &user.DefaultInfo{Name: "client"},
		})
	builder := NewCertificateChains(NewCertificateSigner("root", dir, 24*time.Hour).WithSubCAs(child))
	_, err := builder.Complete()
	require.NoError(t, err)
	// Counters can exceed the serials of the certificates still on disk.
	require.NoError(t, os.WriteFile(cryptomaterial.CASerialsPath(dir), []byte("0A\n"), 0600))
	require.NoError(t, os.WriteFile(cryptomaterial.CASerialsPath(subDir), []byte("20\n"), 0600))

	for range 2 {
		chains, err := builder.Load()
		require.NoError(t, err)
		rootCounter, err := crypto.NewSerialFileGenerator(cryptomaterial.CASerialsPath(dir))
		require.NoError(t, err)
		subCounter, err := crypto.NewSerialFileGenerator(cryptomaterial.CASerialsPath(subDir))
		require.NoError(t, err)

		require.NoError(t, chains.RegenerateForRenewal("root"))
		inventory, err := builder.ValidateRenewal(true, time.Now())
		require.NoError(t, err)
		require.Len(t, inventory, 3)
		require.Equal(t, rootCounter.Serial+1, inventory[1].Certificate.SerialNumber.Int64())
		require.Greater(t, inventory[2].Certificate.SerialNumber.Int64(), subCounter.Serial)

		rootAfter, err := crypto.NewSerialFileGenerator(cryptomaterial.CASerialsPath(dir))
		require.NoError(t, err)
		subAfter, err := crypto.NewSerialFileGenerator(cryptomaterial.CASerialsPath(subDir))
		require.NoError(t, err)
		require.Equal(t, rootCounter.Serial+1, rootAfter.Serial)
		require.Equal(t, inventory[2].Certificate.SerialNumber.Int64(), subAfter.Serial)
	}
}

func TestCompleteRepairsInterruptedCARenewal(t *testing.T) {
	for _, tc := range []struct {
		name          string
		subCA         bool
		incompleteKey bool
	}{
		{name: "root replaced before descendants"},
		{name: "root key write interrupted", incompleteKey: true},
		{name: "intermediate replaced before descendants", subCA: true},
		{name: "intermediate key write interrupted", subCA: true, incompleteKey: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			subDir := filepath.Join(dir, "intermediate")
			newBuilder := func() CertificateChainsBuilder {
				child := NewCertificateSigner("intermediate", subDir, 24*time.Hour).
					WithClientCertificates(&ClientCertificateSigningRequestInfo{
						CSRMeta: CSRMeta{Name: "child-client", Validity: time.Hour}, UserInfo: &user.DefaultInfo{Name: "child", Groups: []string{"test"}},
					})
				root := NewCertificateSigner("root", dir, 24*time.Hour).WithSubCAs(child).
					WithClientCertificates(&ClientCertificateSigningRequestInfo{
						CSRMeta: CSRMeta{Name: "client", Validity: time.Hour}, UserInfo: &user.DefaultInfo{Name: "client", Groups: []string{"test"}},
					}).WithServingCertificates(&ServingCertificateSigningRequestInfo{
					CSRMeta: CSRMeta{Name: "server", Validity: time.Hour}, Hostnames: []string{"localhost"},
				}).WithPeerCertificiates(&PeerCertificateSigningRequestInfo{
					CSRMeta: CSRMeta{Name: "peer", Validity: time.Hour}, UserInfo: &user.DefaultInfo{Name: "peer"}, Hostnames: []string{"localhost"},
				})
				return NewCertificateChains(root).WithCABundle(filepath.Join(dir, "bundle.crt"), []string{"root"}, []string{"root", "intermediate"})
			}
			chains, err := newBuilder().Complete()
			require.NoError(t, err)
			before := chains.Inventory()
			// Preserve issuance history beyond the certificates currently on disk.
			require.NoError(t, os.WriteFile(cryptomaterial.CASerialsPath(dir), []byte("0A\n"), 0600))
			require.NoError(t, os.WriteFile(cryptomaterial.CASerialsPath(subDir), []byte("20\n"), 0600))
			chains, err = newBuilder().Complete()
			require.NoError(t, err)

			// Stop after replacing a CA, before regenerating any descendants.
			keyPath := cryptomaterial.CAKeyPath(dir)
			if tc.subCA {
				keyPath = cryptomaterial.CAKeyPath(subDir)
				_, err = chains.GetSigner("root").signerConfig.MakeAndWriteSubCA(
					cryptomaterial.CABundlePath(subDir), keyPath, "", "intermediate", 24*time.Hour)
				require.NoError(t, err)
			} else {
				require.NoError(t, chains.GetSigner("root").regenerateSelf())
			}
			if tc.incompleteKey {
				// The library truncates the key before writing the certificate.
				require.NoError(t, os.WriteFile(keyPath, nil, 0600))
			}

			builder := newBuilder()
			_, err = builder.Complete()
			require.NoError(t, err)
			after, err := builder.ValidateRenewal(true, time.Now())
			require.NoError(t, err, "startup must repair the chain, key pairs and CA bundles")
			for i, entry := range after {
				if tc.subCA && entry.Name != "intermediate" && entry.ParentCA != "intermediate" {
					require.Equal(t, before[i].Certificate.Raw, entry.Certificate.Raw, entry.Name)
					continue
				}
				require.NotEqual(t, before[i].Certificate.Raw, entry.Certificate.Raw, entry.Name)
				switch entry.ParentCA {
				case "root":
					require.Greater(t, entry.Certificate.SerialNumber.Int64(), int64(10), entry.Name)
				case "intermediate":
					require.Greater(t, entry.Certificate.SerialNumber.Int64(), int64(32), entry.Name)
				}
			}
			// A subsequent startup must reuse the repaired material and counters.
			files := renewalFiles(t, dir)
			_, err = newBuilder().Complete()
			require.NoError(t, err)
			require.True(t, maps.Equal(files, renewalFiles(t, dir)), "a subsequent startup must not change the repaired files")
		})
	}
}

func TestCompleteRejectsInvalidSerialCounter(t *testing.T) {
	for _, signer := range []string{"root", "intermediate"} {
		t.Run(signer, func(t *testing.T) {
			dir := t.TempDir()
			rootDir := filepath.Join(dir, "root")
			subDir := filepath.Join(dir, "intermediate")
			builder := NewCertificateChains(NewCertificateSigner("root", rootDir, 24*time.Hour).
				WithSubCAs(NewCertificateSigner("intermediate", subDir, 24*time.Hour)))
			_, err := builder.Complete()
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(cryptomaterial.CASerialsPath(filepath.Join(dir, signer)), []byte("invalid"), 0600))
			before := renewalFiles(t, dir)
			_, err = builder.Complete()
			require.Error(t, err)
			require.True(t, maps.Equal(before, renewalFiles(t, dir)), "do not reset a corrupt counter or replace its CA")
		})
	}
}

func TestLeafRenewalPreservesLegacyCAs(t *testing.T) {
	dir := t.TempDir()
	child := NewCertificateSigner("intermediate", filepath.Join(dir, "intermediate"), 48*time.Hour).
		WithService("test").WithClientCertificates(&ClientCertificateSigningRequestInfo{
		CSRMeta:  CSRMeta{Name: "client", Service: "test", Validity: 72 * time.Hour, RotationPolicy: RotationPolicyStandard},
		UserInfo: &user.DefaultInfo{Name: "client"},
	})
	root := NewCertificateSigner("root", dir, 24*time.Hour).WithService("test").WithSubCAs(child)
	// Model PKI issued before startup enforced issuer validity bounds.
	root.(*certificateSigner).limitValidity = false
	builder := NewCertificateChains(root)
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
	// Startup may issue a temporary leaf for changed names. It must not prevent
	// replacing the expired CA and issuing the final, bounded leaf.
	serving.Hostnames = []string{"new.example.test"}
	chains, err := builder.Complete()
	require.NoError(t, err)
	require.NoError(t, chains.Regenerate("root"))
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
