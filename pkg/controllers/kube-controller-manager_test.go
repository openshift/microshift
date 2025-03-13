/*
Copyright © 2022 MicroShift Contributors

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

package controllers

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/openshift/library-go/pkg/crypto"

	"github.com/google/go-cmp/cmp"
	embedded "github.com/openshift/microshift/assets"
	"github.com/openshift/microshift/pkg/config"
)

func TestKCMDefaultConfigAsset(t *testing.T) {
	defaultConfigBytes, err := embedded.Asset(kcmDefaultConfigAsset)
	if err != nil {
		t.Errorf("expected valid default config, error: %v", err)
	}
	if len(defaultConfigBytes) == 0 {
		t.Errorf("expected non-empty byte array")
	}
}

func TestConfigure(t *testing.T) {
	cfg := config.NewDefault()
	kcm := NewKubeControllerManager(context.TODO(), cfg)

	clusterSigningKey, clusterSigningCert := kcmClusterSigningCertKeyAndFile()
	argsWant := []string{
		"--allocate-node-cidrs=true",
		fmt.Sprintf("--authentication-kubeconfig=%s", cfg.KubeConfigPath(config.KubeControllerManager)),
		fmt.Sprintf("--authorization-kubeconfig=%s", cfg.KubeConfigPath(config.KubeControllerManager)),
		"--cert-dir=/var/run/kubernetes",
		"--cloud-provider=external",
		"--cluster-cidr=10.42.0.0/16",
		fmt.Sprintf("--cluster-signing-cert-file=%s", clusterSigningCert),
		"--cluster-signing-duration=720h",
		fmt.Sprintf("--cluster-signing-key-file=%s", clusterSigningKey),
		"--controllers=*",
		"--controllers=-bootstrapsigner",
		"--controllers=-tokencleaner",
		"--controllers=-ttl",
		"--controllers=selinux-warning-controller",
		"--enable-dynamic-provisioning=true",
		"--kube-api-burst=300",
		"--kube-api-qps=150",
		fmt.Sprintf("--kubeconfig=%s", cfg.KubeConfigPath(config.KubeControllerManager)),
		"--leader-elect-renew-deadline=12s",
		"--leader-elect-resource-lock=leases",
		"--leader-elect-retry-period=3s",
		"--leader-elect=false",
		fmt.Sprintf("--root-ca-file=%s", kcmRootCAFile()),
		"--secure-port=10257",
		fmt.Sprintf("--service-account-private-key-file=%s", kcmServiceAccountPrivateKeyFile()),
		fmt.Sprintf("--service-cluster-ip-range=%s", cfg.Network.ServiceNetwork[0]),
		fmt.Sprintf("--tls-cipher-suites=%s", strings.Join(crypto.OpenSSLToIANACipherSuites(fixedTLSProfile.Ciphers), ",")),
		fmt.Sprintf("--tls-min-version=%s", string(fixedTLSProfile.MinTLSVersion)),
		"--use-service-account-credentials=true",
		"-v=2",
	}

	argsGot := kcm.args
	if !reflect.DeepEqual(argsWant, argsGot) {
		t.Errorf("expected args to match - diff: %s", cmp.Diff(argsWant, argsGot))
	}
}
