package certchains

import (
	"crypto/x509"
	"path/filepath"
	"testing"
	"time"

	"github.com/openshift/library-go/pkg/crypto"
	"github.com/stretchr/testify/require"
	"k8s.io/apiserver/pkg/authentication/user"
)

func testChains(t *testing.T, tmpDir string) *CertificateChains {
	ret, err := NewCertificateChains(
		NewCertificateSigner("test-signer1", filepath.Join(tmpDir, "test-signer1"), 6*365*24*time.Hour).
			WithClientCertificates(&ClientCertificateSigningRequestInfo{
				CSRMeta: CSRMeta{
					RotationPolicy: RotationPolicyStandard,
					Name:           "test-client1",
					Validity:       365 * 24 * time.Hour,
				},
				UserInfo: &user.DefaultInfo{Name: "test-user", Groups: []string{"test-group1", "test-group2"}},
			},
				&ClientCertificateSigningRequestInfo{
					CSRMeta: CSRMeta{
						RotationPolicy: RotationPolicyStandard,
						Name:           "test-client2",
						Validity:       365 * 24 * time.Hour,
					},
					UserInfo: &user.DefaultInfo{Name: "test-user2"},
				},
			).WithServingCertificates(&ServingCertificateSigningRequestInfo{
			CSRMeta: CSRMeta{
				RotationPolicy: RotationPolicyStandard,
				Name:           "test-signer1-server1",
				Validity:       365 * 24 * time.Hour,
			},
			Hostnames: []string{"behind.the.wardrobe.door"},
		}).WithSubCAs(
			NewCertificateSigner("test-signer1-subca", filepath.Join(tmpDir, "test-signer1", "intemediateDir", "subca"), 6*365*24*time.Hour).
				WithServingCertificates(&ServingCertificateSigningRequestInfo{
					CSRMeta: CSRMeta{
						RotationPolicy: RotationPolicyStandard,
						Name:           "test-signer1-subca-server1",
						Validity:       365 * 24 * time.Hour,
					},
					Hostnames: []string{"newname.host"},
				}).WithSubCAs(
				NewCertificateSigner("test-signer1-subca-too", filepath.Join(tmpDir, "test-signer1", "intemediateDir", "subca", "subca-too"), 6*365*24*time.Hour).
					WithClientCertificates(&ClientCertificateSigningRequestInfo{
						CSRMeta: CSRMeta{
							RotationPolicy: RotationPolicyStandard,
							Name:           "subca-too-test-client1",
							Validity:       365 * 24 * time.Hour,
						},
						UserInfo: &user.DefaultInfo{Name: "test-user", Groups: []string{"test-group1", "test-group2"}},
					}).WithSubCAs(
					NewCertificateSigner("test-signer1-subca-too-too", filepath.Join(tmpDir, "test-signer1", "intemediateDir", "subca", "subca-too", "subca-too-too"), 6*365*24*time.Hour).
						WithClientCertificates(&ClientCertificateSigningRequestInfo{
							CSRMeta: CSRMeta{
								RotationPolicy: RotationPolicyStandard,
								Name:           "subca-too-too-test-client2",
								Validity:       270 * 24 * time.Hour,
							},
							UserInfo: &user.DefaultInfo{Name: "test-user2"},
						}),
					NewCertificateSigner("test-signer1-subca-too-too2", filepath.Join(tmpDir, "test-signer1", "intemediateDir", "subca", "subca-too", "subca-too-too2"), 3*365*24*time.Hour),
				),
			),
		),
		NewCertificateSigner("test-signer2", filepath.Join(tmpDir, "test-signer2"), 6*365*24*time.Hour).
			WithServingCertificates(&ServingCertificateSigningRequestInfo{
				CSRMeta: CSRMeta{
					RotationPolicy: RotationPolicyStandard,
					Name:           "test-signer2-server1",
					Validity:       365 * 24 * time.Hour,
				},
				Hostnames: []string{"somewhere.over.the.rainbow", "bluebirds.fly"},
			}),
		NewCertificateSigner("test-signer3", filepath.Join(tmpDir, "test-signer3"), 4*365*24*time.Hour).
			WithServingCertificates(&ServingCertificateSigningRequestInfo{
				CSRMeta: CSRMeta{
					RotationPolicy: RotationPolicyStandard,
					Name:           "test-signer3-server1",
					Validity:       365 * 24 * time.Hour,
				},
				Hostnames: []string{"castle.brobdingnag"},
			}).
			WithSubCAs(NewCertificateSigner("test-signer3-subca1", filepath.Join(tmpDir, "test-signer3-subca1"), 6*365*24*time.Hour).
				WithClientCertificates(&ClientCertificateSigningRequestInfo{
					CSRMeta: CSRMeta{
						RotationPolicy: RotationPolicyStandard,
						Name:           "test-client1",
						Validity:       365 * 24 * time.Hour,
					},
					UserInfo: &user.DefaultInfo{Name: "test-user", Groups: []string{"test-group1", "test-group2"}},
				}),
			).
			WithPeerCertificiates(&PeerCertificateSigningRequestInfo{
				CSRMeta: CSRMeta{
					RotationPolicy: RotationPolicyStandard,
					Name:           "test-peer1",
					Validity:       365 * 24 * time.Hour,
				},
				UserInfo:  &user.DefaultInfo{Name: "test-user", Groups: []string{"test-group1"}},
				Hostnames: []string{"bring.a.towel"},
			}),
	).Complete()

	require.NoError(t, err)
	return ret
}

