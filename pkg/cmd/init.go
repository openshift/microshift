/*
Copyright © 2021 Microshift Contributors

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
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/openshift/microshift/pkg/util"
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "openshift init",
	Long:  `openshift init`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return initAll(args)
	},
}

func initAll(args []string) error {
	// create CA and keys
	if err := initCerts(); err != nil {
		return err
	}
	// create configs
	if err := initServerConfig(); err != nil {
		return err
	}
	// create kubeconfig for kube-scheduler, kubelet, openshift-apiserver,controller-manager
	/*
		if err := util.Kubeconfig(); err != nil {
			return err
		}
	*/
	return nil
}

func initCerts() error {
	// etcd
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("failed to get hostname: %v", err)
	}

	ip, err := util.GetHostIP()
	if err != nil {
		return fmt.Errorf("failed to get host IP: %v", err)
	}
	// based on https://github.com/openshift/cluster-etcd-operator/blob/master/bindata/bootkube/bootstrap-manifests/etcd-member-pod.yaml#L19
	if err := util.GenCerts("/etc/kubernetes/ushift-certs/secrets/etcd-all-serving",
		"etcd-serving.crt", "etcd-serving.key",
			[]string{"[]string{localhost", ip, "127.0.0.1", hostname}); err != nil {
		return err
	}
	if err := util.StoreRootCA("/etc/kubernetes/ushift-certs/configmaps/etcd-serving-ca",
		"ca-bundle.crt", "ca-bundle.key"); err != nil {
		return err
	}

	if err := util.GenCerts("/etc/kubernetes/ushift-certs/secrets/etcd-all-peer",
		"etcd-peer.crt", "etcd-peer.key",
			[]string{"[]string{localhost", ip, "127.0.0.1", hostname}); err != nil {
		return err
	}
	if err := util.StoreRootCA("/etc/kubernetes/ushift-certs/configmaps/etcd-peer-client-ca",
		"ca-bundle.crt", "ca-bundle.key"); err != nil {
		return err
	}


	// kube-apiserver
	// etcd-cafile: /etc/kubernetes/ushift-resources/configmaps/etcd-serving-ca/ca-bundle.crt
	if err := util.StoreRootCA("/etc/kubernetes/ushift-resources/configmaps/etcd-serving-ca",
		"ca-bundle.crt", "ca-bundle.key"); err != nil {
		return err
	}
	// etcd-certfile: /etc/kubernetes/ushift-resources/secrets/etcd-client/tls.crt
	// etcd-keyfile: /etc/kubernetes/ushift-resources/secrets/etcd-client/tls.key
	if err := util.GenCerts("/etc/kubernetes/ushift-resources/secrets/etcd-client",
		"tls.crt", "tls.key",
			[]string{"[]string{localhost", ip, "127.0.0.1", hostname}); err != nil {
		return err
	}
	// kube-apiserver
	// client-ca-file: /etc/kubernetes/ushift-certs/configmaps/client-ca/ca-bundle.crt
	if err := util.StoreRootCA("/etc/kubernetes/ushift-certs/configmaps/client-ca/",
		"ca-bundle.crt", "ca-bundle.key"); err != nil {
		return err
	}
	// kubelet
	// kubelet-certificate-authority: /etc/kubernetes/static-pod-resources/configmaps/kubelet-serving-ca/ca-bundle.crt
	if err := util.GenCerts("/etc/kubernetes/static-pod-resources/configmaps/kubelet-serving-ca",
		"ca-bundle.crt",
		"ca-bundle.key",
	[]string{"localhost", ip, "127.0.0.1", hostname}); err != nil {
		return err
	}
	// kubelet-client-certificate: /etc/kubernetes/static-pod-resources/secrets/kubelet-client/tls.crt
	if err := util.GenCerts("/etc/kubernetes/static-pod-resources/secrets/kubelet-client",
		"tls.crt",
		"tls.key",
	[]string{"localhost", ip, "127.0.0.1", hostname}); err != nil {
		return err
	}
	// kubelet-client-key: /etc/kubernetes/static-pod-resources/secrets/kubelet-client/tls.key
	if err := util.GenCerts("/etc/kubernetes/static-pod-resources/secrets/kubelet-client",
		"tls.crt",
		"tls.key",
	[]string{"localhost", ip, "127.0.0.1", hostname}); err != nil {
		return err
	}
	// proxy client
	// proxy-client-cert-file: /etc/kubernetes/static-pod-certs/secrets/aggregator-client/tls.crt
	// proxy-client-key-file: /etc/kubernetes/static-pod-certs/secrets/aggregator-client/tls.key
	if err := util.GenCerts("/etc/kubernetes/static-pod-certs/secrets/aggregator-client/",
		"tls.crt",
		"tls.key",
	[]string{"localhost", ip, "127.0.0.1", hostname}); err != nil {
		return err
	}
	// request header
	// requestheader-client-ca-file: /etc/kubernetes/static-pod-certs/configmaps/aggregator-client-ca/ca-bundle.crt
	if err := util.GenCerts("/etc/kubernetes/static-pod-certs/configmaps/aggregator-client-ca/ca-bundle.crt",
		"ca-bundle.crt",
		"ca-bundle.key",
	[]string{"localhost", ip, "127.0.0.1", hostname}); err != nil {
		return err
	}
	// tls
	// tls-cert-file: /etc/kubernetes/static-pod-certs/secrets/service-network-serving-certkey/tls.crt
	// tls-private-key-file: /etc/kubernetes/static-pod-certs/secrets/service-network-serving-certkey/tls.key
	if err := util.GenCerts("/etc/kubernetes/static-pod-certs/secrets/service-network-serving-certkey",
		"tls.crt",
		"tls.key",
	[]string{"localhost", ip, "127.0.0.1", hostname}); err != nil {
		return err
	}
	// kube-controller-manager
	// root-ca-file: /etc/kubernetes/static-pod-resources/configmaps/serviceaccount-ca/ca-bundle.crt
	if err := util.GenCerts("/etc/kubernetes/static-pod-resources/configmaps/serviceaccount-ca/",
		"ca-bundle.crt",
		"ca-bundle.key",
	[]string{"localhost", ip, "127.0.0.1", hostname}); err != nil {
		return err
	}
	// service-account-private-key-file: /etc/kubernetes/static-pod-resources/secrets/service-account-private-key/service-account.key
	if err := util.GenCerts("/etc/kubernetes/static-pod-resources/secrets/service-account-private-key",
		"service-account.crt",
		"service-account.key",
	[]string{"localhost", ip, "127.0.0.1", hostname}); err != nil {
		return err
	}
	// cluster-signing-cert-file: /etc/kubernetes/static-pod-certs/secrets/csr-signer/tls.crt
	// cluster-signing-key-file: /etc/kubernetes/static-pod-certs/secrets/csr-signer/tls.key
	if err := util.GenCerts("/etc/kubernetes/static-pod-certs/secrets/csr-signer",
		"tls.crt",
		"tls.key",
	[]string{"localhost", ip, "127.0.0.1", hostname}); err != nil {
		return err
	}
	// kube-scheduler
	// client-cert-key
	// /etc/kubernetes/static-pod-certs/secrets/kube-scheduler-client-cert-key
	if err := util.GenCerts("/etc/kubernetes/static-pod-certs/secrets/kube-scheduler-client-cert-key",
		"tls.crt",
		"tls.key",
	[]string{"localhost", ip, "127.0.0.1", hostname}); err != nil {
		return err
	}
	// openshift-apiserver
	// /run/secrets/serving-cert
	if err := util.GenCerts("/run/secrets/serving-cert",
		"tls.crt",
		"tls.key",
	[]string{"localhost", ip, "127.0.0.1", hostname}); err != nil {
		return err
	}
	// openshift-controller-manager
	// /run/secrets/serving-cert
	if err := util.GenCerts("/run/secrets/serving-cert",
		"tls.crt",
		"tls.key",
	[]string{"localhost", ip, "127.0.0.1", hostname}); err != nil {
		return err
	}
	return nil
}

func initServerConfig() error {
	if err := util.KubeAPIServerConfig("/etc/kubernetes/ushift-resources/kube-apiserver/config/config.yaml", "" /*svc CIDR*/); err != nil {
		return err
	}
	if err := util.KubeControllerManagerConfig("/etc/kubernetes/ushift-resources/kube-controller-manager/config/config.yaml"); err != nil {
		return err
	}
	if err := util.OpenShiftAPIServerConfig("/etc/kubernetes/ushift-resources/openshift-apiserver/config/config.yaml"); err != nil {
		return err
	}
	if err := util.OpenShiftControllerManagerConfig("/etc/kubernetes/ushift-resources/openshift-controller-manager/config/config.yaml"); err != nil {
		return err
	}
	return nil
}
