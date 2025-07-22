/*
Copyright © 2021 MicroShift Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util/cryptomaterial/certchains"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apiserver/pkg/authentication/user"
)

func Test_certsToRegenerate(t *testing.T) {
	tests := []struct {
		name    string
		chains  *certchains.CertificateChains
		want    [][]string
		wantErr bool
	}{
		{
			name:   "empty chains",
			chains: &certchains.CertificateChains{},
			want:   [][]string{},
		},
		{
			name: "no cert to regenerate",
			chains: mustComplete(t,
				certchains.NewCertificateChains(certchains.NewCertificateSigner("signer", t.TempDir(), 365*24*time.Hour).
					WithClientCertificates(&certchains.ClientCertificateSigningRequestInfo{
						CSRMeta: certchains.CSRMeta{
							Name:     "somename",
							Validity: 280 * 24 * time.Hour,
						},
						UserInfo: &user.DefaultInfo{Name: "someclient"},
					}),
				)),
			want: [][]string{},
		},
		{
			name: "signer needs regen",
			chains: mustComplete(t,
				certchains.NewCertificateChains(certchains.NewCertificateSigner("signer", t.TempDir(), 140*24*time.Hour).
					WithClientCertificates(&certchains.ClientCertificateSigningRequestInfo{
						CSRMeta: certchains.CSRMeta{
							Name:     "somename",
							Validity: 270 * 24 * time.Hour,
						},
						UserInfo: &user.DefaultInfo{Name: "someclient"},
					}),
				)),
			want: [][]string{{"signer"}},
		},
		{
			name: "leaf cert needs regen",
			chains: mustComplete(t,
				certchains.NewCertificateChains(certchains.NewCertificateSigner("signer", t.TempDir(), 270*24*time.Hour).
					WithClientCertificates(&certchains.ClientCertificateSigningRequestInfo{
						CSRMeta: certchains.CSRMeta{
							Name:     "somename",
							Validity: 150 * 24 * time.Hour,
						},
						UserInfo: &user.DefaultInfo{Name: "someclient"},
					}),
				),
			),
			want: [][]string{{"signer", "somename"}},
		},
		{
			name: "leaf cert needs regen",
			chains: mustComplete(t,
				certchains.NewCertificateChains(certchains.NewCertificateSigner("signer", t.TempDir(), 270*24*time.Hour).
					WithClientCertificates(&certchains.ClientCertificateSigningRequestInfo{
						CSRMeta: certchains.CSRMeta{
							Name:     "somename",
							Validity: 150 * 24 * time.Hour,
						},
						UserInfo: &user.DefaultInfo{Name: "someclient"},
					}),
				),
			),
			want: [][]string{{"signer", "somename"}},
		},
		{
			name: "both need regen",
			chains: mustComplete(t,
				certchains.NewCertificateChains(certchains.NewCertificateSigner("signer", t.TempDir(), 160*24*time.Hour).
					WithClientCertificates(&certchains.ClientCertificateSigningRequestInfo{
						CSRMeta: certchains.CSRMeta{
							Name:     "somename",
							Validity: 150 * 24 * time.Hour,
						},
						UserInfo: &user.DefaultInfo{Name: "someclient"},
					}),
				),
			),
			want: [][]string{{"signer"}, {"signer", "somename"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := certsToRegenerate(tt.chains)
			if (err != nil) != tt.wantErr {
				t.Errorf("certsToRegenerate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("certsToRegenerate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_removeStaleKubeconfig(t *testing.T) {
	rootDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("unable to create temporary dir: %v", err)
	}
	defer os.RemoveAll(rootDir)

	cfg := &config.Config{
		Node: config.Node{
			HostnameOverride: "hostname",
		},
		ApiServer: config.ApiServer{
			SubjectAltNames: []string{"altname1", "altname2"},
		},
	}
	for _, dir := range append(cfg.ApiServer.SubjectAltNames, cfg.Node.HostnameOverride) {
		assert.NoError(t, os.Mkdir(filepath.Join(rootDir, dir), 0600))
	}

	staleDir, err := os.MkdirTemp(rootDir, "example")
	if err != nil {
		t.Fatalf("unable to create temporary dir: %v", err)
	}
	assert.NoError(t, cleanupStaleKubeconfigs(cfg, rootDir))
	_, err = os.Stat(staleDir)
	if err == nil {
		t.Fatalf("%s should have been deleted", staleDir)
	}
	if !os.IsNotExist(err) {
		t.Fatalf("unable to check %s existence: %v", staleDir, err)
	}
	for _, dir := range append(cfg.ApiServer.SubjectAltNames, cfg.Node.HostnameOverride) {
		d := filepath.Join(rootDir, dir)
		if _, err = os.Stat(d); err != nil {
			t.Fatalf("dir %s should remain: %v", d, err)
		}
	}
}

func mustComplete(t *testing.T, cs certchains.CertificateChainsBuilder) *certchains.CertificateChains {
	ret, err := cs.Complete()
	require.NoError(t, err)
	return ret
}