func TestWhenToRotateAtEarliest(t *testing.T) {
	tmpDir := t.TempDir()

	testChain := testChains(t, tmpDir)

	certPath, rotationTime, err := WhenToRotateAtEarliest(testChain)
	require.NoError(t, err)
	require.Equal(t, []string{"test-signer1", "test-signer1-subca", "test-signer1-subca-too", "test-signer1-subca-too-too", "subca-too-too-test-client2"}, certPath)

	require.True(t, time.Now().Add(4*30*24*time.Hour).Before(rotationTime) && time.Now().Add(7*30*24*time.Hour).After(rotationTime), "the rotate time is at %s", rotationTime.String())
}

func TestRotationDeadlineUsesEncodedLifetimeAndPolicy(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	for _, policy := range []RotationPolicy{RotationPolicyStandard, RotationPolicyExtended, RotationPolicyUnknown} {
		for _, lifetime := range []time.Duration{1008 * time.Hour, 8760 * time.Hour, 87600 * time.Hour} {
			cert := &x509.Certificate{NotBefore: now, NotAfter: now.Add(lifetime)}
			chains := &CertificateChains{signers: map[string]*CertificateSigner{
				"root": {
					signerName: "root",
					signerConfig: &crypto.CA{Config: &crypto.TLSCertificateConfig{Certs: []*x509.Certificate{
						{NotBefore: now, NotAfter: now.Add(2 * lifetime)},
					}}},
					signedCertificates: map[string]*signedCertificateInfo{
						"client": {
							CSRInfo:   &ClientCertificateSigningRequestInfo{CSRMeta: CSRMeta{Name: "client", RotationPolicy: policy}},
							tlsConfig: &crypto.TLSCertificateConfig{Certs: []*x509.Certificate{cert}},
						},
					},
				},
			}}
			path, deadline, err := WhenToRotateAtEarliest(chains)
			if policy == RotationPolicyUnknown {
				require.ErrorContains(t, err, "unknown rotation policy")
				continue
			}
			require.NoError(t, err)
			require.Equal(t, []string{"root", "client"}, path)
			_, critical, err := policy.StatusThresholds()
			require.NoError(t, err)
			require.Equal(t, cert.NotAfter.Add(-time.Duration(float64(lifetime)*critical)), deadline)
		}
	}
}
