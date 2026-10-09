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
	"strings"
	"testing"
	"time"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util/cryptomaterial/certchains"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		"admin-kubeconfig-signer":                                                      {"authentication", ca, extended},
		"admin-kubeconfig-signer/admin-kubeconfig-client":                              {"authentication", client, extended},
		"admin-kubeconfig-signer/openshift-observability-client":                       {"observability", client, standard},
		"aggregator-signer":                                                            {"kube-apiserver", ca, extended},
		"aggregator-signer/aggregator-client":                                          {"kube-apiserver", client, standard},
		"etcd-signer":                                                                  {"etcd", ca, extended},
		"etcd-signer/apiserver-etcd-client":                                            {"kube-apiserver", client, extended},
		"etcd-signer/etcd-peer":                                                        {"etcd", peer, extended},
		"etcd-signer/etcd-serving":                                                     {"etcd", peer, extended},
		"ingress-ca":                                                                   {"ingress", ca, extended},
		"ingress-ca/router-default-serving":                                            {"ingress", serving, standard},
		"kube-apiserver-external-signer":                                               {"kube-apiserver", ca, extended},
		"kube-apiserver-external-signer/kube-external-serving":                         {"kube-apiserver", serving, standard},
		"kube-apiserver-localhost-signer":                                              {"kube-apiserver", ca, extended},
		"kube-apiserver-localhost-signer/kube-apiserver-localhost-serving":             {"kube-apiserver", serving, standard},
		"kube-apiserver-service-network-signer":                                        {"kube-apiserver", ca, extended},
		"kube-apiserver-service-network-signer/kube-apiserver-service-network-serving": {"kube-apiserver", serving, standard},
		"kube-apiserver-to-kubelet-signer":                                             {"kube-apiserver", ca, extended},
		"kube-apiserver-to-kubelet-signer/kube-apiserver-to-kubelet-client":            {"kube-apiserver", client, standard},
		"kube-apiserver-to-kubelet-signer/metrics-server-kubelet-client":               {"metrics-server", client, standard},
		"kube-control-plane-signer":                                                    {"control-plane", ca, extended},
		"kube-control-plane-signer/cluster-policy-controller":                          {"cluster-policy-controller", client, standard},
		"kube-control-plane-signer/kube-controller-manager":                            {"kube-controller-manager", client, standard},
		"kube-control-plane-signer/kube-scheduler":                                     {"kube-scheduler", client, standard},
		"kube-control-plane-signer/route-controller-manager":                           {"route-controller-manager", client, standard},
		"kubelet-signer":                                                               {"kubelet", ca, extended},
		"kubelet-signer/kube-csr-signer":                                               {"kubelet", ca, extended},
		"kubelet-signer/kube-csr-signer/kubelet-client":                                {"kubelet", client, standard},
		"kubelet-signer/kube-csr-signer/kubelet-server":                                {"kubelet", serving, standard},
		"service-ca": {"service-ca", ca, extended},
		"service-ca/route-controller-manager-serving": {"route-controller-manager", serving, standard},
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
	now := time.Now().UTC().Truncate(time.Second)
	for _, policy := range []certchains.RotationPolicy{certchains.RotationPolicyStandard, certchains.RotationPolicyExtended} {
		for _, lifetime := range []time.Duration{1008 * time.Hour, 8760 * time.Hour, 87600 * time.Hour} {
			warning, _, err := policy.StatusThresholds()
			require.NoError(t, err)
			boundary := time.Duration(float64(lifetime) * warning)
			for _, remaining := range []time.Duration{lifetime, boundary + time.Second, boundary, boundary - time.Second, 0, -time.Second, lifetime + time.Second} {
				entry := certificateInventoryEntry("test", "certificate", certchains.CertificateRoleClient, policy, now.Add(remaining-lifetime), now.Add(remaining))
				entry.Path = []string{"signer", "certificate"}
				got, err := certsToRegenerate(certchains.CertificateInventory{entry}, now)
				require.NoError(t, err)
				if remaining > boundary && remaining <= lifetime {
					require.Empty(t, got)
				} else {
					require.Equal(t, [][]string{entry.Path}, got)
				}
			}
		}
	}
	got, err := certsToRegenerate(nil, now)
	require.NoError(t, err)
	require.Empty(t, got)
	unknown := certificateInventoryEntry("test", "unknown", certchains.CertificateRoleClient, certchains.RotationPolicyUnknown, now.Add(-time.Hour), now.Add(time.Hour))
	_, err = certsToRegenerate(certchains.CertificateInventory{unknown}, now)
	require.ErrorContains(t, err, "unknown rotation policy")
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
