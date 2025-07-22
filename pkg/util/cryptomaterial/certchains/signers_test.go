package certchains

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"k8s.io/apiserver/pkg/authentication/user"

	"github.com/openshift/microshift/pkg/util/cryptomaterial"
)

func Test_certificateSigner_Complete(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		testSigner CertificateSignerBuilder
		wantCerts  []string
		wantSubCAs []string
		wantErr    bool
	}{
		{
			name: "general test",
			testSigner: NewCertificateSigner("test-signer-signer", filepath.Join(tmpDir, "generalTest"), 1*24*time.Hour).
				WithClientCertificates(
					&ClientCertificateSigningRequestInfo{
						CSRMeta: CSRMeta{
							Name:     "test-client",
							Validity: 1 * 24 * time.Hour,
						},
						UserInfo: &user.DefaultInfo{Name: "test-user", Groups: []string{"test-group1", "test-group2"}},
					},
					&ClientCertificateSigningRequestInfo{
						CSRMeta: CSRMeta{
							Name:     "test-client2",
							Validity: 1 * 24 * time.Hour,
						},
						UserInfo: &user.DefaultInfo{Name: "test-user", Groups: []string{"test-group1", "test-group2"}},
					},
				).WithServingCertificates(
				&ServingCertificateSigningRequestInfo{
					CSRMeta: CSRMeta{
						Name:     "test-server",
						Validity: 1 * 24 * time.Hour,
					},
					Hostnames: []string{"localhost", "127.0.0.1"},
				},
			).
				WithSubCAs(NewCertificateSigner("test-signer", filepath.Join(tmpDir, "test-signer"), 1*24*time.Hour)),
			wantCerts:  []string{"test-client", "test-client2", "test-server"},
			wantSubCAs: []string{"test-signer"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.testSigner.Complete()
			if (err != nil) != tt.wantErr {
				t.Errorf("certificateSigner.Complete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotCertNames := got.GetCertNames(); !reflect.DeepEqual(gotCertNames, tt.wantCerts) {
				t.Errorf("the completed signer cert names = %v, want %v", gotCertNames, tt.wantCerts)
			}
			if gotSubCANames := got.GetSubCANames(); !reflect.DeepEqual(gotSubCANames, tt.wantSubCAs) {
				t.Errorf("the completed signer sub-CA names = %v, want %v", gotSubCANames, tt.wantSubCAs)
			}
		})
	}
}

