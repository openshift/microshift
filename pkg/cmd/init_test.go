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
	"strings"
	"testing"
	"time"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util/cryptomaterial/certchains"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apiserver/pkg/authentication/user"
)

func TestProductionCertificateInventory(t *testing.T) {
	cfg := &config.Config{
		Network:   config.Network{ServiceNetwork: []string{"10.43.0.0/16"}},
		Node:      config.Node{HostnameOverride: "test-node", NodeIP: "192.0.2.10"},
		DNS:       config.DNS{BaseDomain: "example.test"},
		ApiServer: config.ApiServer{AdvertiseAddress: "10.44.0.0"},
	}
	dataDir := t.TempDir()
	builder, err := certificateChainsSetup(cfg, dataDir)
	require.NoError(t, err)
	chains, err := builder.Complete()
	require.NoError(t, err)
	// Use a fresh builder to exercise status's disk-loading path independently
	// of the builder that created the certificates.
	builder, err = certificateChainsSetup(cfg, dataDir)
	require.NoError(t, err)
	inventory, err := builder.LoadInventory()
	require.NoError(t, err)
	require.Equal(t, chains.Inventory(), inventory)

	const (
		ca       = certchains.CertificateRoleCA
		client   = certchains.CertificateRoleClient
		serving  = certchains.CertificateRoleServing
		peer     = certchains.CertificateRolePeer
		standard = certchains.RotationPolicyStandard
		extended = certchains.RotationPolicyExtended
	)
	type metadata struct {
		service string
		role    certchains.CertificateRole
		policy  certchains.RotationPolicy
	}
	expected := map[string]metadata{
		// Consolidated client CA — replaces kube-control-plane-signer, kube-apiserver-to-kubelet-signer,
		// admin-kubeconfig-signer, and kubelet-signer/kube-csr-signer
		"client-ca":                                  {"authentication", ca, extended},
		"client-ca/kube-controller-manager":          {"kube-controller-manager", client, standard},
		"client-ca/kube-scheduler":                   {"kube-scheduler", client, standard},
		"client-ca/cluster-policy-controller":        {"cluster-policy-controller", client, standard},
		"client-ca/route-controller-manager":         {"route-controller-manager", client, standard},
		"client-ca/kube-apiserver-to-kubelet-client": {"kube-apiserver", client, standard},
		"client-ca/metrics-server-kubelet-client":    {"metrics-server", client, standard},
		"client-ca/admin-kubeconfig-client":          {"authentication", client, extended},
		"client-ca/openshift-observability-client":   {"observability", client, standard},
		"client-ca/kubelet-client":                   {"kubelet", client, standard},
		// Consolidated serving CA — replaces kube-apiserver-external/localhost/service-network-signer
		"serving-ca":                        {"kube-apiserver", ca, extended},
		"serving-ca/kube-apiserver-serving": {"kube-apiserver", serving, standard},
		"serving-ca/kubelet-server":         {"kubelet", serving, standard},
		// Peer CA — replaces etcd-signer
		"peer-ca":                       {"etcd", ca, extended},
		"peer-ca/apiserver-etcd-client": {"kube-apiserver", client, extended},
		"peer-ca/etcd-peer":             {"etcd", peer, extended},
		"peer-ca/etcd-serving":          {"etcd", peer, extended},
		// Unchanged CAs
		"aggregator-signer":                           {"kube-apiserver", ca, extended},
		"aggregator-signer/aggregator-client":         {"kube-apiserver", client, standard},
		"service-ca":                                  {"service-ca", ca, extended},
		"service-ca/route-controller-manager-serving": {"route-controller-manager", serving, standard},
		"ingress-ca":                                  {"ingress", ca, extended},
		"ingress-ca/router-default-serving":           {"ingress", serving, standard},
	}
	require.Len(t, inventory, len(expected))
	for _, entry := range inventory {
		path := strings.Join(entry.Path, "/")
		t.Run(path, func(t *testing.T) {
			want, ok := expected[path]
			require.True(t, ok, "unexpected certificate %q", path)
			require.Equal(t, want, metadata{entry.Service, entry.Role, entry.RotationPolicy})
			parts := strings.Split(path, "/")
			require.Equal(t, parts[len(parts)-1], entry.Name)
			if len(parts) == 1 {
				require.Empty(t, entry.ParentCA)
			} else {
				require.Equal(t, parts[len(parts)-2], entry.ParentCA)
			}
		})
		delete(expected, path)
	}
	require.Empty(t, expected, "every production certificate must be covered")
}

func Test_certsToRegenerate(t *testing.T) {
	tests := []struct {
		name   string
		chains *certchains.CertificateChains
		want   [][]string
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
			got := certsToRegenerate(tt.chains)
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
