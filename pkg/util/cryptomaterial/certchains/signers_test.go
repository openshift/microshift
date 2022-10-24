package certchains

import (
	"path/filepath"
	"reflect"
	"testing"

	"k8s.io/apiserver/pkg/authentication/user"
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
			testSigner: NewCertificateSigner("test-signer-signer", filepath.Join(tmpDir, "generalTest"), 1).
				WithClientCertificates(
					&ClientCertificateSigningRequestInfo{
						CertificateSigningRequestInfo: CertificateSigningRequestInfo{
							Name:         "test-client",
							ValidityDays: 1,
						},
						UserInfo: &user.DefaultInfo{Name: "test-user", Groups: []string{"test-group1", "test-group2"}},
					},
					&ClientCertificateSigningRequestInfo{
						CertificateSigningRequestInfo: CertificateSigningRequestInfo{
							Name:         "test-client2",
							ValidityDays: 1,
						},
						UserInfo: &user.DefaultInfo{Name: "test-user", Groups: []string{"test-group1", "test-group2"}},
					},
				).WithServingCertificates(
				&ServingCertificateSigningRequestInfo{
					CertificateSigningRequestInfo: CertificateSigningRequestInfo{
						Name:         "test-server",
						ValidityDays: 1,
					},
					Hostnames: []string{"localhost", "127.0.0.1"},
				},
			).
				WithSubCAs(NewCertificateSigner("test-signer", filepath.Join(tmpDir, "test-signer"), 1)),
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