func TestCertificateSigner_Regenerate(t *testing.T) {
	tmpDir := t.TempDir()

	filesStruct := map[string]string{
		"":                                     cryptomaterial.CACertPath(filepath.Join(tmpDir, "root")),
		"test-client":                          cryptomaterial.ClientCertPath(filepath.Join(tmpDir, "root", "test-client")),
		"test-client2":                         cryptomaterial.ClientCertPath(filepath.Join(tmpDir, "root", "test-client2")),
		"test-server":                          cryptomaterial.ServingCertPath(filepath.Join(tmpDir, "root", "test-server")),
		"test-signer":                          cryptomaterial.CACertPath(filepath.Join(tmpDir, "test-signer")),
		"test-signer/test-signer-sub":          cryptomaterial.CACertPath(filepath.Join(tmpDir, "test-signer", "test-signer-sub")),
		"test-signer/test-signer-sub/sub-peer": cryptomaterial.PeerCertPath(filepath.Join(tmpDir, "test-signer", "test-signer-sub", "sub-peer")),
		"trust-bundle":                         filepath.Join(tmpDir, "trust", "ca-bundle.crt"),
	}

	testSigner := mustCompleteSigner(t,
		NewCertificateSigner("test-signer-signer", filepath.Join(tmpDir, "root"), 1*24*time.Hour).
			WithClientCertificates(
				&ClientCertificateSigningRequestInfo{
					CSRMeta: CSRMeta{
						Name:     "test-client",
						Validity: 1 * 24 * time.Hour,
					},
					UserInfo: &user.DefaultInfo{Name: "test-user", Groups: []string{"test-group1", "test-group2"}},
				},
				&ClientCertificateSigningRequestInfo{
					CSRMeta: CSRMeta{
						Name:     "test-client2",
						Validity: 1 * 24 * time.Hour,
					},
					UserInfo: &user.DefaultInfo{Name: "test-user", Groups: []string{"test-group1", "test-group2"}},
				},
			).WithServingCertificates(
			&ServingCertificateSigningRequestInfo{
				CSRMeta: CSRMeta{
					Name:     "test-server",
					Validity: 1 * 24 * time.Hour,
				},
				Hostnames: []string{"localhost", "127.0.0.1"},
			},
		).WithSubCAs(
			NewCertificateSigner("test-signer", filepath.Join(tmpDir, "test-signer"), 1*24*time.Hour).
				WithSubCAs(
					NewCertificateSigner("test-signer-sub", filepath.Join(tmpDir, "test-signer", "test-signer-sub"), 1*24*time.Hour).
						WithPeerCertificiates(&PeerCertificateSigningRequestInfo{
							CSRMeta: CSRMeta{
								Name:     "sub-peer",
								Validity: 1 * 24 * time.Hour,
							},
							UserInfo:  &user.DefaultInfo{Name: "test-peer"},
							Hostnames: []string{"some.hostname.yay"},
						},
						).WithCABundlePaths(filepath.Join(tmpDir, "trust", "ca-bundle.crt")),
				),
		),
	)
	_ = filepath.Walk(tmpDir, func(name string, info os.FileInfo, err error) error {
		fmt.Println(name)
		return nil
	})

	tests := []struct {
		name            string
		regenCertPath   []string
		changedFileKeys []string
		wantErr         bool
	}{
		{
			name:            "regen 1st level leaf client",
			regenCertPath:   []string{"test-client2"},
			changedFileKeys: []string{"test-client2"},
		},
		{
			name:            "regen 1st level leaf server",
			regenCertPath:   []string{"test-server"},
			changedFileKeys: []string{"test-server"},
		},
		{
			name:            "regen 3rd level leaf peer",
			regenCertPath:   []string{"test-signer", "test-signer-sub", "sub-peer"},
			changedFileKeys: []string{"test-signer/test-signer-sub/sub-peer"},
		},
		{
			name:          "regen a 2nd level sub",
			regenCertPath: []string{"test-signer", "test-signer-sub"},
			changedFileKeys: []string{
				"test-signer/test-signer-sub",
				"test-signer/test-signer-sub/sub-peer",
				"trust-bundle",
			},
		},
		{
			name:          "regen the whole signer",
			regenCertPath: nil,
			changedFileKeys: []string{
				"",
				"test-client",
				"test-client2",
				"test-server",
				"test-signer",
				"test-signer/test-signer-sub",
				"test-signer/test-signer-sub/sub-peer",
				"trust-bundle",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preFileContent := make([][]byte, 0, len(tt.changedFileKeys))
			for _, k := range tt.changedFileKeys {
				fileContent, err := os.ReadFile(filesStruct[k])
				require.NoError(t, err)

				preFileContent = append(preFileContent, fileContent)
			}

			if err := testSigner.Regenerate(tt.regenCertPath...); (err != nil) != tt.wantErr {
				t.Errorf("CertificateSigner.Regenerate() error = %v, wantErr %v", err, tt.wantErr)
			}

			for i, k := range tt.changedFileKeys {
				postFileContent, err := os.ReadFile(filesStruct[k])
				require.NoError(t, err)

				require.NotZero(t, bytes.Compare(preFileContent[i], postFileContent), "the file %s did not change", filesStruct[k])
			}
		})
	}
}

func mustCompleteSigner(t *testing.T, s CertificateSignerBuilder) *CertificateSigner {
	ret, err := s.Complete()
	require.NoError(t, err)
	return ret
}
