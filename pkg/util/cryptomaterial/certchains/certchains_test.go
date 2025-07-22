package certchains

import (
	"crypto/x509"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/diff"
	"k8s.io/apiserver/pkg/authentication/user"
)

func testChains(t *testing.T, tmpDir string) *CertificateChains {
	ret, err := NewCertificateChains(
		NewCertificateSigner("test-signer1", filepath.Join(tmpDir, "test-signer1"), 6*365*24*time.Hour).
			WithClientCertificates(&ClientCertificateSigningRequestInfo{
				CSRMeta: CSRMeta{
					Name:     "test-client1",
					Validity: 365 * 24 * time.Hour,
				},
				UserInfo: &user.DefaultInfo{Name: "test-user", Groups: []string{"test-group1", "test-group2"}},
			},
				&ClientCertificateSigningRequestInfo{
					CSRMeta: CSRMeta{
						Name:     "test-client2",
						Validity: 365 * 24 * time.Hour,
					},
					UserInfo: &user.DefaultInfo{Name: "test-user2"},
				},
			).WithServingCertificates(&ServingCertificateSigningRequestInfo{
			CSRMeta: CSRMeta{
				Name:     "test-signer1-server1",
				Validity: 365 * 24 * time.Hour,
			},
			Hostnames: []string{"behind.the.wardrobe.door"},
		}).WithSubCAs(
			NewCertificateSigner("test-signer1-subca", filepath.Join(tmpDir, "test-signer1", "intemediateDir", "subca"), 6*365*24*time.Hour).
				WithServingCertificates(&ServingCertificateSigningRequestInfo{
					CSRMeta: CSRMeta{
						Name:     "test-signer1-subca-server1",
						Validity: 365 * 24 * time.Hour,
					},
					Hostnames: []string{"newname.host"},
				}).WithSubCAs(
				NewCertificateSigner("test-signer1-subca-too", filepath.Join(tmpDir, "test-signer1", "intemediateDir", "subca", "subca-too"), 6*365*24*time.Hour).
					WithClientCertificates(&ClientCertificateSigningRequestInfo{
						CSRMeta: CSRMeta{
							Name:     "subca-too-test-client1",
							Validity: 365 * 24 * time.Hour,
						},
						UserInfo: &user.DefaultInfo{Name: "test-user", Groups: []string{"test-group1", "test-group2"}},
					}).WithSubCAs(
					NewCertificateSigner("test-signer1-subca-too-too", filepath.Join(tmpDir, "test-signer1", "intemediateDir", "subca", "subca-too", "subca-too-too"), 6*365*24*time.Hour).
						WithClientCertificates(&ClientCertificateSigningRequestInfo{
							CSRMeta: CSRMeta{
								Name:     "subca-too-too-test-client2",
								Validity: 270 * 24 * time.Hour,
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
					Name:     "test-signer2-server1",
					Validity: 365 * 24 * time.Hour,
				},
				Hostnames: []string{"somewhere.over.the.rainbow", "bluebirds.fly"},
			}),
		NewCertificateSigner("test-signer3", filepath.Join(tmpDir, "test-signer3"), 4*365*24*time.Hour).
			WithServingCertificates(&ServingCertificateSigningRequestInfo{
				CSRMeta: CSRMeta{
					Name:     "test-signer3-server1",
					Validity: 365 * 24 * time.Hour,
				},
				Hostnames: []string{"castle.brobdingnag"},
			}).
			WithSubCAs(NewCertificateSigner("test-signer3-subca1", filepath.Join(tmpDir, "test-signer3-subca1"), 6*365*24*time.Hour).
				WithClientCertificates(&ClientCertificateSigningRequestInfo{
					CSRMeta: CSRMeta{
						Name:     "test-client1",
						Validity: 365 * 24 * time.Hour,
					},
					UserInfo: &user.DefaultInfo{Name: "test-user", Groups: []string{"test-group1", "test-group2"}},
				}),
			).
			WithPeerCertificiates(&PeerCertificateSigningRequestInfo{
				CSRMeta: CSRMeta{
					Name:     "test-peer1",
					Validity: 365 * 24 * time.Hour,
				},
				UserInfo:  &user.DefaultInfo{Name: "test-user", Groups: []string{"test-group1"}},
				Hostnames: []string{"bring.a.towel"},
			}),
	).Complete()

	require.NoError(t, err)
	return ret
}

func TestCertificateChains_WalkChains(t *testing.T) {
	tmpDir := t.TempDir()

	testChain := testChains(t, tmpDir)

	tests := []struct {
		name             string
		path             []string
		expectedSubjects string
		wantErr          bool
	}{
		{
			name:    "full tree traversal",
			path:    nil,
			wantErr: false,
			expectedSubjects: `
CN=test-signer1
	CN=test-signer1-subca
		CN=test-signer1-subca-too
			CN=test-signer1-subca-too-too
				CN=test-user2
			CN=test-signer1-subca-too-too2
			CN=test-user,O=test-group1+O=test-group2
		CN=newname.host
	CN=test-user,O=test-group1+O=test-group2
	CN=test-user2
	CN=behind.the.wardrobe.door
CN=test-signer2
	CN=bluebirds.fly
CN=test-signer3
	CN=test-signer3-subca1
		CN=test-user,O=test-group1+O=test-group2
	CN=test-user,O=test-group1
	CN=castle.brobdingnag`,
		},
		{
			name:    "1-level signer",
			path:    []string{"test-signer2"},
			wantErr: false,
			expectedSubjects: `
CN=test-signer2
	CN=bluebirds.fly`,
		},
		{
			name:    "signer w/ subca",
			path:    []string{"test-signer3"},
			wantErr: false,
			expectedSubjects: `
CN=test-signer3
	CN=test-signer3-subca1
		CN=test-user,O=test-group1+O=test-group2
	CN=test-user,O=test-group1
	CN=castle.brobdingnag`,
		},
		{
			name:    "signer/subca",
			path:    []string{"test-signer3", "test-signer3-subca1"},
			wantErr: false,
			expectedSubjects: `
	CN=test-signer3-subca1
		CN=test-user,O=test-group1+O=test-group2`,
		},
		{
			name:    "leaf cert",
			path:    []string{"test-signer2", "test-signer2-server1"},
			wantErr: false,
			expectedSubjects: `
	CN=bluebirds.fly`,
		},
		{
			name:    "leaf cert of subca",
			path:    []string{"test-signer3", "test-signer3-subca1", "test-client1"},
			wantErr: false,
			expectedSubjects: `
		CN=test-user,O=test-group1+O=test-group2`,
		},
		{
			name:    "nonexistent signer",
			path:    []string{"test-signer4"},
			wantErr: true,
		},
		{
			name:    "nonexistent intermediate signer",
			path:    []string{"test-signer3", "test-signer3-subca2", "test-client1"},
			wantErr: true,
		},
		{
			name:    "nonexistent leaf",
			path:    []string{"test-signer3", "test-signer3-subca1", "test-client2"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var subjects string
			walkFunc := func(path []string, c x509.Certificate) error {
				t.Helper()
				subjects += "\n" + strings.Repeat("\t", len(path)-1) + c.Subject.String()
				return nil
			}

			if err := testChain.WalkChains(tt.path, walkFunc); (err != nil) != tt.wantErr {
				t.Errorf("CertificateChains.WalkChains() error = %v, wantErr %v", err, tt.wantErr)
			}

			require.Equal(t, tt.expectedSubjects, subjects, "diff %s", diff.StringDiff(subjects, tt.expectedSubjects))
		})
	}
}

func TestWhenToRotateAtEarliest(t *testing.T) {
	tmpDir := t.TempDir()

	testChain := testChains(t, tmpDir)

	certPath, rotationTime, err := WhenToRotateAtEarliest(testChain)
	require.NoError(t, err)
	require.Equal(t, []string{"test-signer1", "test-signer1-subca", "test-signer1-subca-too", "test-signer1-subca-too-too", "subca-too-too-test-client2"}, certPath)

	require.True(t, time.Now().Add(4*30*24*time.Hour).Before(rotationTime) && time.Now().Add(7*30*24*time.Hour).After(rotationTime), "the rotate time is at %s", rotationTime.String())
}
