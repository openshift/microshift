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
	if err := util.Kubeconfig(); err != nil {
		return err
	}
	return nil
}

func initCerts() error {
	// etcd
	// etcd-cafile: /etc/kubernetes/static-pod-resources/configmaps/etcd-serving-ca/ca-bundle.crt
	if err := util.GenCerts("etcd-cafile", "/etc/kubernetes/static-pod-resources/configmaps/etcd-serving-ca",
		"ca-bundle.crt", "ca-bundle.key"); err != nil {
		return err
	}
	// etcd-certfile: /etc/kubernetes/static-pod-resources/secrets/etcd-client/tls.crt
	// etcd-keyfile: /etc/kubernetes/static-pod-resources/secrets/etcd-client/tls.key
	if err := util.GenCerts("etcd-cert", "/etc/kubernetes/static-pod-resources/secrets/etcd-client",
		"tls.crt", "tls.key"); err != nil {
		return err
	}
	// kube-apiserver
	// client-ca-file: /etc/kubernetes/static-pod-certs/configmaps/client-ca/ca-bundle.crt

	// kubelet
	// kubelet-certificate-authority: /etc/kubernetes/static-pod-resources/configmaps/kubelet-serving-ca/ca-bundle.crt
	// kubelet-client-certificate: /etc/kubernetes/static-pod-resources/secrets/kubelet-client/tls.crt
	// kubelet-client-key: /etc/kubernetes/static-pod-resources/secrets/kubelet-client/tls.key

	// proxy client
	// proxy-client-cert-file: /etc/kubernetes/static-pod-certs/secrets/aggregator-client/tls.crt
	// proxy-client-key-file: /etc/kubernetes/static-pod-certs/secrets/aggregator-client/tls.key

	// request header
	// requestheader-client-ca-file: /etc/kubernetes/static-pod-certs/configmaps/aggregator-client-ca/ca-bundle.crt

	// tls
	// tls-cert-file: /etc/kubernetes/static-pod-certs/secrets/service-network-serving-certkey/tls.crt
	// tls-private-key-file: /etc/kubernetes/static-pod-certs/secrets/service-network-serving-certkey/tls.key

	// kube-controller-manager
	// root-ca-file: /etc/kubernetes/static-pod-resources/configmaps/serviceaccount-ca/ca-bundle.crt
	// service-account-private-key-file: /etc/kubernetes/static-pod-resources/secrets/service-account-private-key/service-account.key
	// cluster-signing-cert-file: /etc/kubernetes/static-pod-certs/secrets/csr-signer/tls.crt
	// cluster-signing-key-file: /etc/kubernetes/static-pod-certs/secrets/csr-signer/tls.key

	// kube-scheduler

	// openshift-apiserver

	// openshift-controller-manager

	return nil
}

func initServerConfig() error {
	if err := util.KubeAPIServerConfig(); err != nil {
		return err
	}
	if err := util.KubeControllerManagerConfig(); err != nil {
		return err
	}
	if err := util.OpenShiftAPIServerConfig(); err != nil {
		return err
	}
	if err := util.OpenShiftControllerManagerConfig(); err != nil {
		return err
	}

	return nil
}
